package test

import (
	"os"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/ca"
	"github.com/neptulon/neptulon"
)

const (
	host, port = "127.0.0.1", "3001"
	laddr      = host + ":" + port
)

// ServerHelper is a Neptulon Server wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ServerHelper struct {
	Server *neptulon.Server

	// PEM encoded X.509 certificate and private key pairs, if TLS server is used
	RootCACert,
	RootCAKey,
	IntCACert, // Intermediate signing cert for server and client certificates
	IntCAKey,
	ServerCert,
	ServerKey []byte
	Address string

	testing    *testing.T
	listenerWG sync.WaitGroup // server listener instance goroutine wait group
}

// NewServerHelper creates a new server helper object.
func NewServerHelper(t *testing.T) *ServerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode.")
	}

	return &ServerHelper{
		Server:  neptulon.NewServer(laddr),
		Address: laddr,
		testing: t,
	}
}

// UseTLS enables Transport Layer Security for the connections.
func (sh *ServerHelper) UseTLS() *ServerHelper {
	// generate TLS certs
	certChain, err := ca.GenCertChain("FooBar", host, host, time.Hour, 512)
	if err != nil {
		sh.testing.Fatal("Failed to create TLS certificate chain:", err)
	}

	sh.RootCACert = certChain.RootCACert
	sh.RootCAKey = certChain.RootCAKey
	sh.IntCACert = certChain.IntCACert
	sh.IntCAKey = certChain.IntCAKey
	sh.ServerCert = certChain.ServerCert
	sh.ServerKey = certChain.ServerKey

	sh.Server.UseTLS(sh.ServerCert, sh.ServerKey, sh.IntCACert)

	return sh
}

// ListenAndServe starts the server.
func (sh *ServerHelper) ListenAndServe() *ServerHelper {
	// start the server immediately
	sh.listenerWG.Add(1)
	go func() {
		defer sh.listenerWG.Done()
		if err := sh.Server.ListenAndServe(); err != nil {
			sh.testing.Fatal("Failed to accept connection(s):", err)
		}
	}()

	// give server enough time to initiate
	if os.Getenv("TRAVIS") != "" || os.Getenv("CI") != "" {
		time.Sleep(time.Millisecond * 25)
	} else {
		time.Sleep(time.Millisecond)
	}

	return sh
}

// GetConnHelper creates a client connection to this server instance and returns the connection wrapped in a ClientHelper.
func (sh *ServerHelper) GetConnHelper() *ConnHelper {
	return NewConnHelper(sh.testing, "ws://"+sh.Address)
}

// CloseWait stops the server listener and connections.
// Waits for all the goroutines handling the client connection to quit.
func (sh *ServerHelper) CloseWait() {
	if err := sh.Server.Close(); err != nil {
		sh.testing.Fatal("Failed to stop the server:", err)
	}

	sh.listenerWG.Wait()
	sh.Server.Wait()

	// give connections enough time to disconnect properly
	if os.Getenv("TRAVIS") != "" || os.Getenv("CI") != "" {
		time.Sleep(time.Millisecond * 50)
	} else {
		time.Sleep(time.Millisecond * 5)
	}
}
