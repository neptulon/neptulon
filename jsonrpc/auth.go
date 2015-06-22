package jsonrpc

import "log"

// CertAuth is a TLS certificate authentication middleware for Neptulon JSON-RPC app.
type CertAuth struct {
}

// NewCertAuth creates and registers a new certificate authentication middleware instance with a Neptulon JSON-RPC app.
func NewCertAuth(app *App) (*CertAuth, error) {
	a := CertAuth{}
	app.ReqMiddleware(a.reqMiddleware)
	return &a, nil
}

func (a *CertAuth) reqMiddleware(ctx *ReqContext) {
	if ctx.Conn.Session.Get("userid") != nil {
		return
	}

	// if provided, client certificate is verified by the TLS listener so the peerCerts list in the connection is trusted
	certs := ctx.Conn.ConnectionState().PeerCertificates
	if len(certs) == 0 {
		ctx.ResErr = &ResError{Code: 666, Message: "Invalid client certificate.", Data: certs}
		log.Println("Invalid client-certificate connection attempt:", ctx.Conn.RemoteAddr())
		// todo: close conn
		return
	}

	userID := certs[0].Subject.CommonName
	ctx.Conn.Session.Set("userid", userID)
	log.Println("Client-certificate authenticated:", ctx.Conn.RemoteAddr(), userID)
}

// todo: also check notification and response routes but how to streamline this? revive generic app.Middleware(ctx *Message) ???
