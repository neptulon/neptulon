package neptulon

// Listener is a generic network listener for stream-oriented protocols.
type Listener interface {
	SetReadDeadline(seconds int)
	Accept(handleConn func(conn Conn), handleMsg func(conn Conn, msg []byte), handleDisconn func(conn Conn)) error
	Close() error
}
