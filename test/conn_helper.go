package test

import (
	"net"
	"os"
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
}

// NewConnHelper creates a new client helper object.
func NewConnHelper(t *testing.T, addr string) *ConnHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode.")
	}

	conn, err := neptulon.NewConn()
	if err != nil {
		t.Fatal("Failed to create connection:", err)
	}

	conn.SetDeadline(10)
	return &ConnHelper{Conn: conn, testing: t, serverAddr: addr}
}

// Connect connects to a server.
func (ch *ConnHelper) Connect() *ConnHelper {
	// retry connect in case we're operating on a very slow machine
	for i := 0; i <= 5; i++ {
		if err := ch.Conn.Connect(ch.serverAddr); err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "dial" && operr.Err.Error() == "connection refused" {
				time.Sleep(time.Millisecond * 50)
				continue
			} else if i == 5 {
				ch.testing.Fatalf("Cannot connect to server address %v after 5 retries, with error: %v", ch.serverAddr, err)
			}
			ch.testing.Fatalf("Cannot connect to server address %v with error: %v", ch.serverAddr, err)
		}

		if i != 0 {
			ch.testing.Logf("WARNING: it took %v retries to connect to the server, which might indicate code issues or slow machine.", i)
		}

		break
	}

	return ch
}

// SendRequestSync sends a JSON-RPC request through the client connection with an auto generated request ID.
// resHandler is called when a response is returned.
// This function is synchronous and blocking.
func (ch *ConnHelper) SendRequestSync(method string, params interface{}, resHandler func(ctx *neptulon.ResCtx) error) *ConnHelper {
	gotRes := make(chan bool)

	_, err := ch.Conn.SendRequest(method, params, func(ctx *neptulon.ResCtx) error {
		defer func() { gotRes <- true }()
		return resHandler(ctx)
	})

	if err != nil {
		ch.testing.Fatal("Failed to send request:", err)
	}

	select {
	case <-gotRes:
	case <-time.After(time.Second * 3):
		ch.testing.Fatal("did not get a response in time")
	}

	return ch
}

// CloseWait closes a connection.
// Waits till all the goroutines handling messages quit.
func (ch *ConnHelper) CloseWait() {
	if err := ch.Conn.Close(); err != nil {
		ch.testing.Fatal("Failed to close connection:", err)
	}
	ch.Conn.Wait()

	if os.Getenv("TRAVIS") != "" || os.Getenv("CI") != "" {
		time.Sleep(time.Millisecond * 50)
	} else {
		time.Sleep(time.Millisecond * 5)
	}
}
