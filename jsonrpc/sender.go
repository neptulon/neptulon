package jsonrpc

import "github.com/nbusy/neptulon"

// Sender is a JSON-RPC request/notification sending middleware.
type Sender struct {
	pendinRequests map[string]bool
}

func (s *Sender) Request(req *Request) {
	s.pendinRequests[req.ID] = true
}

func (s *Sender) middleware(conn *neptulon.Conn, res *Response) {
	if s.pendinRequests[res.ID] {
		// ...
		delete(s.pendinRequests, res.ID)
	}
}
