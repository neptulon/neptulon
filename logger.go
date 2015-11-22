package neptulon

// todo: separete this into its own repo

// Logger provides low level request logging, performance metrics, and other metrics data.
type Logger struct{}

func perfLoggerMiddleware(conn Conn, msg []byte) {
}

func requestResponseLoggerMiddleware(conn Conn, msg []byte) {
}
