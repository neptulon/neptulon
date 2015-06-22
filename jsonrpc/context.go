package jsonrpc

import "github.com/nbusy/neptulon"

// ReqContext encapsulates connection, request, and reponse objects for a request.
type ReqContext struct {
	Conn    *neptulon.Conn
	Req     *Request
	Res     interface{}
	ResErr  *ResError
	handled bool
	err     error // returns error to user (if not empty) and closes conn
}

// // Res returns the response object if it was set.
// func (r *ReqContext) Res() interface{} {
// 	return nil
// }
//
// // SetRes sets the response object and marks the request handled.
// func (r *ReqContext) SetRes(res interface{}) {
// 	r.handled = true
// }
//
// // Handled returns true if a response was set or if the request was explicitly marked handled.
// func (r *ReqContext) Handled() bool {
// 	return r.handled
// }
//
// // SetHandled marks the request as handled. This is automatically done when SetRes is used.
// func (r *ReqContext) SetHandled() bool {
// 	return r.handled
// }

// NotContext encapsulates connection and notification objects.
type NotContext struct {
	Conn    *neptulon.Conn
	Not     *Notification
	handled bool
	err     error // returns error to user (if not empty) and closes conn
}

// ResContext encapsulates connection and response objects.
type ResContext struct {
	Conn    *neptulon.Conn
	Res     *Response
	handled bool
	err     error // returns error to user (if not empty) and closes conn
}
