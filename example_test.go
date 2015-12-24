package neptulon_test

import (
	"log"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/client"
)

// Example demonstrating the Neptulon server.
func Example() {
	s, err := neptulon.NewTCPServer("127.0.0.1:3001", false)
	if err != nil {
		log.Fatalln("Failed to start Neptulon server:", err)
	}

	// middleware for echoing all incoming messages as is
	s.MiddlewareIn(func(ctx *client.Ctx) error {
		if err := ctx.Client.Send(ctx.Msg); err != nil {
			return err
		}

		return ctx.Next()
	})

	s.Start()
}
