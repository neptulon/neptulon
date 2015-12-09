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

	testing  *testing.T
	server   *neptulon.Server
	serverWG sync.WaitGroup // server instance goroutine wait group
}

// NewTLSServerHelper creates a new TLS server helper object.
// Neptulon server instance is initialized and ready to accept connection after this function returns.
func NewTLSServerHelper(t *testing.T) *ServerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	// generate TLS certs
	certChain, err := ca.GenCertChain("FooBar", "127.0.0.1", "127.0.0.1", time.Hour, 512)
	if err != nil {
		t.Fatal("Failed to create TLS certificate chain:", err)
	}

	laddr := "127.0.0.1:3001"
	s, err := neptulon.NewTLSServer(certChain.ServerCert, certChain.ServerKey, certChain.IntCACert, laddr, false)
	if err != nil {
		t.Fatal("Failed to create server:", err)
	}

	h := ServerHelper{
		RootCACert: certChain.RootCACert,
		RootCAKey:  certChain.RootCAKey,
		IntCACert:  certChain.IntCACert,
		IntCAKey:   certChain.IntCAKey,
		ServerCert: certChain.ServerCert,
		ServerKey:  certChain.ServerKey,

		testing: t,
		server:  s,
	}

	h.serverWG.Add(1)
	go func() {
		defer h.serverWG.Done()
		s.Run()
	}()

	time.Sleep(time.Millisecond) // give Run() enough time to initiate

	return &h
}
