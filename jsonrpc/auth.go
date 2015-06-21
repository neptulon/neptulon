package jsonrpc

import "log"

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
		// todo: better use sender and send back a notification and close conn immediately as we don't know the type of incoming message at this point
		// other approach would be to generate a response based on incoming message type
		ctx.OutMsg = &Message{Error: &ResError{Code: 666, Message: "Invalid client certificate.", Data: certs}}
		ctx.Conn.Close()
		return

		// ctx.Sender(...)
		// ctx.Conn.Close()
		// return
	}

	userID := certs[0].Subject.CommonName
	ctx.Conn.Session.Set("userid", userID)
	log.Println("Client-certificate authenticated:", ctx.Conn.RemoteAddr(), userID)
}
