// Package neptulon is a RPC framework with middleware support.
package neptulon

import (
	"log"
	"net/http"
	"net/url"

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
	return &Server{
		addr:  addr,
		conns: cmap.New(),
	}
}

// Middleware registers middleware to handle incoming request messages.
func (s *Server) Middleware(middleware ...func(ctx *ReqCtx) error) {
	s.middleware = append(s.middleware, middleware...)
}

// Start the Neptulon server. This function blocks until server is closed.
func (s *Server) Start() error {
	http.Handle("/", websocket.Server{
		Handler: s.connHandler,
		Handshake: func(config *websocket.Config, req *http.Request) error {
			config.Origin, _ = url.Parse(req.RemoteAddr) // we're interested in remote address and not origin header text
			return nil
		},
	})
	log.Println("Neptulon server started at address:", s.addr)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) connHandler(ws *websocket.Conn) {
	log.Println("Client connected:", ws.RemoteAddr())
	c, err := NewConn(ws, s.middleware)
	if err != nil {
		log.Println("Error while accepting connection:", err)
		return
	}

	s.conns.Set(c.ID, c)
	c.StartReceive()
	s.conns.Delete(c.ID)
	log.Println("Connection closed:", ws.RemoteAddr())
}
