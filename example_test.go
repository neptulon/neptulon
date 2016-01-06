package neptulon_test

import "github.com/neptulon/neptulon"

const debug = false

// Example demonstrating the Neptulon server.
func Example() {
	type echoMsg struct {
		Message string `json:"message"`
	}

	s := neptulon.NewServer("127.0.0.1:3010")

	// echo message body back to the client
	s.Middleware(func(ctx *neptulon.ReqCtx) error {
		var msg interface{}
		if err := ctx.Params(&msg); err != nil {
			return err
		}

		ctx.Res = msg
		return ctx.Next()
	})

	go s.Start()

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

	// ** Output: Server started
}
