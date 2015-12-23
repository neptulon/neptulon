package client

// Ctx is the incoming message context.
type Ctx struct {
	Client  *Client
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
