// Package neptulon is a RPC framework with middleware support.
package neptulon

import (
	"net/http"

	"github.com/neptulon/cmap"

	"golang.org/x/net/websocket"
)

// Server is a Neptulon server.
type Server struct {
	addr       string
	conns      *cmap.CMap // conn ID -> Conn
	middleware []func(ctx *Ctx) error
}

// NewServer creates a new Neptulon server.
func NewServer(addr string) *Server {
	return &Server{addr: addr}
}

// Start the Neptulon server. This function blocks until server is closed.
func (s *Server) Start() error {
	http.Handle("/", websocket.Handler(s.connHandler))
	return http.ListenAndServe(":12345", nil)
}

func (s *Server) connHandler(ws *websocket.Conn) {
	// todo: we need auth middleware to work before message deserialization so shall we put deserialization into Ctx and read with byte read or put err into Ctx?

	// receive JSON type T
	var m Message
	err := websocket.JSON.Receive(ws, &m)
	if err != nil {
		panic(err)
	}

	// send JSON type T
	err = websocket.JSON.Send(ws, m)
	if err != nil {
		panic(err)
	}
}
