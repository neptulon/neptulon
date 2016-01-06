package middleware

import "github.com/neptulon/neptulon"

// CertAtuh is TLS client-certificate authentication.
// If successful, certificate common name will stored with the key "userid" in session.
// If unsuccessful, connection will be closed right away.
func CertAtuh(ctx *neptulon.ReqCtx) {
	// todo: ...
}

// JWT is JSON Web Token authentication.
// If successful, token context will be store with the key "userid" in session.
// If unsuccessful, connection will be closed right away.
func JWT(ctx *neptulon.ReqCtx) {
	// todo: ...
}
