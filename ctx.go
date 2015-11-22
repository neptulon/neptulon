package neptulon

import "log"

// Ctx is the incoming message context.
type Ctx struct {
	Conn Conn
	Msg  []byte
	Res  []byte

	m []func(ctx *Ctx)
	i int
}

// Next executes the next middleware in the middleware stack.
func (c *Ctx) Next() {
	c.i++

	if c.i < len(c.m) {
		c.m[c.i](c)
	} else {
		if err := c.Conn.Write(c.Res); err != nil {
			log.Fatalln("Errored while writing response to connection:", err)
		}
	}
}
