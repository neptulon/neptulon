package test

import (
	"testing"

	"github.com/neptulon/client"
)

// ClientHelper is a client.Client wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ClientHelper struct {
	client    *client.Client
	server    *ServerHelper // server that this connection will be made to
	testing   *testing.T
	cert, key []byte
}

// NewClientHelper creates a new client helper object.
// Takes target server as an argument to retrieve server certs, address, etc.
func NewClientHelper(t *testing.T, s *ServerHelper) *ClientHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	return &ClientHelper{testing: t, server: s}
}

// Dial initiates a TLS connection.
func (c *ClientHelper) DialTLS() *ClientHelper {

}
