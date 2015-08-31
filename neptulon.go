// Package neptulon is a socket framework with middleware support.
package neptulon

import (
	"fmt"
	"log"
	"sync"
)

// Server is a Neptulon server.
type Server struct {
	debug      bool
	err        error
	errMutex   sync.RWMutex
	listener   *Listener
	middleware []func(conn *Conn, msg []byte) []byte
	conns      map[string]*Conn
	connMutex  sync.Mutex
}

// NewServer creates a Neptulon server. This is the default TLS constructor.
// Debug mode dumps raw TCP data to stderr (log.Println() default).
func NewServer(cert, privKey, clientCACert []byte, laddr string, debug bool) (*Server, error) {
	l, err := Listen(cert, privKey, clientCACert, laddr, debug)
	if err != nil {
		return nil, err
	}

	return &Server{
		debug:    debug,
		listener: l,
		conns:    make(map[string]*Conn),
	}, nil
}

// Middleware registers a new middleware to handle incoming messages.
func (s *Server) Middleware(middleware func(conn *Conn, msg []byte) []byte) {
	s.middleware = append(s.middleware, middleware)
}

// Run starts accepting connections on the internal listener and handles connections with registered middleware.
// This function blocks and never returns, unless there was an error while accepting a new connection or the listner was closed.
func (s *Server) Run() error {
	err := s.listener.Accept(handleConn(s), handleMsg(s), handleDisconn(s))
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
	return s.conns[connID].Write(msg)
}

// Stop stops a server instance.
func (s *Server) Stop() error {
	err := s.listener.Close()

	// close all active connections discarding any read/writes that is going on currently
	// this is not a problem as we always require an ACK but it will also mean that message deliveries will be at-least-once; to-and-from the server
	s.connMutex.Lock()
	for _, conn := range s.conns {
		conn.Close()
	}
	s.connMutex.Unlock()

	s.errMutex.RLock()
	if s.err != nil {
		return fmt.Errorf("Past internal error: %v", s.err)
	}
	s.errMutex.RUnlock()
	return err
}

func handleConn(s *Server) func(conn *Conn) {
	return func(conn *Conn) {
		s.connMutex.Lock()
		s.conns[conn.ID] = conn
		s.connMutex.Unlock()
	}
}

func handleMsg(s *Server) func(conn *Conn, msg []byte) {
	return func(conn *Conn, msg []byte) {
		for _, m := range s.middleware {
			res := m(conn, msg)
			if res == nil {
				continue
			}

			if err := conn.Write(res); err != nil {
				log.Fatalln("Errored while writing response to connection:", err)
			}

			break
		}
	}
}

func handleDisconn(s *Server) func(conn *Conn) {
	return func(conn *Conn) {
		s.connMutex.Lock()
		delete(s.conns, conn.ID)
		s.connMutex.Unlock()
	}
}
