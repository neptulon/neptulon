package neptulon

import (
	"encoding/json"
	"fmt"

	"github.com/neptulon/cmap"
)

// Ctx is the message context.
type Ctx struct {
	Res interface{} // Response to be returned.
	Err *ResError   // Error to be returned.

	m       *Message
	mw      []func(ctx *Ctx) error
	mwIndex int
	session *cmap.CMap
}

func newCtx(m *Message, mw []func(ctx *Ctx) error, session *cmap.CMap) *Ctx {
	// append the last middleware to stack, which will write the response to connection, if any
	mw = append(mw, func(ctx *Ctx) error {
		if ctx.Res != nil || ctx.Err != nil {
			// return ctx.Client.SendResponse(ctx.id, ctx.Res, ctx.Err)
		}

		return nil
	})

	return &Ctx{m: m, mw: mw, session: session}
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.
func (ctx *Ctx) Session() *cmap.CMap {
	return ctx.session
}

// Params reads request parameters into given object.
// Object should be passed by reference.
func (ctx *Ctx) Params(v interface{}) error {
	if ctx.m.Params != nil {
		if err := json.Unmarshal(ctx.m.Params, v); err != nil {
			return fmt.Errorf("cannot deserialize request params: %v", err)
		}
	}

	return nil
}

// Next executes the next middleware in the middleware stack.
func (ctx *Ctx) Next() error {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		return ctx.mw[ctx.mwIndex-1](ctx)
	}

	return nil
}
