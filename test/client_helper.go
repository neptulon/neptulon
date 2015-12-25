package test

import (
	"net"
	"sync"
	"testing"
	"time"

	"github.com/neptulon/neptulon/client"
)

// ClientHelper is a client.Client wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ClientHelper struct {
	Client *client.Client

	testing *testing.T
	addr    string
	msgWG   sync.WaitGroup
}

// NewClientHelper creates a new client helper object.
func NewClientHelper(t *testing.T, addr string) *ClientHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	ch := &ClientHelper{testing: t, addr: addr}
	ch.Client = client.NewClient(&ch.msgWG, nil)
	ch.Client.SetDeadline(10)
	return ch
}

// MiddlewareIn registers middleware to handle incoming messagesh.
func (ch *ClientHelper) MiddlewareIn(middleware ...func(ctx *client.Ctx) error) *ClientHelper {
	ch.Client.MiddlewareIn(middleware...)
	return ch
}

// MiddlewareOut registers middleware to handle/intercept outgoing messages before they are sent.
func (ch *ClientHelper) MiddlewareOut(middleware ...func(ctx *client.Ctx) error) *ClientHelper {
	ch.Client.MiddlewareOut(middleware...)
	return ch
}

// UseTLS connects to server using TLS.
func (ch *ClientHelper) UseTLS(serverCA, clientCert, clientCertKey []byte) *ClientHelper {
	ch.Client.UseTLS(serverCA, clientCert, clientCertKey)
	return ch
}

// Connect connects to a server.
func (ch *ClientHelper) Connect() *ClientHelper {
	// retry connect in case we're operating on a very slow machine
	for i := 0; i <= 5; i++ {
		if err := ch.Client.Connect(ch.addr, false); err != nil {
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
func (ch *ClientHelper) Send(msg []byte) {
	if err := ch.Client.Send(msg); err != nil {
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
func (ch *ClientHelper) Close() {
	if err := ch.Client.Close(); err != nil {
		ch.testing.Fatal("Failed to close connection:", err)
	}

	ch.msgWG.Wait()
}
