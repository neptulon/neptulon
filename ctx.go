package neptulon

import "github.com/neptulon/cmap"

// Ctx is the incoming message context.
type Ctx struct {
	Msg    []byte  // Message body.
	Client *Client // Connected client.

	mw      []func(ctx *Ctx) error
	mwIndex int
	session *cmap.CMap
}

func newCtx(msg []byte, client *Client, mw []func(ctx *Ctx) error) *Ctx {
	return &Ctx{Msg: msg, Client: client, mw: mw, session: cmap.New()}
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.
func (ctx *Ctx) Session() *cmap.CMap {
	return ctx.session
}

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() error {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		return ctx.mw[ctx.mwIndex-1](ctx)
	}

	return nil
}
