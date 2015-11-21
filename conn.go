package neptulon

import "github.com/neptulon/cmap"

// Conn is a generic stream-oriented network connection wrapper.
type Conn interface {
	ID() string
	Data() *cmap.CMap
	SetReadDeadline(seconds int)
	Read() (msg []byte, err error)
	Write(msg []byte) error
	Close() error
}
