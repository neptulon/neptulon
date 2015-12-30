package neptulon

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/neptulon/cmap"
	"github.com/neptulon/shortid"
)

// Client is a Neptulon client.
type Client struct {
	Conn *Conn // Low level client connection object. Avoid using this unless you need low level read/writes directly to the connection for testing.

	connID  string
	session *cmap.CMap

	// middleware for incoming and outgoing messages
	middlewareIn  []func(ctx *Ctx) error
	middlewareOut []func(ctx *Ctx) error

	disconnHandler func(client *Client)
	msgWG          *sync.WaitGroup
	deadline       time.Duration

	tls                           bool
	ca, clientCert, clientCertKey []byte
}

// NewClient creates a new Client object.
// msgWG = (optional) sets the given *sync.WaitGroup reference to be used for counting active gorotuines that are used for handling incoming/outgoing messages.
// disconnHandler = (optional) registers a function to handle client disconnection events.
func NewClient(msgWG *sync.WaitGroup, disconnHandler func(client *Client)) *Client {
	if msgWG == nil {
		msgWG = &sync.WaitGroup{}
	}

	if disconnHandler == nil {
		disconnHandler = func(client *Client) {}
	}

	return &Client{
		session:        cmap.New(),
		msgWG:          msgWG,
		disconnHandler: disconnHandler,
	}
}

// ConnID is a randomly generated unique client connection ID.
func (c *Client) ConnID() string {
	return c.connID
}

// Session is a thread-safe data store for storing arbitrary data for this connection session.
func (c *Client) Session() *cmap.CMap {
	return c.session
}

// MiddlewareIn registers middleware to handle incoming messages.
func (c *Client) MiddlewareIn(middleware ...func(ctx *Ctx) error) {
	c.middlewareIn = append(c.middlewareIn, middleware...)
}

// MiddlewareOut registers middleware to handle/intercept outgoing messages before they are sent.
func (c *Client) MiddlewareOut(middleware ...func(ctx *Ctx) error) {
	c.middlewareOut = append(c.middlewareOut, middleware...)
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
func (c *Client) SetDeadline(seconds int) {
	c.deadline = time.Second * time.Duration(seconds)
}

// UseTLS enables Transport Layer Security for the connection.
// ca = Optional CA certificate to be used for verifying the server certificate. Useful for using self-signed server certificates.
// clientCert, clientCertKey = Optional certificate/privat key pair for TLS client certificate authentication.
// All certificates/private keys are in PEM encoded X.509 format.
func (c *Client) UseTLS(ca, clientCert, clientCertKey []byte) {
	c.tls = true
	c.ca = ca
	c.clientCert = clientCert
	c.clientCertKey = clientCertKey
}

// Connect connectes to the server at given network address and starts receiving messages.
func (c *Client) Connect(addr string, debug bool) error {
	var conn *Conn
	var err error

	if c.tls {
		conn, err = dialTLS(addr, c.ca, c.clientCert, c.clientCertKey, debug)

		// Conn has the certificates parsed so free up the memory as the PEM encoded X.509 certificates can be quite big
		c.ca = nil
		c.clientCert = nil
		c.clientCertKey = nil
	} else {
		conn, err = dialTCP(addr, debug)
	}

	if err != nil {
		return err
	}

	return c.useConn(conn)
}

// UseTCPConn reuses an established *net.TCPConn and starts receiving messages.
func (c *Client) UseTCPConn(conn *net.TCPConn, debug bool) error {
	tcpc, err := newConn(conn, false, debug)
	if err != nil {
		return err
	}

	return c.useConn(tcpc)
}

// UseTLSConn reuses an established *tls.Conn and starts receiving messages.
func (c *Client) UseTLSConn(conn *tls.Conn, debug bool) error {
	tlsc, err := newConn(conn, true, debug)
	if err != nil {
		return err
	}

	return c.useConn(tlsc)
}

// Send writes the given message to the connection immediately.
func (c *Client) Send(msg []byte) error {
	return newCtx(msg, c, c.middlewareOut).Next()
}

// SendAsync writes a message to the connection on a saparate gorotuine.
func (c *Client) SendAsync(msg []byte, callback func(error)) {
	c.msgWG.Add(1)
	go func() {
		defer c.msgWG.Done()
		if err := c.Send(msg); err != nil {
			// todo: better use an error handler middleware -or- both approaches?
			// todo2: use a single gorotuine + queue otherwise messages get interleaved
			callback(err)
		}
	}()
}

// Close closes the client connection.
func (c *Client) Close() error {
	return c.Conn.Close()
}

// UseConn creates a Client object wrapping an established Conn object.
func (c *Client) useConn(conn *Conn) error {
	if c.deadline != 0 {
		conn.deadline = c.deadline
	}

	id, err := shortid.UUID()
	if err != nil {
		return err
	}

	c.connID = id

	// append the last middleware to stack, which will write the response to connection, if any
	c.middlewareOut = append(c.middlewareOut, func(ctx *Ctx) error {
		if ctx.Msg != nil {
			return ctx.Client.Conn.Write(ctx.Msg)
		}

		return nil
	})

	c.Conn = conn
	c.msgWG.Add(1)
	go c.receive()
	return nil
}

// Receive reads from the connection until the connection is closed.
// If the connection is terminated unexpectedly, an error is logged.
// This method blocks and does not exit until connection is closed.
func (c *Client) receive() {
	defer c.Conn.Close()
	defer c.disconnHandler(c)
	defer c.msgWG.Done()

	for {
		msg, err := c.Conn.Read()
		if err != nil {
			// if the connected was closed by the other end
			if err == io.EOF {
				log.Printf("Peer disconnected. Conn ID: %v, Remote Addr: %v\n", c.connID, c.Conn.RemoteAddr())
				break
			}

			// if the connection was closed (possibly by us)
			if operr, ok := err.(*net.OpError); ok && operr.Op == "read" && operr.Err.Error() == "use of closed network connection" {
				log.Printf("Connection closed. Conn ID: %v, Remote Addr: %v\n", c.connID, c.Conn.RemoteAddr())
				break
			}

			log.Println("Unexpected error while reading from the connection:", err)
			break
		}

		c.msgWG.Add(1)
		go func() {
			defer c.msgWG.Done()
			ctx := newCtx(msg, c, c.middlewareIn)
			if err := ctx.Next(); err != nil {
				log.Println("Unhandled error in middleware stack:", err)
			}
		}()
	}
}
