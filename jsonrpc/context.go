package jsonrpc

import (
	"encoding/json"
	"log"

	"github.com/nbusy/neptulon"
)

// ReqCtx encapsulates connection, request, and reponse objects.
type ReqCtx struct {
	Conn *neptulon.Conn
	Res  interface{} // Response to be returned
	Err  *ResError   // Error to be returned
	Done bool        // If set, this will prevent further middleware from handling the request

	id     string          // message ID
	method string          // called method
	params json.RawMessage // request parameters
}

// Params reads request parameters into given object.
// Object should be passed by reference.
func (r *ReqCtx) Params(v interface{}) {
	if err := json.Unmarshal(r.params, v); err != nil {
		log.Fatal("Cannot deserialize request params:", err)
	}
}

// NotCtx encapsulates connection and notification objects.
type NotCtx struct {
	Conn *neptulon.Conn
	Done bool // If set, this will prevent further middleware from handling the request

	method string          // called method
	params json.RawMessage // notification parameters
}

// Params reads response parameters into given object.
// Object should be passed by reference.
func (r *NotCtx) Params(v interface{}) {
	if err := json.Unmarshal(r.params, v); err != nil {
		log.Fatal("Cannot deserialize notification params:", err)
	}
}

// ResCtx encapsulates connection and response objects.
type ResCtx struct {
	Conn *neptulon.Conn
	Done bool // if set, this will prevent further middleware from handling the request

	id     string          // message ID
	result json.RawMessage // result parameters

	code    int             // error code
	message string          // error message
	data    json.RawMessage // error data
}

// Result reads response result data into given object.
// Object should be passed by reference.
func (r *ResCtx) Result(v interface{}) {
	if err := json.Unmarshal(r.result, v); err != nil {
		log.Fatalln("Cannot deserialize response result:", err)
	}
}
