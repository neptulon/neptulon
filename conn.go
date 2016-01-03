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

	reqMiddleware []func(ctx *ReqCtx) error
	resMiddleware []func(ctx *ResCtx) error
	resRoutes     *cmap.CMap // message ID (string) -> handler func(ctx *ResCtx) error : expected responses for requests that we've sent
	ws            *websocket.Conn
	deadline      time.Duration
}

// NewConn creates a new Neptulon connection wrapping given websocket.Conn.
func NewConn(ws *websocket.Conn, reqMiddleware []func(ctx *ReqCtx) error, resMiddleware []func(ctx *ResCtx) error) (*Conn, error) {
	id, err := shortid.UUID()
	if err != nil {
		return nil, err
	}

	// append the last middleware to request stack, which will write the response to connection, if any
	reqMW := append(reqMiddleware, func(ctx *ReqCtx) error {
		if ctx.Res != nil || ctx.Err != nil {
			return ctx.Conn.send(&Response{ID: ctx.id, Result: ctx.Res, Error: ctx.Err})
		}

		return nil
	})

	resRoutes := cmap.New()

	// append the last middleware to response stack, which will read the response for a previous request, if any
	resMW := append(resMiddleware, func(ctx *ResCtx) error {
		if resHandler, ok := resRoutes.GetOk(ctx.id); ok {
			err := resHandler.(func(ctx *ResCtx) error)(ctx)
			resRoutes.Delete(ctx.id)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return &Conn{ID: id, Session: cmap.New(), reqMiddleware: reqMW, resMiddleware: resMW, resRoutes: resRoutes, ws: ws}, nil
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
