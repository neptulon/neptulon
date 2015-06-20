package jsonrpc

import (
	"log"
	"strconv"
)

// CertAuth is a TLS certificate authentication middleware for Neptulon JSON-RPC app.
type CertAuth struct {
}

// NewCertAuth creates and registers a new certificate authentication middleware instance with a Neptulon JSON-RPC app.
func NewCertAuth(app *App) (*CertAuth, error) {
	a := CertAuth{}
	app.Middleware(a.middleware)
	return &a, nil
}

func (a *CertAuth) middleware(ctx *Context) {
	if ctx.Conn.Session.Get("userid") != nil {
		return
	}

	// if provided, client certificate is verified by the TLS listener so the peerCerts list in the connection is trusted
	certs := ctx.Conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		// we should write the response immediately: ctx.Sender(...)
		ctx.ResErr = &ResError{Code: 666, Message: "Invalid client certificate.", Data: certs}
		ctx.Conn.Close()
		return
	}

	idstr := certs[0].Subject.CommonName
	uid64, err := strconv.ParseUint(idstr, 10, 32)
	if err != nil {
		// we should write the response immediately: ctx.Sender(...)
		ctx.ResErr = &ResError{Code: 666, Message: "Invalid client certificate.", Data: certs}
		ctx.Conn.Close()
		return
	}

	userID := uint32(uid64)
	ctx.Conn.Session.Set("userid", userID)
	log.Println("Client-certificate authenticated:", ctx.Conn.RemoteAddr(), userID)
}
