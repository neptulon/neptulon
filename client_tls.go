package neptulon

// TLSClient is a Neptulon connection client using Transport Layer Security.
type TLSClient struct {
	// middleware for incoming and outgoing messages
	in  []func(ctx *Ctx)
	out []func(ctx *Ctx)
}

// NewTLSClient creates a new client using a given tls.Conn.
func NewTLSClient() *TLSClient {
	return &TLSClient{}
}

// todo: add variadic functions to insert in/out msg middleware + interface definitions.
