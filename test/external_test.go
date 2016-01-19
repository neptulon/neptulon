package test

import (
	"flag"
	"testing"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
)

var ext = flag.Bool("ext", false, "Run external client test case.")

// Helper method for testing client implementations in other languages.
// Flow of events for this function is:
// * Wait to receive any request message.
// * Echo the message body as a response.
// * Send an {"method":"echo", "params":"Lorem ip sum..."} request to client.
// * Wait for response and verify that message body is echoed properly in the response body.
// * Repeat ad infinitum, until {"method":"close", "params":"..."} is received. Close message body is logged.

func TestExternalClient(t *testing.T) {
	sh := NewServerHelper(t).Start()
	defer sh.Close()

	for {
		rout := middleware.NewRouter()
		sh.Middleware(rout.Middleware)
		rout.Request("echo", middleware.Echo)

		if !*ext {
			t.Log("Skipping external client integration test since -ext flag is not provided.")

			// use internal conn implementation instead to test the test case itself
			ch := sh.GetConnHelper().Connect()
			defer ch.Close()

			ch.SendRequest("echo", echoMsg{Message: "Hello!"}, func(ctx *neptulon.ResCtx) error {
				var msg echoMsg
				if err := ctx.Result(&msg); err != nil {
					t.Fatal(err)
				}
				if msg.Message != "Hello!" {
					t.Fatalf("expected: %v got: %v", "Hello!", msg.Message)
				}
				return nil
			})
		}
	}
}
