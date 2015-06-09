package jsonrpc

import "github.com/nbusy/neptulon"

// Sender is a JSON-RPC request/notification sending middleware.
type Sender struct {
	routes map[string]func(conn *neptulon.Conn, msg *Message)
}

func (s *Sender) middleware(conn *neptulon.Conn, msg *Message) {
	s.routes[msg.Method](conn, msg)
}
