package neptulon

import (
	"fmt"
	"log"
)

// Ctx is the incoming message context.
type Ctx struct {
	Conn    Conn
	Msg     []byte
	Session interface{} // Session is a data store for storing arbitrary data within this context to communicate with middleware further down the stack.

	m  []func(ctx *Ctx)
	mi int
}

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() {
	ctx.mi++

	if ctx.mi <= len(ctx.m) {
		ctx.m[ctx.mi-1](ctx)
	}
}

// Send writes the given message to the connection.
func (ctx *Ctx) Send(msg []byte) error {
	// todo: neptulon.Send vs Ctx.Send

	if err := ctx.Conn.Write(msg); err != nil {
		e := fmt.Errorf("Errored while writing response to connection: %v", err)
		log.Fatalln(e)
		return e
	}

	return nil
}
