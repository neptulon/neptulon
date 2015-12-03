package neptulon

import "github.com/neptulon/client"

// todo: separete this into its own repo

// Logger provides low level request logging, performance metrics, and other metrics data.
type Logger struct{}

func perfLoggerMiddleware(ctx *client.Ctx) {
}

func messageLoggerMiddleware(ctx *client.Ctx) {
}
