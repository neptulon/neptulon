package neptulon

import "log"

// Ctx is the incoming message context.
type Ctx struct {
	Conn Conn
	Msg  []byte
	Res  []byte

	m  []func(ctx *Ctx)
	mi int
}

// Next executes the next middleware in the middleware stack.
func (c *Ctx) Next() {
	c.mi++

	if c.mi <= len(c.m) {
		c.m[c.mi-1](c)
	} else if c.Res != nil {
		if err := c.Conn.Write(c.Res); err != nil {
			log.Fatalln("Errored while writing response to connection:", err)
		}
	}
}
