package jsonrpc

import (
	"encoding/json"
	"log"

	"github.com/nbusy/neptulon"
)

// Sender is a JSON-RPC request/notification sending middleware.
type Sender struct {
	pendinRequests map[string]chan *Response
}

// NewSender creates a JSON-RPC sender instance and registers it with the Neptulon JSON-RPC app.
func NewSender(app *App) (*Sender, error) {
	s := Sender{
		pendinRequests: make(map[string]chan *Response),
	}

	app.Middleware(s.middleware)
	return &s, nil
}

// Request sends a JSON-RPC request throught the connection denoted by the session ID.
func (s *Sender) Request(sessionID string, req *Request) chan<- *Response {
	data, err := json.Marshal(req)
	if err != nil {
		log.Fatalln("Cannot serialize outgoing request:", err)
	}

	_, err = conn.Write(data) // todo: delegate serialization to jsonrpc app which should delegate writing to neptulon app
	if err != nil {
		log.Fatalln("Errored while writing response to connection:", err)
	}

	ch := make(chan *Response)
	s.pendinRequests[req.ID] = ch
	return ch
}

// Notification sends a JSON-RPC notification throught the connection denoted by the session ID.
func (s *Sender) Notification(sessionID string, not *Notification) {

}

func (s *Sender) middleware(conn *neptulon.Conn, msg *Message) (result interface{}, resErr *ResError) {
	if ch, ok := s.pendinRequests[msg.ID]; ok {
		ch <- &Response{ID: msg.ID, Result: msg.Result, Error: msg.Error}
		delete(s.pendinRequests, msg.ID)
	}

	return nil, nil
}
