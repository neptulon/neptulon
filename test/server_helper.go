package test

import (
	"crypto/x509/pkix"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/ca"
	"github.com/neptulon/client/test"
	"github.com/neptulon/neptulon"
)

// ServerHelper is a neptulon.Server wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ServerHelper struct {
	Server *neptulon.Server

	// PEM encoded X.509 certificate and private key pairs, if TLS server is used
	RootCACert,
	RootCAKey,
	IntCACert,
	IntCAKey,
	ServerCert,
	ServerKey []byte
	Address string

	testing  *testing.T
	serverWG sync.WaitGroup // server instance goroutine wait group
	startErr error
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

	server, err := neptulon.NewTLSServer(certChain.ServerCert, certChain.ServerKey, certChain.IntCACert, laddr, false)
	if err != nil {
		t.Fatal("Failed to create server:", err)
	}

	sh := &ServerHelper{
		Server:     server,
		RootCACert: certChain.RootCACert,
		RootCAKey:  certChain.RootCAKey,
		IntCACert:  certChain.IntCACert,
		IntCAKey:   certChain.IntCAKey,
		ServerCert: certChain.ServerCert,
		ServerKey:  certChain.ServerKey,
		Address:    laddr,

		testing: t,
	}

	// start the server immediately
	sh.serverWG.Add(1)
	go func() {
		defer sh.serverWG.Done()
		sh.startErr = sh.Server.Start()
	}()

	time.Sleep(time.Millisecond) // give Accept() enough CPU cycles to initiate
	return sh
}

// GetTLSClient creates a client connection to this server instance using TLS and returns the connection wrapped in a ClientHelper.
func (s *ServerHelper) GetTLSClient(useClientCert bool) *test.ClientHelper {
	var cert, key []byte
	var err error
	if useClientCert {
		cert, key, err = ca.GenClientCert(pkix.Name{
			Organization: []string{"FooBar"},
			CommonName:   "1",
		}, time.Hour, 512, s.IntCACert, s.IntCAKey)
		if err != nil {
			s.testing.Fatal(err)
		}
	}

	return test.NewClientHelper(s.testing).ConnectTLS(s.Address, s.IntCACert, cert, key)
}

// Close stops the server listener and connections.
func (s *ServerHelper) Close() {
	if err := s.Server.Close(); err != nil {
		s.testing.Fatal("Failed to stop the server:", err)
	}

	if s.startErr != nil {
		s.testing.Fatal("Failed to accept connection(s):", s.startErr)
	}

	s.serverWG.Wait()
}
