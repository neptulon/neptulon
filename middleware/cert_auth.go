package middleware

import "github.com/neptulon/neptulon"

// CertAtuh is TLS client-certificate authentication.
// If successful, certificate common name will stored with the key "userid" in session.
// If unsuccessful, connection will be closed right away.
func CertAtuh(ctx *neptulon.ReqCtx) {
	// todo: ...
}
