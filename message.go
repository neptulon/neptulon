package neptulon

import "encoding/json"

// JSON-RPC 2.0 message types. Version field is ommited for brevity.

// Request is a JSON-RPC request object.
type Request struct {
	ID     string      `json:"id"`
	Method string      `json:"method"`
	Params interface{} `json:"params,omitempty"`
}

// Response is a JSON-RPC response object.
type Response struct {
	ID     string      `json:"id"`
	Result interface{} `json:"result,omitempty"`
	Error  *ResError   `json:"error,omitempty"`
}

// ResError is a JSON-RPC response error object.
type ResError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// message is a JSON-RPC request, response, or notification message.
// This is used internally only to manage incoming messages.
// We don't need this for outgoing messages as we always know their specific types.
type message struct {
	ID     string          `json:"id,omitempty"`
	Method string          `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"` // request params
	Result json.RawMessage `json:"result,omitempty"` // response result
	Error  *ResError       `json:"error,omitempty"`  // response error
}
