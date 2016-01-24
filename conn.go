package neptulon

import (
	"errors"
	"io"
	"log"
	"net"
	"runtime"
	"sync"
	"time"

	"github.com/neptulon/cmap"
	"github.com/neptulon/shortid"

	"golang.org/x/net/websocket"
)

// Conn is a client connection.
type Conn struct {
	ID           string
	Session      *cmap.CMap
	middleware   []func(ctx *ReqCtx) error
	resRoutes    *cmap.CMap // message ID (string) -> handler func(ctx *ResCtx) error : expected responses for requests that we've sent
	ws           *websocket.Conn
	wg           sync.WaitGroup
	deadline     time.Duration
	isClientConn bool
	connected    bool
}

// NewConn creates a new Conn object.
func NewConn() (*Conn, error) {
	id, err := shortid.UUID()
	if err != nil {
		return nil, err
	}

	return &Conn{
		ID:        id,
		Session:   cmap.New(),
		resRoutes: cmap.New(),
		deadline:  time.Second * time.Duration(300),
	}, nil
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
// Default value for read/write deadline is 300 seconds.
func (c *Conn) SetDeadline(seconds int) {
	c.deadline = time.Second * time.Duration(seconds)
}

// Middleware registers middleware to handle incoming request messages.
func (c *Conn) Middleware(middleware ...func(ctx *ReqCtx) error) {
	c.middleware = append(c.middleware, middleware...)
}

// Connect connects to the given WebSocket server.
func (c *Conn) Connect(addr string) error {
	ws, err := websocket.Dial(addr, "", "http://localhost")
	if err != nil {
		return err
	}

	c.ws = ws
	c.connected = true
	c.wg.Add(1)
	go func() {
		defer recoverAndLog(c, &c.wg)
		c.startReceive()
	}()
	time.Sleep(time.Millisecond) // give receive goroutine a few cycles to start
	return nil
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	if c.ws == nil {
		return nil
	}

	return c.ws.RemoteAddr()
}

// SendRequest sends a JSON-RPC request through the connection with an auto generated request ID.
// resHandler is called when a response is returned.
func (c *Conn) SendRequest(method string, params interface{}, resHandler func(res *ResCtx) error) (reqID string, err error) {
	id, err := shortid.UUID()
	if err != nil {
		return "", err
	}

	req := request{ID: id, Method: method, Params: params}
	if err = c.send(req); err != nil {
		return "", err
	}

	c.resRoutes.Set(req.ID, resHandler)
	return id, nil
}

// SendRequestArr sends a JSON-RPC request through the connection, with array params and auto generated request ID.
// resHandler is called when a response is returned.
func (c *Conn) SendRequestArr(method string, resHandler func(res *ResCtx) error, params ...interface{}) (reqID string, err error) {
	return c.SendRequest(method, params, resHandler)
}

// Close closes the connection.
func (c *Conn) Close() error {
	c.connected = false
	return c.ws.Close()
}

// SendResponse sends a JSON-RPC response message through the connection.
func (c *Conn) sendResponse(id string, result interface{}, err *ResError) error {
	return c.send(response{ID: id, Result: result, Error: err})
}

// Send sends the given message through the connection.
func (c *Conn) send(msg interface{}) error {
	if !c.connected {
		return errors.New("use of closed connection")
	}

	if err := c.ws.SetWriteDeadline(time.Now().Add(c.deadline)); err != nil {
		return err
	}

	return websocket.JSON.Send(c.ws, msg)
}

// Receive receives message from the connection.
func (c *Conn) receive(msg *message) error {
	if !c.connected {
		return errors.New("use of closed connection")
	}

	if err := c.ws.SetReadDeadline(time.Now().Add(c.deadline)); err != nil {
		return err
	}

	return websocket.JSON.Receive(c.ws, &msg)
}

// UseConn reuses an established websocket.Conn.
// This function blocks and does not return until the connection is closed by another goroutine.
func (c *Conn) useConn(ws *websocket.Conn) {
	c.ws = ws
	c.connected = true
	c.startReceive()
}

// startReceive starts receiving messages. This method blocks and does not return until the connection is closed.
func (c *Conn) startReceive() {
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
			// if we closed the connection
			if !c.connected {
				log.Printf("conn: closed %v: %v", c.ID, c.RemoteAddr())
				break
			}

			// if peer closed the connection
			if err == io.EOF {
				log.Printf("conn: peer disconnected %v: %v", c.ID, c.RemoteAddr())
				break
			}

			log.Printf("conn: error while receiving message: %v", err)
			break
		}

		// if the message is a request
		if m.Method != "" {
			c.wg.Add(1)
			go func() {
				defer recoverAndLog(c, &c.wg)
				if err := newReqCtx(c, m.ID, m.Method, m.Params, c.middleware).Next(); err != nil {
					log.Printf("conn: error while handling request: %v", err)
					c.Close()
				}
			}()

			continue
		}

		// if the message is not a JSON-RPC message
		if m.ID == "" || (m.Result == nil && m.Error == nil) {
			log.Printf("conn: received an unknown message %v: %v\n%v", c.ID, c.RemoteAddr(), m)
			break
		}

		// if the message is a response
		if resHandler, ok := c.resRoutes.GetOk(m.ID); ok {
			c.wg.Add(1)
			go func() {
				defer recoverAndLog(c, &c.wg)
				err := resHandler.(func(ctx *ResCtx) error)(newResCtx(c, m.ID, m.Result, m.Error))
				c.resRoutes.Delete(m.ID)
				if err != nil {
					log.Printf("conn: error while handling response: %v", err)
					c.Close()
				}
			}()
		} else {
			log.Printf("conn: error while handling response: got response to a request with unknown ID: %v", m.ID)
			break
		}
	}

	c.Close()
}

func recoverAndLog(c *Conn, wg *sync.WaitGroup) {
	if err := recover(); err != nil {
		const size = 64 << 10
		buf := make([]byte, size)
		buf = buf[:runtime.Stack(buf, false)]
		log.Printf("conn: panic handling response %v: %v\n%s", c.RemoteAddr(), err, buf)
	}
	wg.Done()
}
