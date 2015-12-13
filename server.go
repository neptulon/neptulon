// Package neptulon is a socket framework with middleware support.
package neptulon

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/neptulon/client"
	"github.com/neptulon/cmap"
)

// Server is a Neptulon server.
type Server struct {
	debug          bool
	net            string // "tls", "tcp", "tcp4", "tcp6", "unix" or "unixpacket"
	listener       *listener
	clients        *cmap.CMap // conn ID -> Client
	connWG         sync.WaitGroup
	msgWG          sync.WaitGroup
	middlewareIn   []func(ctx *client.Ctx)
	middlewareOut  []func(ctx *client.Ctx)
	connHandler    func(client *client.Client)
	disconnHandler func(client *client.Client)
}

// NewTLSServer creates a Neptulon server using Transport Layer Security.
func NewTLSServer(cert, privKey, clientCACert []byte, laddr string, debug bool) (*Server, error) {
	l, err := listenTLS(cert, privKey, clientCACert, laddr)
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
func (s *Server) Conn(handler func(client *client.Client)) {
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

// Start starts accepting connections on the internal listener and handles connections with registered middleware.
// This function blocks and never returns until the server is closed by another goroutine or an internal error occurs.
func (s *Server) Start() error {
	if err := s.listener.Accept(s.handleConn); err != nil {
		return fmt.Errorf("And error occured during or after accepting a new connection: %v", err)
	}

	return nil
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

	if err != nil {
		return fmt.Errorf("And error occured before or while stopping the server: %v", err)
	}

	return nil
}

func (s *Server) handleConn(c net.Conn) error {
	switch s.net {
	case "tls":
		tlsc, ok := c.(*tls.Conn)
		if !ok {
			c.Close()
			return errors.New("cannot cast net.Conn interface to tls.Conn type")
		}

		nepTLSConn, err := client.NewTLSConn(tlsc, s.debug)
		if err != nil {
			return err
		}

		client := client.NewClient(&s.msgWG, s.handleDisconn).MiddlewareIn(s.middlewareIn...).MiddlewareOut(s.middlewareOut...).UseConn(nepTLSConn)
		s.clients.Set(nepTLSConn.ID, client)
		s.connWG.Add(1)

		if s.connHandler != nil {
			s.connHandler(client)
		}
	}

	return errors.New("connection is of unknown type")
}

func (s *Server) handleDisconn(c *client.Client) {
	s.clients.Delete(c.Conn.ID)
	s.connWG.Done()
	if s.disconnHandler != nil {
		s.disconnHandler(c)
	}
}
