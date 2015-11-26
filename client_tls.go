package neptulon

import (
	"fmt"
	"log"
)

// TLSClient is a Neptulon connection client using Transport Layer Security.
type TLSClient struct {
	Conn Conn

	// middleware for incoming and outgoing messages
	in  []func(ctx *Ctx)
	out []func(ctx *Ctx)
}

// NewTLSClient creates a new client using a given tls.Conn.
func NewTLSClient(conn Conn) *TLSClient {
	return &TLSClient{
		Conn: conn,
	}
}

// Send writes the given message to the connection.
func (c *TLSClient) Send(msg []byte) error {
	if err := c.Conn.Write(msg); err != nil {
		e := fmt.Errorf("Errored while writing response to connection: %v", err)
		log.Fatalln(e)
		return e
	}

	return nil
}

// todo: add variadic functions to insert in/out msg middleware + interface definitions.
// move listener handleClient functionality here
// should writing be queue based on a separate thread or configurable?
