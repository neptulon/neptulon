package neptulon

import (
	"fmt"
	"log"
)

// Ctx is the incoming message context.
type Ctx struct {
	Conn Conn
	Msg  []byte

	m  []func(ctx *Ctx)
	mi int
}

// Next executes the next middleware in the middleware stack.
func (c *Ctx) Next() {
	c.mi++

	if c.mi <= len(c.m) {
		c.m[c.mi-1](c)
	}
}

// Send writes the given message to the connection.
func (c *Ctx) Send(msg []byte) error {
	// todo: neptulon.Send vs Ctx.Send

	if err := c.Conn.Write(msg); err != nil {
		e := fmt.Errorf("Errored while writing response to connection: %v", err)
		log.Fatalln(e)
		return e
	}

	return nil
}
