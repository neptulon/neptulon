package neptulon

// Conn is a generic stream-oriented network connection wrapper.
type Conn interface {
	Read() (msg []byte, err error)
	SetReadDeadline(seconds int)
	Write(msg []byte) error
	Close() error
	ID() string
}
