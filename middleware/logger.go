package logger

import "github.com/neptulon/neptulon"

// Logger is an incoming/outgoing message logger.
func Logger(ctx *neptulon.Ctx) {
	// todo: evaluate options for minimal performance impact
}

// Perf is a performance logger for logging request/response times.
func Perf(ctx *neptulon.Ctx) {
	// todo: this chould an extensible Perf package also..
}
