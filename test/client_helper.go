package test

import (
	"testing"

	"github.com/neptulon/client"
)

// ConnHelper is a client.Client wrapper.
// All the functions are wrapped with proper test runner error logging.
type ConnHelper struct {
	client    *client.Client
	server    *ServerHelper // server that this connection will be made to
	testing   *testing.T
	cert, key []byte
}

// NewConnHelper creates a new connection helper object.
// Takes target server as an argument to retrieve server certs, address, etc.
func NewConnHelper(t *testing.T, s *ServerHelper) *ConnHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	return &ConnHelper{testing: t, server: s}
}
