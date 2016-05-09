package neptulon_test

import (
	"fmt"
	"log"
	"time"

	"github.com/neptulon/neptulon"
)

const debug = false

// Example demonstrating the Neptulon server.
func Example() {
	type SampleMsg struct {
		Message string `json:"message"`
	}

	// start the server and echo incoming messages back to the sender
	s := neptulon.NewServer("127.0.0.1:3000")
	s.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		var msg SampleMsg
		if err := ctx.Params(&msg); err != nil {
			return err
		}
		ctx.Res = msg
		return ctx.Next()
	})
	go s.ListenAndServe()
	defer s.Close()

	time.Sleep(time.Millisecond * 50) // let server goroutine to warm up

	// connect to the server and send a message
	c, err := neptulon.NewConn()
	if err != nil {
		log.Fatal(err)
	}
	if err := c.Connect("ws://127.0.0.1:3000"); err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	_, err = c.SendRequest("echo", SampleMsg{Message: "Hello!"}, func(ctx *neptulon.ResCtx) error {
		var msg SampleMsg
		if err := ctx.Result(&msg); err != nil {
			return err
		}
		fmt.Println("Server says:", msg.Message)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Millisecond * 50) // wait to get an answer

	// Output: Server says: Hello!
}
