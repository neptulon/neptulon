package neptulon

import (
	"time"

	"github.com/neptulon/cmap"
	"github.com/neptulon/shortid"

	"golang.org/x/net/websocket"
)

// Conn is a client connection.
type Conn struct {
	ID      string
	Session *cmap.CMap

	ws       *websocket.Conn
	deadline time.Duration
}

// NewConn creates a new Neptulon connection wrapping given websocket.Conn.
func NewConn(ws *websocket.Conn, reqMiddleware []func(ctx *ReqCtx) error, resMiddleware []func(ctx *ResCtx) error) (*Conn, error) {
	id, err := shortid.UUID()
	if err != nil {
		return nil, err
	}

	return &Conn{ID: id, Session: cmap.New()}, nil
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
func (c *Conn) SetDeadline(seconds int) {
	c.deadline = time.Second * time.Duration(seconds)
}

// Send sends the given message through the connection.
func (c *Conn) send(msg interface{}) error {
	if err := c.ws.SetWriteDeadline(time.Now().Add(c.deadline)); err != nil {
		return err
	}

	return websocket.JSON.Send(c.ws, msg)
}

// Receive receives message from the connection.
func (c *Conn) receive(msg *message) error {
	if err := c.ws.SetReadDeadline(time.Now().Add(c.deadline)); err != nil {
		return err
	}

	return websocket.JSON.Receive(c.ws, &msg)
}
