package jsonrpc

import "github.com/nbusy/neptulon"

// Sender is a JSON-RPC request/notification sending middleware.
type Sender struct {
	jsonrpc        *App
	pendinRequests map[string]chan *Response
}

// NewSender creates a JSON-RPC sender instance and registers it with the Neptulon JSON-RPC app.
func NewSender(app *App) (*Sender, error) {
	s := Sender{
		jsonrpc:        app,
		pendinRequests: make(map[string]chan *Response),
	}

	app.Middleware(s.middleware)
	return &s, nil
}

// Request sends a JSON-RPC request throught the connection denoted by the connection ID.
func (s *Sender) Request(connID string, req *Request) chan<- *Response {
	s.jsonrpc.Send(connID, req)
	ch := make(chan *Response)
	s.pendinRequests[req.ID] = ch
	return ch
}

// Notification sends a JSON-RPC notification throught the connection denoted by the connection ID.
func (s *Sender) Notification(connID string, not *Notification) {
	s.jsonrpc.Send(connID, not)
}

func (s *Sender) middleware(conn *neptulon.Conn, msg *Message) (result interface{}, resErr *ResError) {
	if ch, ok := s.pendinRequests[msg.ID]; ok {
		ch <- &Response{ID: msg.ID, Result: msg.Result, Error: msg.Error}
		delete(s.pendinRequests, msg.ID)
	}

	return nil, nil
}
