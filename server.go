// Package neptulon is a socket framework with middleware support.
package neptulon

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/neptulon/client"
	"github.com/neptulon/cmap"
)

// Server is a Neptulon server.
type Server struct {
	debug          bool
	err            error
	errMutex       sync.RWMutex
	listener       *listener
	middlewareIn   []func(ctx *client.Ctx)
	middlewareOut  []func(ctx *client.Ctx)
	clients        *cmap.CMap // conn ID -> Client
	connHandler    func(conn *client.Client)
	disconnHandler func(conn *client.Client)
	net            string // "tls", "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	reqWG          sync.WaitGroup
}

// NewTLSServer creates a Neptulon server using Transport Layer Security.
// Debug mode dumps raw TCP data to stderr (log.Println() default).
func NewTLSServer(cert, privKey, clientCACert []byte, laddr string, debug bool) (*Server, error) {
	l, err := listenTLS(cert, privKey, clientCACert, laddr, debug)
	if err != nil {
		return nil, err
	}

	return &Server{
		debug:    debug,
		listener: l,
		clients:  cmap.New(),
		net:      "tls",
	}, nil
}

// Conn registers a function to handle client connection events.
func (s *Server) Conn(handler func(conn *client.Client)) {
	s.connHandler = handler
}

// MiddlewareIn registers middleware to handle incoming messages.
func (s *Server) MiddlewareIn(middleware ...func(ctx *client.Ctx)) {
	s.middlewareIn = append(s.middlewareIn, middleware...)
}

// MiddlewareOut registers middleware to handle/intercept outgoing messages before they are sent.
func (s *Server) MiddlewareOut(middleware ...func(ctx *client.Ctx)) {
	s.middlewareOut = append(s.middlewareOut, middleware...)
}

// Disconn registers a function to handle client disconnection events.
func (s *Server) Disconn(handler func(c *client.Client)) {
	s.disconnHandler = handler
}

// Run starts accepting connections on the internal listener and handles connections with registered middleware.
// This function blocks and never returns, unless there was an error while accepting a new connection or the listner was closed.
func (s *Server) Run() error {
	err := s.listener.Accept(s.handleConn)
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
	if c, ok := s.clients.GetOk(connID); ok {
		return c.(*client.Client).Send(msg)
	}

	return fmt.Errorf("Connection ID not found: %v", connID)
}

// Stop stops a server instance.
func (s *Server) Stop() error {
	err := s.listener.Close()

	// close all active connections discarding any read/writes that is going on currently
	// this is not a problem as we always require an ACK but it will also mean that message deliveries will be at-least-once; to-and-from the server
	s.clients.Range(func(c interface{}) {
		c.(*client.Client).Disconnect()
	})

	s.errMutex.RLock()
	if s.err != nil {
		return fmt.Errorf("There was a recorded internal error before closing the connection: %v", s.err)
	}

	s.errMutex.RUnlock()
	return err
}

func (s *Server) handleConn(c net.Conn) error {
	switch s.net {
	case "tls":
		tlsc, ok := c.(*tls.Conn)
		if !ok {
			c.Close()
			return errors.New("cannot cast net.Conn interface to tls.Conn type")
		}

		ntlsc, err := client.NewTLSConn(tlsc, 0, 0, 0, s.debug)
		if err != nil {
			return err
		}

		client, err := client.NewClient(ntlsc, &s.reqWG, s.handleDisconn, s.middlewareIn, s.middlewareOut)
		if err != nil {
			return err
		}

		s.clients.Set(ntlsc.ID, client)

		if s.connHandler != nil {
			s.connHandler(client)
		}
	}

	return errors.New("connection is of unknown type")
}

func (s *Server) handleDisconn(c *client.Client) {
	s.clients.Delete(c.Conn.ID)
	if s.disconnHandler != nil {
		s.disconnHandler(c)
	}
}
