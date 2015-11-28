package neptulon

// TLSClient is a Neptulon connection client using Transport Layer Security.
type TLSClient struct {
	Conn Conn

	// middleware for incoming and outgoing messages
	in  []func(ctx *Ctx)
	out []func(ctx *Ctx)
}

// newTLSClient creates a new client using a given Conn.
func newTLSClient(conn Conn) *TLSClient {
	return &TLSClient{
		Conn: conn,
	}
}

// Send writes the given message to the connection.
func (c *TLSClient) Send(msg []byte) error {
	return c.Conn.Write(msg)
}

// SendAsync writes a message to the connection on a saparate gorotuine.
func (c *TLSClient) SendAsync(msg []byte, callback func(error)) {
	go func() {
		if err := c.Conn.Write(msg); err != nil {
			callback(err)
		}
	}()
}

// SendAsync or client_tls_async to send messages in a separate goroutine not to block?
// if we go client_tls_async, we can have
// * client_tls_async / listener / sender + client / server ? or just peer?

// add variadic functions to insert in/out msg middleware + interface definitions.
// move listener handleClient functionality here
// should writing be queue based on a separate thread or configurable?
