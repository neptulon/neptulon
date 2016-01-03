package neptulon

import (
	"encoding/json"
	"fmt"

	"github.com/neptulon/cmap"
)

/*
 * Context object definitions for Request, Response, and Notification middleware.
 */

// ReqCtx encapsulates connection, request, and reponse objects.
type ReqCtx struct {
	Conn *Conn       // Client connection.
	Res  interface{} // Response to be returned.
	Err  *ResError   // Error to be returned.

	id     string          // message ID
	method string          // called method
	params json.RawMessage // request parameters

	mw      []func(ctx *ReqCtx) error
	mwIndex int
	session *cmap.CMap
}

func newReqCtx(id, method string, params json.RawMessage, conn *Conn, mw []func(ctx *ReqCtx) error, session *cmap.CMap) *ReqCtx {
	// append the last middleware to stack, which will write the response to connection, if any
	mw = append(mw, func(ctx *ReqCtx) error {
		if ctx.Res != nil || ctx.Err != nil {
			return ctx.Conn.Send(&Response{ID: ctx.id, Result: ctx.Res, Error: ctx.Err})
		}

		return nil
	})

	return &ReqCtx{Conn: conn, id: id, method: method, params: params, mw: mw}
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.
func (ctx *ReqCtx) Session() *cmap.CMap {
	return ctx.session
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

// ResCtx encapsulates connection and response objects.
type ResCtx struct {
	Conn *Conn // Client connection.

	id     string          // message ID
	result json.RawMessage // result parameters

	err *ResError // response error (if any)

	mw      []func(ctx *ResCtx) error
	mwIndex int
	session *cmap.CMap
}

func newResCtx(id string, result json.RawMessage, conn *Conn, mw []func(ctx *ResCtx) error, session *cmap.CMap) *ResCtx {
	return &ResCtx{Conn: conn, id: id, result: result, mw: mw}
}

// Session is a data store for storing arbitrary data within this context to communicate with other middleware handling this message.
func (ctx *ResCtx) Session() *cmap.CMap {
	return ctx.session
}

// Result reads response result data into given object.
// Object should be passed by reference.
func (ctx *ResCtx) Result(v interface{}) error {
	if ctx.result != nil {
		if err := json.Unmarshal(ctx.result, v); err != nil {
			return fmt.Errorf("cannot deserialize response result: %v", err)
		}
	}

	return nil
}

// Next executes the next middleware in the middleware stack.
func (ctx *ResCtx) Next() error {
	ctx.mwIndex++

	if ctx.mwIndex <= len(ctx.mw) {
		return ctx.mw[ctx.mwIndex-1](ctx)
	}

	return nil
}
