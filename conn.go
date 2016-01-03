package neptulon

import (
	"log"
	"time"

	"github.com/neptulon/cmap"
	"github.com/neptulon/shortid"

	"golang.org/x/net/websocket"
)

// Conn is a client connection.
type Conn struct {
	ID         string
	Session    *cmap.CMap
	middleware []func(ctx *ReqCtx) error
	resRoutes  *cmap.CMap // message ID (string) -> handler func(ctx *ResCtx) error : expected responses for requests that we've sent
	ws         *websocket.Conn
	deadline   time.Duration
}

// NewConn creates a new Neptulon connection wrapping given websocket.Conn.
func NewConn(ws *websocket.Conn, middleware []func(ctx *ReqCtx) error) (*Conn, error) {
	id, err := shortid.UUID()
	if err != nil {
		return nil, err
	}

	return &Conn{
		ID:         id,
		Session:    cmap.New(),
		middleware: middleware,
		resRoutes:  cmap.New(),
		ws:         ws,
	}, nil
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
func (c *Conn) SetDeadline(seconds int) {
	c.deadline = time.Second * time.Duration(seconds)
}

// StartReceive starts receiving messages. This method blocks and does not return until the connection is closed.
func (c *Conn) StartReceive() {
	// append the last middleware to request stack, which will write the response to connection, if any
	c.middleware = append(c.middleware, func(ctx *ReqCtx) error {
		if ctx.Res != nil || ctx.Err != nil {
			return ctx.Conn.sendResponse(ctx.ID, ctx.Res, ctx.Err)
		}

		return nil
	})

	for {
		var m message
		err := c.receive(&m)
		if err != nil {
			log.Println("Error while receiving message:", err)
			break
		}

		// if the message is a request
		if m.Method != "" {
			if err := newReqCtx(c, m.ID, m.Method, m.Params, c.middleware).Next(); err != nil {
				log.Println("Error while handling request:", err)
				break
			}
		}

		// if the message is a response
		if resHandler, ok := c.resRoutes.GetOk(m.ID); ok {
			err := resHandler.(func(ctx *ResCtx) error)(newResCtx(c, m.ID, m.Result, m.Error))
			c.resRoutes.Delete(m.ID)
			if err != nil {
				log.Println("Error while handling response:", err)
				break
			}
		} else {
			log.Println("Error while handling response: got response to a request with unknown ID:", m.ID)
			break
		}
	}
}

// SendRequest sends a JSON-RPC request through the connection with an auto generated request ID.
// resHandler is called when a response is returned.
func (c *Conn) SendRequest(method string, params interface{}, resHandler func(res *Response) error) (reqID string, err error) {
	id, err := shortid.UUID()
	if err != nil {
		return "", err
	}

	req := Request{ID: id, Method: method, Params: params}
	if err = c.send(req); err != nil {
		return "", err
	}

	c.resRoutes.Set(req.ID, resHandler)
	return id, nil
}

// SendRequestArr sends a JSON-RPC request through the connection, with array params and auto generated request ID.
// resHandler is called when a response is returned.
func (c *Conn) SendRequestArr(method string, resHandler func(res *Response) error, params ...interface{}) (reqID string, err error) {
	return c.SendRequest(method, params, resHandler)
}

// SendResponse sends a JSON-RPC response message through the connection.
func (c *Conn) sendResponse(id string, result interface{}, err *ResError) error {
	return c.send(Response{ID: id, Result: result, Error: err})
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
