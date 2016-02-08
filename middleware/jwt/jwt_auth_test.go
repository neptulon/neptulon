package jwt

import (
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
)

var (
	pass = "pass"
	now  = time.Now().Unix()
)

type testMsg struct {
	Message string `json:"message"`
	Token   string `json:"token"`
}

func TestTokenGen(t *testing.T) {
	// token := genToken(t)
}

func TestMiddleware(t *testing.T) {

	//
	// mid := HMAC(pass)
}

func genToken(t *testing.T) string {
	token := jwt.New(jwt.SigningMethodHS256)
	token.Claims["userid"] = 1
	token.Claims["created"] = now
	tokenString, err := token.SignedString([]byte(pass))
	if err != nil {
		t.Fatal(err)
	}

	return tokenString
}
