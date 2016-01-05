package test

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/neptulon"
)

// ConnHelper is a Neptulon Conn wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ConnHelper struct {
	Conn *neptulon.Conn

	testing    *testing.T
	serverAddr string
	msgWG      sync.WaitGroup
}

// NewConnHelper creates a new client helper object.
func NewConnHelper(t *testing.T, addr string) *ConnHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode.")
	}

	ch := &ConnHelper{testing: t, serverAddr: addr}
	ch.Conn = neptulon.NewConn(&ch.msgWG, nil)
	ch.Conn.SetDeadline(10)
	return ch
}

// MiddlewareIn registers middleware to handle incoming messagesh.
func (ch *ConnHelper) MiddlewareIn(middleware ...func(ctx *neptulon.Ctx) error) *ConnHelper {
	ch.Conn.MiddlewareIn(middleware...)
	return ch
}

// MiddlewareOut registers middleware to handle/intercept outgoing messages before they are sent.
func (ch *ConnHelper) MiddlewareOut(middleware ...func(ctx *neptulon.Ctx) error) *ConnHelper {
	ch.Conn.MiddlewareOut(middleware...)
	return ch
}

// UseTLS connects to server using TLS.
func (ch *ConnHelper) UseTLS(serverCA, clientCert, clientCertKey []byte) *ConnHelper {
	ch.Conn.UseTLS(serverCA, clientCert, clientCertKey)
	return ch
}

// Connect connects to a server.
func (ch *ConnHelper) Connect() *ConnHelper {
	// retry connect in case we're operating on a very slow machine
	for i := 0; i <= 5; i++ {
		if err := ch.Conn.Connect(ch.addr, false); err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "dial" && operr.Err.Error() == "connection refused" {
				time.Sleep(time.Millisecond * 50)
				continue
			} else if i == 5 {
				ch.testing.Fatalf("Cannot connect to server address %v after 5 retries, with error: %v", ch.addr, err)
			}
			ch.testing.Fatalf("Cannot connect to server address %v with error: %v", ch.addr, err)
		}

		if i != 0 {
			ch.testing.Logf("WARNING: it took %v retries to connect to the server, which might indicate code issues or slow machine.", i)
		}

		break
	}

	return ch
}

// Send sends a message to connected peer.
func (ch *ConnHelper) Send(msg []byte) {
	if err := ch.Conn.Send(msg); err != nil {
		ch.testing.Fatal("Error while sending message to peer:", err)
	}

	n := len(msg)
	if n < 100 {
		ch.testing.Logf("Sent message to listener from client: %v (%v bytes)", string(msg), n)
	} else {
		ch.testing.Logf("Sent message to listener from client: (...messages longer than 100 characters not shown...) (%v bytes)", n)
	}
}

// Close closes a connection.
func (ch *ConnHelper) Close() {
	if err := ch.Conn.Close(); err != nil {
		ch.testing.Fatal("Failed to close connection:", err)
	}

	ch.msgWG.Wait()
}
