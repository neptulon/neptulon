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
// * Send a {"method":"echo", "params":{"message": "..."}} request to client upon connection,
//   and verify that message body is echoed properly in the response body.
// * Echo any incoming request message body as is within a response message.
// * Repeat ad infinitum, until {"method":"close", "params":"{"message": "..."}"} is received. Close message body is logged.
func TestExternalClient(t *testing.T) {
	sh := NewServerHelper(t).Start()
	defer sh.CloseWait()
	var wg sync.WaitGroup
	wg.Add(1) // one for response handler below, other for "close" request handler

	m := "Hello!"

	sh.Server.ConnHandler(func(c *neptulon.Conn) error {
		c.SendRequest("echo", echoMsg{Message: m}, func(ctx *neptulon.ResCtx) error {
			defer wg.Done()
			var msg echoMsg
			if err := ctx.Result(&msg); err != nil {
				t.Fatal(err)
			}
			if msg.Message != m {
				t.Fatalf("expected: %v got: %v", m, msg.Message)
			}
			return nil
		})
		return nil
	})

	rout := middleware.NewRouter()
	sh.Middleware(rout.Middleware)
	rout.Request("echo", middleware.Echo)

	rout.Request("close", func(ctx *neptulon.ReqCtx) error {
		defer wg.Done()
		if err := ctx.Params(&ctx.Res); err != nil {
			return err
		}
		err := ctx.Next()
		ctx.Conn.Close()
		t.Logf("test: closed connection with message from client: %v\n", ctx.Res)
		return err
	})

	if *ext {
		wg.Wait()
		return
	}

	// use internal conn implementation instead to test the test case itself
	t.Log("Skipping external client integration test since -ext flag is not provided.")
	ch := sh.GetConnHelper().Connect()
	defer ch.CloseWait()
	cm := "Thanks for echoing! Over and out."

	ch.SendRequest("echo", echoMsg{Message: m}, func(ctx *neptulon.ResCtx) error {
		var msg echoMsg
		if err := ctx.Result(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Message != m {
			t.Fatalf("expected: %v got: %v", m, msg.Message)
		}
		t.Log("client: server accepted and echoed 'echo' request message body")

		// send close request after getting our echo message back
		ch.SendRequest("close", echoMsg{Message: cm}, func(ctx *neptulon.ResCtx) error {
			var msg echoMsg
			if err := ctx.Result(&msg); err != nil {
				t.Fatal(err)
			}
			if msg.Message != cm {
				t.Fatalf("expected: %v got: %v", cm, msg.Message)
			}
			t.Log("client: server accepted and echoed 'close' request message body. bye!")
			return nil
		})

		return nil
	})

	wg.Wait()
}
