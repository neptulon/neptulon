package neptulon

import "github.com/neptulon/conn-go"

// todo: separete this into its own repo

// Logger provides low level request logging, performance metrics, and other metrics data.
type Logger struct{}

func perfLoggerMiddleware(ctx *conn.Ctx) {
}

func messageLoggerMiddleware(ctx *conn.Ctx) {
}
