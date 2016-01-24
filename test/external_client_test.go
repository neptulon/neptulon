package test

import (
	"flag"
	"sync"
	"testing"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
)

var ext = flag.Bool("ext", false, "Run external client test case.")

// Helper method for testing client implementations in other languages.
// Flow of events for this function is:
// * Send a {"method":"echo", "params":{"message": "Hello!"}} request to client upon connection,
//   and verify that message body is echoed properly in the response body.
// * Echo any incoming request message body as is within a response message.
// * Repeat ad infinitum, until {"method":"close", "params":"..."} is received. Close message body is logged.
func TestExternalClient(t *testing.T) {
	sh := NewServerHelper(t).Start()
	defer sh.Close()

	var wg sync.WaitGroup
	wg.Add(2) // one for response handler below, other for "close" request handler

	sh.Server.ConnHandler(func(c *neptulon.Conn) error {
		c.SendRequest("echo", echoMsg{Message: "Hello!"}, handleEchoRes(t))
		return nil
	})

	rout := middleware.NewRouter()
	sh.Middleware(rout.Middleware)
	rout.Request("echo", middleware.Echo)

	rout.Request("close", func(ctx *neptulon.ReqCtx) error {
		defer wg.Done()
		var body interface{}
		if err := ctx.Params(&body); err != nil {
			return err
		}
		t.Logf("Closed connection with message from client: %v\n", body)
		ctx.Conn.Close()
		return nil
	})

	// use internal conn implementation instead to test the test case itself
	if !*ext {
		t.Log("Skipping external client integration test since -ext flag is not provided.")
		ch := sh.GetConnHelper().Connect()
		defer ch.Close()

		ch.SendRequest("echo", echoMsg{Message: "Hello!"}, handleEchoRes(t))
		ch.SendRequest("close", echoMsg{Message: "Thanks for echoing! Over and out."}, handleEchoRes(t))
	}

	wg.Wait()
}
