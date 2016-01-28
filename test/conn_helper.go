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
	resWG      sync.WaitGroup // to be able to blocking wait for pending responses
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

	ch := &ConnHelper{Conn: conn, testing: t, serverAddr: addr}
	ch.Conn.SetDeadline(10)
	return ch
}

// Middleware registers middleware to handle incoming request messages.
func (ch *ConnHelper) Middleware(middleware ...func(ctx *neptulon.ReqCtx) error) {
	ch.Conn.Middleware(middleware...)
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

// SendRequest sends a JSON-RPC request through the client connection with an auto generated request ID.
// resHandler is called when a response is returned.
func (ch *ConnHelper) SendRequest(method string, params interface{}, resHandler func(ctx *neptulon.ResCtx) error) *ConnHelper {
	ch.resWG.Add(1)
	_, err := ch.Conn.SendRequest(method, params, func(ctx *neptulon.ResCtx) error {
		defer ch.resWG.Done()
		return resHandler(ctx)
	})

	if err != nil {
		ch.testing.Fatal("Failed to send request:", err)
	}

	return ch
}

// CloseWait closes a connection.
// Waits till all the goroutines handling messages quit.
func (ch *ConnHelper) CloseWait() {
	ch.resWG.Wait()
	if err := ch.Conn.Close(); err != nil {
		ch.testing.Fatal("Failed to close connection:", err)
	}
	ch.Conn.Wait()
	time.Sleep(time.Millisecond * 5)
}
