package main

import (
	"crypto/rsa"
	"errors"
	"fmt"

	jwt "github.com/dgrijalva/jwt-go"
	wrapper "github.com/pkg/errors"
)

const (
	signedMethod = "RS256"
	tokenType    = "search_access"
)

// TokenInfo using to store info inside token
type tokenInfo struct {
	Source string
}

// TokenClaims using to store std and custom info inside token
type tokenClaims struct {
	*jwt.StandardClaims
	TokenType string
	Info      tokenInfo
}

// MakeToken create new encrypted token with claims
func makeToken(signKey *rsa.PrivateKey, source string, expireAt int64) (string, error) {
	t := jwt.New(jwt.GetSigningMethod(signedMethod))
	fmt.Println(expireAt)
	// set our claims
	t.Claims = &tokenClaims{
		&jwt.StandardClaims{
			ExpiresAt: expireAt,
		},
		tokenType,
		tokenInfo{Source: source},
	}

	generatedToken, err := t.SignedString(signKey)
	if err != nil {
		return "", wrapper.Wrap(err, "error while get signed string")
	}

	return generatedToken, nil
}

// VerifyToken should parse token info and return claims if token not invalid
func verifyToken(token string, verifyKey *rsa.PublicKey) (tokenClaims, error) {
	if len(token) == 0 {
		return tokenClaims{}, errors.New("token is empty")
	}

	// validate the token
	parsedToken, err := jwt.ParseWithClaims(token, &tokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return verifyKey, nil
	})

	// branch out into the possible error from signing
	switch err.(type) {
	case nil: // no error
		if !parsedToken.Valid { // but may still be invalid
			return tokenClaims{}, errors.New("invalid token")
		}

		claims, ok := parsedToken.Claims.(*tokenClaims)
		if !ok {
			return tokenClaims{}, errors.New("can't parse claims")
		}

		if claims == nil {
			return tokenClaims{}, errors.New("token claims is nil")
		}

		return *claims, nil

	case *jwt.ValidationError: // something was wrong during the validation
		vErr := err.(*jwt.ValidationError)

		switch vErr.Errors {
		case jwt.ValidationErrorExpired:
			return tokenClaims{}, fmt.Errorf("token expired: %v", vErr)

		default:
			return tokenClaims{}, fmt.Errorf("token invalid: %v", vErr)
		}

	default: // something else went wrong
		return tokenClaims{}, err
	}
}

func parsePrivateKey(key []byte) (*rsa.PrivateKey, error) {
	return jwt.ParseRSAPrivateKeyFromPEM(key)
}

func parsePublicKey(key []byte) (*rsa.PublicKey, error) {
	return jwt.ParseRSAPublicKeyFromPEM(key)
}
