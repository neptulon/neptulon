package test

import (
	"crypto/x509/pkix"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/ca"
	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/client"
)

const (
	host, port = "127.0.0.1", "3001"
	laddr      = host + ":" + port
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
}

// NewTCPServerHelper creates a new TCP server helper object.
func NewTCPServerHelper(t *testing.T) *ServerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	server, err := neptulon.NewTCPServer(laddr, false)
	if err != nil {
		t.Fatal("Failed to create server:", err)
	}

	return &ServerHelper{
		Server:  server,
		Address: laddr,

		testing: t,
	}
}

// NewTLSServerHelper creates a new server helper object with Transport Layer Security.
func NewTLSServerHelper(t *testing.T) *ServerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	// generate TLS certs
	certChain, err := ca.GenCertChain("FooBar", host, host, time.Hour, 512)
	if err != nil {
		t.Fatal("Failed to create TLS certificate chain:", err)
	}

	server, err := neptulon.NewTLSServer(certChain.ServerCert, certChain.ServerKey, certChain.IntCACert, laddr, false)
	if err != nil {
		t.Fatal("Failed to create server:", err)
	}

	return &ServerHelper{
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
}

// MiddlewareIn registers middleware to handle incoming messagesh.
func (sh *ServerHelper) MiddlewareIn(middleware ...func(ctx *client.Ctx)) *ServerHelper {
	sh.Server.MiddlewareIn(middleware...)
	return sh
}

// MiddlewareOut registers middleware to handle/intercept outgoing messages before they are sent.
func (sh *ServerHelper) MiddlewareOut(middleware ...func(ctx *client.Ctx)) *ServerHelper {
	sh.Server.MiddlewareOut(middleware...)
	return sh
}

// Start starts the server.
func (sh *ServerHelper) Start() *ServerHelper {
	// start the server immediately
	sh.serverWG.Add(1)
	go func() {
		defer sh.serverWG.Done()
		if err := sh.Server.Start(); err != nil {
			sh.testing.Fatal("Failed to accept connection(s):", err)
		}
	}()

	time.Sleep(time.Millisecond) // give Accept() enough CPU cycles to initiate
	return sh
}

// GetTCPClientHelper creates a client connection to this server instance using TCP and returns the connection wrapped in a ClientHelper.
func (sh *ServerHelper) GetTCPClientHelper() *ClientHelper {
	return NewClientHelper(sh.testing, sh.Address)
}

// GetTLSClientHelper creates a client connection to this server instance using TLS and returns the connection wrapped in a ClientHelper.
func (sh *ServerHelper) GetTLSClientHelper() *ClientHelper {
	cert, key, err := ca.GenClientCert(pkix.Name{
		Organization: []string{"FooBar"},
		CommonName:   "1",
	}, time.Hour, 512, sh.IntCACert, sh.IntCAKey)

	if err != nil {
		sh.testing.Fatal(err)
	}

	return sh.GetTCPClientHelper().UseTLS(sh.IntCACert, cert, key)
}

// Close stops the server listener and connectionsh.
func (sh *ServerHelper) Close() {
	if err := sh.Server.Close(); err != nil {
		sh.testing.Fatal("Failed to stop the server:", err)
	}

	sh.serverWG.Wait()
}
