package test

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/client"
)

// ClientHelper is a client.Client wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ClientHelper struct {
	client    *client.Client
	testing   *testing.T
	msgWG     sync.WaitGroup
	cert, key []byte
}

// NewClientHelper creates a new client helper object.
// Takes target server as an argument to retrieve server certs, address, etc.
func NewClientHelper(t *testing.T) *ClientHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	ch := &ClientHelper{testing: t}
	ch.client = client.NewClient(&ch.msgWG, nil)
	return ch
}

// DialTLS initiates a TLS connection.
func (c *ClientHelper) DialTLS() *ClientHelper {
	// retry connect in case we're operating on a very slow machine
	for i := 0; i <= 5; i++ {
		conn, err := client.DialTLS(addr, c.server.IntCACert, c.cert, c.key, false)
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "dial" && operr.Err.Error() == "connection refused" {
				time.Sleep(time.Millisecond * 50)
				continue
			} else if i == 5 {
				c.testing.Fatalf("Cannot connect to server address %v after 5 retries, with error: %v", addr, err)
			}
			c.testing.Fatalf("Cannot connect to server address %v with error: %v", addr, err)
		}

		if i != 0 {
			c.testing.Logf("WARNING: it took %v retries to connect to the server, which might indicate code issues or slow machine.", i)
		}

		conn.SetReadDeadline(10)
		c.conn = conn
		return c
	}
}

// VerifyConnClosed verifies that the connection is in closed state.
// Verification is done via reading from the channel and checking that returned error is io.EOF.
func (c *ConnHelper) VerifyConnClosed() bool {
	_, _, _, err := c.conn.ReadMsg(nil, nil)
	if err != io.EOF {
		return false
	}

	return true
}

// Close closes a connection.
func (c *ConnHelper) Close() {
	if err := c.conn.Close(); err != nil {
		c.testing.Fatal("Failed to close connection:", err)
	}
}
