package test

import (
	"sync"
	"testing"
	"time"

	"github.com/neptulon/ca"
	"github.com/neptulon/neptulon"
)

// ServerHelper is a neptulon.Server wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ServerHelper struct {
	// PEM encoded X.509 certificate and private key pairs, if TLS server is used
	RootCACert,
	RootCAKey,
	IntCACert,
	IntCAKey,
	ServerCert,
	ServerKey []byte
	Address string

	testing  *testing.T
	server   *neptulon.Server
	serverWG sync.WaitGroup // server instance goroutine wait group
}

// NewTLSServerHelper creates a new server helper object with Transport Layer Security.
func NewTLSServerHelper(t *testing.T) *ServerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	host, port := "127.0.0.1", "3001"
	laddr := host + ":" + port

	// generate TLS certs
	certChain, err := ca.GenCertChain("FooBar", host, host, time.Hour, 512)
	if err != nil {
		t.Fatal("Failed to create TLS certificate chain:", err)
	}

	s, err := neptulon.NewTLSServer(certChain.ServerCert, certChain.ServerKey, certChain.IntCACert, laddr, false)
	if err != nil {
		t.Fatal("Failed to create server:", err)
	}

	return &ServerHelper{
		RootCACert: certChain.RootCACert,
		RootCAKey:  certChain.RootCAKey,
		IntCACert:  certChain.IntCACert,
		IntCAKey:   certChain.IntCAKey,
		ServerCert: certChain.ServerCert,
		ServerKey:  certChain.ServerKey,
		Address:    laddr,

		testing: t,
		server:  s,
	}
}

// Run initializes the Neptulon server instance which is ready to accept connections after this function returns.
func (s *ServerHelper) Run() {
	s.serverWG.Add(1)
	go func() {
		defer s.serverWG.Done()
		s.server.Run()
	}()

	time.Sleep(time.Millisecond) // give Run() enough CPU cycles to initiate
}

// GetClient creates a connection to this server instance and returns it wrapped in a ClientHelper.
func (s *ServerHelper) GetClient() *ClientHelper {
	return nil
}

// Stop stops the server instance.
func (s *ServerHelper) Stop() {
	if err := s.server.Stop(); err != nil {
		s.testing.Fatal("Failed to stop the server:", err)
	}
	s.serverWG.Wait()
}
