package client

import "github.com/neptulon/cmap"

// Ctx is the incoming message context.
type Ctx struct {
	Msg  []byte // Incoming message.
	Res  []byte // Response message.
	Conn *Conn  // Client connection.

	mw      []func(ctx *Ctx) error
	mwIndex int
	session *cmap.CMap
}

func newCtx(msg []byte, conn *Conn, mw []func(ctx *Ctx) error) *Ctx {
	// append the last middleware to stack, which will write the response to connection, if any
	mw = append(mw, func(ctx *Ctx) error {
		if ctx.Res != nil {
			return ctx.Conn.Write(ctx.Res)
		}

		return nil
	})

	return &Ctx{Msg: msg, Conn: conn, mw: mw, session: cmap.New()}
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
