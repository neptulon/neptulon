package jsonrpc

import "github.com/nbusy/neptulon"

// ReqContext encapsulates connection, request, and reponse objects for a request.
type ReqContext struct {
	Conn   *neptulon.Conn
	Req    *Request
	Res    interface{}
	ResErr *ResError
	Done   bool // if set, this will prevent further middleware from handling the request
}

// NotContext encapsulates connection and notification objects.
type NotContext struct {
	Conn *neptulon.Conn
	Not  *Notification
	Done bool // if set, this will prevent further middleware from handling the request
}

// ResContext encapsulates connection and response objects.
type ResContext struct {
	Conn *neptulon.Conn
	Res  *Response
	Done bool // if set, this will prevent further middleware from handling the request
}
