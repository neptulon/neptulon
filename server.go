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
	c, err := NewConn(ws, s.reqMiddleware, s.resMiddleware)
	if err != nil {
		log.Println("Error while accepting connection:", err)
		return
	}

	s.conns.Set(c.ID, c)

	for {
		var m message
		err := c.receive(&m)
		if err != nil {
			log.Println("Error while receiving message:", err)
			break
		}

		// if the message is a request
		if m.Method != "" {
			if err := newReqCtx(c, m.ID, m.Method, m.Params, s.reqMiddleware).Next(); err != nil {
				log.Println("Error while handling request:", err)
				break
			}
		}

		// if the message is a response
		if err := newResCtx(c, m.ID, m.Result, s.resMiddleware).Next(); err != nil {
			log.Println("Error while handling response:", err)
			break
		}
	}

	s.conns.Delete(c.ID)
}
