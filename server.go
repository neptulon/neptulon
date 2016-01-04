// Package neptulon is a RPC framework with middleware support.
package neptulon

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/neptulon/cmap"

	"golang.org/x/net/websocket"
)

// Server is a Neptulon server.
type Server struct {
	addr       string
	conns      *cmap.CMap // conn ID -> *Conn
	middleware []func(ctx *ReqCtx) error
	listener   net.Listener
	wsConfig   websocket.Config
	wg         sync.WaitGroup
}

// NewServer creates a new Neptulon server.
func NewServer(addr string) *Server {
	return &Server{
		addr:  addr,
		conns: cmap.New(),
	}
}

// UseTLS enables Transport Layer Security for the connection.
// cert, key = Server certificate/private key pair.
// clientCACert = Optional certificate for verifying client certificates.
// All certificates/private keys are in PEM encoded X.509 format.
func (s *Server) UseTLS(cert, privKey, clientCACert []byte) error {
	tlsCert, err := tls.X509KeyPair(cert, privKey)
	if err != nil {
		return fmt.Errorf("failed to parse the server certificate or the private key: %v", err)
	}

	c, _ := pem.Decode(cert)
	if tlsCert.Leaf, err = x509.ParseCertificate(c.Bytes); err != nil {
		return fmt.Errorf("failed to parse the server certificate: %v", err)
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(clientCACert)
	if !ok {
		return fmt.Errorf("failed to parse the CA certificate: %v", err)
	}

	s.wsConfig.TlsConfig = &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    pool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}

	return nil
}

// Middleware registers middleware to handle incoming request messages.
func (s *Server) Middleware(middleware ...func(ctx *ReqCtx) error) {
	s.middleware = append(s.middleware, middleware...)
}

// Start the Neptulon server. This function blocks until server is closed.
func (s *Server) Start() error {
	http.Handle("/", websocket.Server{
		Config:  s.wsConfig,
		Handler: s.wsHandler,
		Handshake: func(config *websocket.Config, req *http.Request) error {
			s.wg.Add(1)
			config.Origin, _ = url.Parse(req.RemoteAddr) // we're interested in remote address and not origin header text
			return nil
		},
	})

	l, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to create TLS listener on network address %v with error: %v", s.addr, err)
	}
	s.listener = l

	log.Println("Neptulon server started at address:", s.addr)
	return http.Serve(l, nil)
}

func (s *Server) wsHandler(ws *websocket.Conn) {
	defer s.wg.Done()
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

// SendRequest sends a JSON-RPC request through the connection denoted by the connection ID with an auto generated request ID.
// resHandler is called when a response is returned.
func (s *Server) SendRequest(connID string, method string, params interface{}, resHandler func(ctx *ResCtx) error) (reqID string, err error) {
	if conn, ok := s.conns.GetOk(connID); ok {
		return conn.(*Conn).SendRequest(method, params, resHandler)
	}

	return "", fmt.Errorf("connection with requested ID: %v does not exist", connID)
}

// SendRequestArr sends a JSON-RPC request through the connection denoted by the connection ID, with array params and auto generated request ID.
// resHandler is called when a response is returned.
func (s *Server) SendRequestArr(connID string, method string, resHandler func(ctx *ResCtx) error, params ...interface{}) (reqID string, err error) {
	return s.SendRequest(connID, method, params, resHandler)
}

// Close closes the network listener and the active connections.
func (s *Server) Close() error {
	err := s.listener.Close()

	// close all active connections discarding any read/writes that is going on currently
	s.conns.Range(func(c interface{}) {
		c.(*Conn).Close()
	})

	if err != nil {
		return fmt.Errorf("And error occured before or while stopping the server: %v", err)
	}

	s.wg.Wait()
	return nil
}
