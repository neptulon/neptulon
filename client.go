package neptulon

// Client is a Neptulon connection client using Transport Layer Security.
type Client struct {
	Conn *Conn // todo: Conn *TLSConn

	// middleware for incoming and outgoing messages
	in  []func(ctx *Ctx)
	out []func(ctx *Ctx)
}

// todo: remove this!
func newTLSClient(c *Conn, in []func(ctx *Ctx)) *Client {
	return &Client{
		Conn: c,
		in:   in,
	}
}

// Send writes the given message to the connection.
func (c *Client) Send(msg []byte) error {
	ctx := Ctx{m: c.out, Client: c, Msg: msg}
	ctx.Next()
	return c.Conn.Write(ctx.Msg)
}

// SendAsync writes a message to the connection on a saparate gorotuine.
func (c *Client) SendAsync(msg []byte, callback func(error)) {
	go func() {
		if err := c.Send(msg); err != nil {
			// todo: better use an error handler middleware -or- both approaches?
			// todo2: use a single gorotuine + queue otherwise messages get interleaved
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
