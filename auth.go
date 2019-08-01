package main

import (
	"crypto/rsa"
	"errors"

	jwt "github.com/dgrijalva/jwt-go"
	wrapper "github.com/pkg/errors"
)

const (
	// RS256 is the algorithm for encoding/decoding JWT tokens.
	// You can read more there -> https://auth0.com/blog/navigating-rs256-and-jwks/
	signedMethod = "RS256"
	// tokenType represents functionality of it's token - 'access' means that token can be used
	// for grant access to some resource, refresh can be used for refreshing access token.
	// Ususally no problems to have your own custom token types is you need to override this functionality
	tokenType = "access"
)

// TokenInfo using to store info inside token. Developers usually put there info which can be used
// for authorization/authentication
// https://jwt.io/introduction
type tokenInfo struct {
	// Source contains info about from where token was taken
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

	if verifyKey == nil {
		return tokenClaims{}, errors.New("verify key is nil")
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
			return tokenClaims{}, wrapper.Wrap(err, "token expired")

		default:
			return tokenClaims{}, wrapper.Wrap(err, "token invalid")
		}

	default: // something else went wrong
		return tokenClaims{}, wrapper.Wrap(err, "unexpected token error")
	}
}
