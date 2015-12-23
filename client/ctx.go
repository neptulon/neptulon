package client

import "github.com/neptulon/cmap"

// Ctx is the incoming message context.
type Ctx struct {
	Msg     []byte     // Message body.
	Session *cmap.CMap // Session is a data store for storing arbitrary data within this context to communicate with middleware inthe stack.
	Client  *Client    // Connected client.

	m  []func(ctx *Ctx)
	mi int
}

func newCtx(msg []byte, c *Client, m []func(ctx *Ctx)) *Ctx {
	return &Ctx{Msg: msg, Session: cmap.New(), Client: c, m: m}
}

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() {
	ctx.mi++

	if ctx.mi <= len(ctx.m) {
		ctx.m[ctx.mi-1](ctx)
	}
}

// todo:
// * Ctx.SessionVar("var_name"), Ctx.SetSessionVar("")
// * Note in docs, use Client.Session if you want data to persist for entire connection and not just this message context
// * Ctx.Send/Ctx.SendAsync(or Queue)
