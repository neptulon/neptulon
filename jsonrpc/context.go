package jsonrpc

import "github.com/nbusy/neptulon"

// ReqContext encapsulates connection, request, and reponse objects for a JSON-RPC request.
type ReqContext struct {
	conn    *neptulon.Conn
	req     *Request
	res     interface{}
	resErr  *ResError
	handled bool
}

// Res returns the response object if it was set.
func (r *ReqContext) Res() interface{} {
	return nil
}

// SetRes sets the response object and marks the request handled.
func (r *ReqContext) SetRes(res interface{}) {
	r.handled = true
}

// Handled returns true if a response was set or if the request was explicitly marked handled.
func (r *ReqContext) Handled() bool {
	return r.handled
}

// SetHandled marks the request as handled. This is automatically done when SetRes is used.
func (r *ReqContext) SetHandled() bool {
	return r.handled
}

// NotContext encapsulates connection and notification objects for a JSON-RPC notification.
type NotContext struct {
	conn    *neptulon.Conn
	not     *Notification
	handled bool
}
