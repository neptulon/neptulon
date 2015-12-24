package client

import "github.com/neptulon/cmap"

// Ctx is the incoming message context.
type Ctx struct {
	Msg  []byte // Message body
	Conn *Conn  // Client connection

	mw      []func(ctx *Ctx)
	mwIndex int
	session *cmap.CMap
}

func newCtx(msg []byte, conn *Conn, mw []func(ctx *Ctx)) *Ctx {
	return &Ctx{Msg: msg, Conn: conn, mw: mw, session: cmap.New()}
}

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		ctx.mw[ctx.mwIndex-1](ctx)
	}
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.

// todo:
// * Ctx.SessionVar("var_name"), Ctx.SetSessionVar("")
// * Note in docs, use Client.Session if you want data to persist for entire connection and not just this message context
// * Ctx.Send/Ctx.SendAsync(or Queue)
// * Remove client and just expose Conn or nothing? If nothing, we need ConnSession

// Ctx: msg/session-get-set/conn/res/m/mi
// Ctx.Conn: session/connID
