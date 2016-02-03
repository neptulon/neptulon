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
// * Send a {"method":"echo", "params":{"message": "..."}} request to client upon first 'echo' request from client,
//   and verify that message body is echoed properly in the response body.
// * Echo any incoming request message body as is within a response message.
// * Repeat ad infinitum, until {"method":"close", "params":"{"message": "..."}"} is received. Close message body is logged.
func TestExternalClient(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	var wg sync.WaitGroup
	m := "Hello from Neptulon server!"

	// handle 'echo' requests via the 'echo middleware'
	srout := middleware.NewRouter()
	sh.Server.Middleware(srout)
	srout.Request("echo", func(ctx *neptulon.ReqCtx) error {
		// send 'echo' request to client upon connection (blocks test if no response is received)
		wg.Add(1)
		ctx.Conn.SendRequest("echo", echoMsg{Message: m}, func(ctx *neptulon.ResCtx) error {
			defer wg.Done()
			var msg echoMsg
			if err := ctx.Result(&msg); err != nil {
				t.Fatal(err)
			}
			if msg.Message != m {
				t.Fatalf("server: expected: %v got: %v", m, msg.Message)
			}
			t.Logf("server: client sent response to our 'echo' request: %v", msg.Message)
			return nil
		})

		// unmarshall incoming message into response directly
		if err := ctx.Params(&ctx.Res); err != nil {
			return err
		}
		return ctx.Next()
	})

	// handle 'close' request (blocks test if no response is received)
	wg.Add(1)
	srout.Request("close", func(ctx *neptulon.ReqCtx) error {
		defer wg.Done()
		if err := ctx.Params(&ctx.Res); err != nil {
			return err
		}
		err := ctx.Next()
		// ctx.Conn.Close() // todo: investigate the error message!!!
		t.Logf("server: closed connection with message from client: %v\n", ctx.Res)
		return err
	})

	defer sh.Start().CloseWait()

	if *ext {
		t.Log("Starter server waiting for external client integration test since.")
		wg.Wait()
		return
	}

	// use internal conn implementation instead to test the test case itself
	t.Log("Skipping external client integration test since -ext flag is not provided.")
	ch := sh.GetConnHelper()
	ch.Conn.MiddlewareFunc(middleware.Logger)

	// handle 'echo' requests via the 'echo middleware'
	crout := middleware.NewRouter()
	ch.Conn.Middleware(crout)
	crout.Request("echo", middleware.Echo)
	defer ch.Connect().CloseWait()

	// handle 'echo' request and send 'close' request upon echo response
	mc := "Hello from Neptulon Go client!"
	ch.SendRequest("echo", echoMsg{Message: mc}, func(ctx *neptulon.ResCtx) error {
		var msg echoMsg
		if err := ctx.Result(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Message != mc {
			t.Fatalf("client: expected: %v got: %v", mc, msg.Message)
		}
		t.Log("client: server accepted and echoed 'echo' request message body")

		// send close request after getting our echo message back
		mb := "Thanks for echoing! Over and out."
		ch.SendRequest("close", echoMsg{Message: mb}, func(ctx *neptulon.ResCtx) error {
			var msg echoMsg
			if err := ctx.Result(&msg); err != nil {
				t.Fatal(err)
			}
			if msg.Message != mb {
				t.Fatalf("client: expected: %v got: %v", mb, msg.Message)
			}
			t.Log("client: server accepted and echoed 'close' request message body. bye!")
			return nil
		})

		return nil
	})
}
