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
	middleware []func(ctx *ReqCtx) error
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
	// receive JSON type T
	var msg message
	websocket.JSON.Receive(ws, &msg)

	// send JSON type T
	websocket.JSON.Send(ws, msg)
}
