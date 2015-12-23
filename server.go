// Package neptulon is a socket framework with middleware support.
package neptulon

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/neptulon/cmap"
	"github.com/neptulon/neptulon/client"
)

// Server is a Neptulon server.
type Server struct {
	debug          bool
	tls            bool
	listener       *listener
	clients        *cmap.CMap // conn ID -> Client
	connWG         sync.WaitGroup
	msgWG          sync.WaitGroup
	middlewareIn   []func(ctx *client.Ctx)
	middlewareOut  []func(ctx *client.Ctx)
	connHandler    func(client *client.Client)
	disconnHandler func(client *client.Client)
}

// NewTCPServer creates a Neptulon TCP server.
func NewTCPServer(laddr string, debug bool) (*Server, error) {
	l, err := listenTCP(laddr)
	if err != nil {
		return nil, err
	}

	return &Server{
		debug:    debug,
		listener: l,
		clients:  cmap.New(),
	}, nil
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
		tls:      true,
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

// Send writes a message to the connection denoted by the connection ID.
func (s *Server) Send(connID string, msg []byte) error {
	if c, ok := s.clients.GetOk(connID); ok {
		return c.(*client.Client).Send(msg)
	}

	return fmt.Errorf("Connection ID not found: %v", connID)
}

// Close closes the network listener and the active connections.
func (s *Server) Close() error {
	err := s.listener.Close()

	// close all active connections discarding any read/writes that is going on currently
	// this is not a problem as we always require an ACK but it will also mean that message deliveries will be at-least-once; to-and-from the server
	s.clients.Range(func(c interface{}) {
		c.(*client.Client).Close()
	})

	if err != nil {
		return fmt.Errorf("And error occured before or while stopping the server: %v", err)
	}

	return nil
}

func (s *Server) handleConn(conn net.Conn) error {
	var c *client.Client
	if s.tls {
		tlsc, ok := conn.(*tls.Conn)
		if !ok {
			conn.Close()
			return errors.New("cannot cast net.Conn interface to tls.Conn type")
		}

		c = client.NewClient(&s.msgWG, s.handleDisconn)
		c.MiddlewareIn(s.middlewareIn...)
		c.MiddlewareOut(s.middlewareOut...)
		if err := c.UseTLSConn(tlsc, s.debug); err != nil {
			return err
		}
	} else {
		tcpc, ok := conn.(*net.TCPConn)
		if !ok {
			conn.Close()
			return errors.New("cannot cast net.Conn interface to net.TCPConn type")
		}

		c = client.NewClient(&s.msgWG, s.handleDisconn)
		c.MiddlewareIn(s.middlewareIn...)
		c.MiddlewareOut(s.middlewareOut...)
		if err := c.UseTCPConn(tcpc, s.debug); err != nil {
			return err
		}
	}

	s.clients.Set(c.ConnID(), c)
	s.connWG.Add(1)

	if s.connHandler != nil {
		s.connHandler(c)
	}

	return nil
}

func (s *Server) handleDisconn(c *client.Client) {
	s.clients.Delete(c.ConnID())
	s.connWG.Done()
	if s.disconnHandler != nil {
		s.disconnHandler(c)
	}
}
