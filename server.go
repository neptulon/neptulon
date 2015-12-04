// Package neptulon is a socket framework with middleware support.
package neptulon

import (
	"fmt"
	"log"
	"sync"

	"github.com/neptulon/client"
	"github.com/neptulon/cmap"
)

// Server is a Neptulon server.
type Server struct {
	debug          bool
	err            error
	errMutex       sync.RWMutex
	listener       *Listener
	middleware     []func(ctx *client.Ctx)
	conns          *cmap.CMap // conn ID -> Conn
	connHandler    func(conn *client.Conn)
	disconnHandler func(conn *client.Conn)
}

// NewTLSServer creates a Neptulon server using Transport Layer Security.
// Debug mode dumps raw TCP data to stderr (log.Println() default).
func NewTLSServer(cert, privKey, clientCACert []byte, laddr string, debug bool) (*Server, error) {
	l, err := ListenTLS(cert, privKey, clientCACert, laddr, debug)
	if err != nil {
		return nil, err
	}

	return &Server{
		debug:          debug,
		listener:       l,
		conns:          cmap.New(),
		connHandler:    func(conn *client.Conn) {},
		disconnHandler: func(conn *client.Conn) {},
	}, nil
}

// Conn registers a function to handle client connection events.
func (s *Server) Conn(handler func(conn *client.Conn)) {
	s.connHandler = handler
}

// Middleware registers middleware to handle incoming messages.
func (s *Server) Middleware(middleware ...func(ctx *client.Ctx)) {
	s.middleware = append(s.middleware, middleware...)
}

// Disconn registers a function to handle client disconnection events.
func (s *Server) Disconn(handler func(conn *client.Conn)) {
	s.disconnHandler = handler
}

// Run starts accepting connections on the internal listener and handles connections with registered middleware.
// This function blocks and never returns, unless there was an error while accepting a new connection or the listner was closed.
func (s *Server) Run() error {
	err := s.listener.Accept(s.handleConn, s.handleMsg, s.handleDisconn)
	if err != nil && s.debug {
		log.Fatalln("Listener returned an error while closing:", err)
	}

	s.errMutex.Lock()
	s.err = err
	s.errMutex.Unlock()

	return err
}

// Send sends a message throught the connection denoted by the connection ID.
func (s *Server) Send(connID string, msg []byte) error {
	if conn, ok := s.conns.GetOk(connID); ok {
		return conn.(*client.Conn).Write(msg)
	}

	return fmt.Errorf("Connection ID not found: %v", connID)
}

// Stop stops a server instance.
func (s *Server) Stop() error {
	err := s.listener.Close()

	// close all active connections discarding any read/writes that is going on currently
	// this is not a problem as we always require an ACK but it will also mean that message deliveries will be at-least-once; to-and-from the server
	s.conns.Range(func(conn interface{}) {
		conn.(*client.Conn).Close()
	})

	s.errMutex.RLock()
	if s.err != nil {
		return fmt.Errorf("There was a recorded internal error before closing the connection: %v", s.err)
	}
	s.errMutex.RUnlock()
	return err
}

func (s *Server) handleConn(c *client.Client) {
	s.conns.Set(c.Conn.ID, c.Conn)
	c.MiddlewareIn(s.middleware...)
	s.connHandler(c.Conn)
}

func (s *Server) handleMsg(c *client.Client, msg []byte) {
	ctx, _ := client.NewCtx(c.Conn, msg, s.middleware)
	ctx.Next()
}

func (s *Server) handleDisconn(c *client.Client) {
	s.conns.Delete(c.Conn.ID)
	s.disconnHandler(c.Conn)
}
