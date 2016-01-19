package jwt

import (
	"fmt"
	"log"

	"github.com/dgrijalva/jwt-go"
	"github.com/neptulon/neptulon"
)

type token struct {
	Token string `json:"message"`
}

// HMAC is JSON Web Token authentication using HMAC.
// If successful, token context will be store with the key "userid" in session.
// If unsuccessful, connection will be closed right away.
func HMAC(password string) func(ctx *neptulon.ReqCtx) error {
	return func(ctx *neptulon.ReqCtx) error {
		if _, ok := ctx.Session.GetOk("userid"); ok {
			return ctx.Next()
		}

		var t token
		if err := ctx.Params(&t); err != nil {
			ctx.Conn.Close()
			return err
		}

		jt, err := jwt.Parse(t.Token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return password, nil
		})

		if err != nil || !jt.Valid {
			log.Println("Invalid JWT authentication attempt:", ctx.Conn.RemoteAddr())
			ctx.Conn.Close()
			return err
		}

		userID := jt.Claims["userid"].(string)
		ctx.Session.Set("userid", userID)
		log.Printf("Client authenticated. TLS/IP: %v, User ID: %v, Conn ID: %v\n", ctx.Conn.RemoteAddr(), userID, ctx.Conn.ID)
		return ctx.Next()
	}
}
