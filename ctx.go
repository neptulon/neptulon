package neptulon

import (
	"encoding/json"
	"fmt"

	"github.com/neptulon/cmap"
)

// ReqCtx encapsulates connection, request, and reponse objects.
type ReqCtx struct {
	Conn    *Conn       // Client connection.
	Session *cmap.CMap  // Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.
	Res     interface{} // Response to be returned.
	Err     *ResError   // Error to be returned.

	id     string          // message ID
	method string          // called method
	params json.RawMessage // request parameters

	mw      []func(ctx *ReqCtx) error
	mwIndex int
}

func newReqCtx(conn *Conn, id, method string, params json.RawMessage, mw []func(ctx *ReqCtx) error) *ReqCtx {
	return &ReqCtx{Conn: conn, Session: cmap.New(), id: id, method: method, params: params, mw: mw}
}

// Params reads request parameters into given object.
// Object should be passed by reference.
func (ctx *ReqCtx) Params(v interface{}) error {
	if ctx.params != nil {
		if err := json.Unmarshal(ctx.params, v); err != nil {
			return fmt.Errorf("cannot deserialize request params: %v", err)
		}
	}

	return nil
}

// Next executes the next middleware in the middleware stack.
func (ctx *ReqCtx) Next() error {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		return ctx.mw[ctx.mwIndex-1](ctx)
	}

	return nil
}
