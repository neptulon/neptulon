package client

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

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() error {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		return ctx.mw[ctx.mwIndex-1](ctx)
	}

	return nil
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.

// todo:
// * Ctx.SessionVar("var_name"), Ctx.SetSessionVar("")
// * Note in docs, use Client.Session if you want data to persist for entire connection and not just this message context
// * Ctx.Send/Ctx.SendAsync(or Queue)
// * Remove client and just expose Conn or nothing? If nothing, we need ConnSession

// Ctx: msg/session-get-set/conn/res/m/mi
// Ctx.Conn: session/connID

// MiddlewareOut is very weird though, with Msg = nil, Res = msg! We might need to separate InCtx/OutCtx or just use Ctx and remove request/resp paradigm?
