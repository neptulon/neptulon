package jwt

import (
	"log"

	"github.com/neptulon/neptulon"
)

type token struct {
	Token string `json:"message"`
}

// JWT is JSON Web Token authentication.
// If successful, token context will be store with the key "userid" in session.
// If unsuccessful, connection will be closed right away.
func JWT(ctx *neptulon.ReqCtx) error {
	if _, ok := ctx.Session.GetOk("userid"); ok {
		return nil
	}

	// if provided, client certificate is verified by the TLS listener so the peerCerts list in the connection is trusted
	connState, _ := c.Conn.ConnectionState()
	certs := connState.PeerCertificates
	if len(certs) == 0 {
		log.Println("Invalid JWT authentication attempt:", c.Conn.RemoteAddr())
		c.Close()
		return false
	}

	userID := certs[0].Subject.CommonName
	c.Session().Set("userid", userID)
	log.Printf("Client authenticated. TLS/IP: %v, User ID: %v, Conn ID: %v\n", c.Conn.RemoteAddr(), userID, c.ConnID())
	return true

	var t token
	if err := ctx.Params(&t); err != nil {
		return err
	}
	return ctx.Next()
}
