package neptulon_test

import (
	"log"

	"github.com/neptulon/client"
	"github.com/neptulon/neptulon"
)

// Example demonstrating the Neptulon server.
func Example() {
	s, err := neptulon.NewTCPServer("127.0.0.1:3001", false)
	if err != nil {
		log.Fatalln("Failed to start Neptulon server:", err)
	}

	// middleware for echoing all incoming messages as is
	s.MiddlewareIn(func(ctx *client.Ctx) {
		ctx.Client.Send(ctx.Msg)
		ctx.Next()
	})

	s.Start()
}
