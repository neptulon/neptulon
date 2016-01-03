// Package neptulon is a RPC framework with middleware support.
package neptulon

import (
	"log"
	"net/http"

	"github.com/neptulon/cmap"

	"golang.org/x/net/websocket"
)

// Server is a Neptulon server.
type Server struct {
	addr          string
	conns         *cmap.CMap // conn ID -> Conn
	reqMiddleware []func(ctx *ReqCtx) error
	resMiddleware []func(ctx *ResCtx) error
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
	c := NewConn(ws)
	s.conns.Set(c.ID, c)

	for {
		var m Message
		err := c.Receive(&m)
		if err != nil {
			log.Println("Error while receiving message:", err)
		}

		err = c.Send(m)
		if err != nil {
			log.Println("Error while sending message:", err)
		}
	}
}
