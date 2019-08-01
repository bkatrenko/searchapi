package main

import (
	"crypto/rsa"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	// TestSignKeysPath contains path to the test RSA key (that using for development) and
	// should be used for sign tokens
	TestSignKeysPath = "certs/private.rsa"
	// TestVerifyKeysPath contains test RSA key that should be used for validation tokens
	TestVerifyKeysPath = "certs/public.rsa"
)

func TestMakeToken(t *testing.T) {
	Convey("Test make token", t, func() {
		token, err := makeToken(getSignKey(t), "000000000", 3600)
		if err != nil {
			t.Fatal(err)
		}

		if len(token) == 0 {
			t.Fatal("token length is nil")
		}
	})
}

func TestTokenValidation(t *testing.T) {
	Convey("Test token validation", t, func() {
		stringToken, err := makeToken(getSignKey(t), "000000000", time.Now().Add(time.Minute).Unix())
		if err != nil {
			t.Fatal(err)
		}

		if len(stringToken) == 0 {
			t.Fatal("token length is nil")
		}

		_, err = verifyToken(stringToken, getVerifyKey(t))
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestTokenExpiration(t *testing.T) {
	Convey("Test token expiration", t, func() {
		stringToken, err := makeToken(getSignKey(t), "000000000", time.Now().Add(time.Second).Unix())
		if err != nil {
			t.Fatal(err)
		}

		<-time.After(time.Second * 2)
		if len(stringToken) == 0 {
			t.Fatal("token length is nil")
		}

		_, err = verifyToken(stringToken, getVerifyKey(t))
		if err == nil {
			t.Fatal("expect that token should be expired")
		}

		if !strings.HasPrefix(err.Error(), "token expired") {
			t.Fatal(err, "error should start from token expired message")
		}
	})
}

func getSignKey(t *testing.T) *rsa.PrivateKey {
	keyBytes, err := ioutil.ReadFile(TestSignKeysPath)
	if err != nil {
		t.Fatal(err)
	}

	signKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyBytes)
	if err != nil {
		t.Fatal(err)
	}

	return signKey
}

func getVerifyKey(t *testing.T) *rsa.PublicKey {
	keyBytes, err := ioutil.ReadFile(TestVerifyKeysPath)
	if err != nil {
		t.Fatal(err)
	}

	verifyKey, err := jwt.ParseRSAPublicKeyFromPEM(keyBytes)
	if err != nil {
		t.Fatal(err)
	}

	return verifyKey
}
