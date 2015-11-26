package neptulon

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/binary"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/neptulon/cmap"
)

// TLSConn is a full-duplex bidirectional client-server connection.
type TLSConn struct {
	id                 string
	session            *cmap.CMap
	conn               *tls.Conn
	headerSize         int
	maxMsgSize         int
	readDeadline       time.Duration
	debug              bool
	err                error
	clientDisconnected bool // hack: Whether the client disconnected from server before server closed connection
}

// NewTLSConn creates a new TLSConn object which wraps the given tls.Conn object.
// Default values for headerSize, maxMsgSize, and readDeadline are 4 bytes, 4294967295 bytes (4GB), and 300 seconds, respectively.
// Debug mode logs all raw TCP communication.
func NewTLSConn(conn *tls.Conn, headerSize, maxMsgSize, readDeadline int, debug bool) (*TLSConn, error) {
	if headerSize == 0 {
		headerSize = 4
	}
	if maxMsgSize == 0 {
		maxMsgSize = 4294967295
	}
	if readDeadline == 0 {
		readDeadline = 300
	}

	id, err := GenID()
	if err != nil {
		return nil, err
	}

	return &TLSConn{
		id:           id,
		session:      cmap.New(),
		conn:         conn,
		headerSize:   headerSize,
		maxMsgSize:   maxMsgSize,
		readDeadline: time.Second * time.Duration(readDeadline),
		debug:        debug,
	}, nil
}

// Dial creates a new client side connection to a server at the given network address,
// with optional CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func Dial(addr string, ca []byte, clientCert []byte, clientCertKey []byte, debug bool) (*TLSConn, error) {
	var cas *x509.CertPool
	var certs []tls.Certificate
	if ca != nil {
		cas = x509.NewCertPool()
		ok := cas.AppendCertsFromPEM(ca)
		if !ok {
			return nil, errors.New("failed to parse the CA certificate")
		}
	}
	if clientCert != nil {
		tlsCert, err := tls.X509KeyPair(clientCert, clientCertKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse the client certificate: %v", err)
		}

		c, _ := pem.Decode(clientCert)
		if tlsCert.Leaf, err = x509.ParseCertificate(c.Bytes); err != nil {
			return nil, fmt.Errorf("failed to parse the client certificate: %v", err)
		}

		certs = []tls.Certificate{tlsCert}
	}

	// todo: dial timeout like that of net.TLSConn.DialTimeout
	c, err := tls.Dial("tcp", addr, &tls.Config{RootCAs: cas, Certificates: certs})
	if err != nil {
		return nil, err
	}

	return NewTLSConn(c, 0, 0, 0, debug)
}

// ID is a randomly generated unique connection ID
func (c *TLSConn) ID() string {
	return c.id
}

// Session is a thread-safe data store for storing arbitrary data for this connection session.
func (c *TLSConn) Session() *cmap.CMap {
	return c.session
}

// SetReadDeadline set the read deadline for the connection in seconds.
func (c *TLSConn) SetReadDeadline(seconds int) {
	c.readDeadline = time.Second * time.Duration(seconds)
}

// Read waits for and reads the next incoming message from the TLS connection.
func (c *TLSConn) Read() (msg []byte, err error) {
	if err = c.conn.SetReadDeadline(time.Now().Add(c.readDeadline)); err != nil {
		return
	}

	// read the content length header
	h := make([]byte, c.headerSize)
	var n int
	n, err = c.conn.Read(h)
	if err != nil {
		return
	}
	if n != c.headerSize {
		err = fmt.Errorf("expected to read header size %v bytes but instead read %v bytes", c.headerSize, n)
		return
	}

	// calculate the content length
	n = readHeaderBytes(h)

	// read the message content
	msg = make([]byte, n)
	total := 0
	for total < n {
		// todo: log here in case it gets stuck, or there is a dos attack, pumping up cpu usage!
		i, err := c.conn.Read(msg[total:])
		if err != nil {
			err = fmt.Errorf("errored while reading incoming message: %v", err)
			break
		}
		total += i
	}
	if total != n {
		err = fmt.Errorf("expected to read %v bytes instead read %v bytes", n, total)
	}

	if c.debug {
		log.Println("Incoming message:", string(msg))
	}

	return
}

// Write writes given message to the connection.
func (c *TLSConn) Write(msg []byte) error {
	l := len(msg)
	h := makeHeaderBytes(l, c.headerSize)

	// write the header
	n, err := c.conn.Write(h)
	if err != nil {
		return err
	}
	if n != c.headerSize {
		err = fmt.Errorf("expected to write %v bytes but only wrote %v bytes", l, n)
	}

	// write the body
	n, err = c.conn.Write(msg)
	if err != nil {
		return err
	}
	if n != l {
		err = fmt.Errorf("expected to write %v bytes but only wrote %v bytes", l, n)
	}

	return nil
}

// RemoteAddr returns the remote network address.
func (c *TLSConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// ConnectionState returns basic TLS details about the connection.
func (c *TLSConn) ConnectionState() tls.ConnectionState {
	return c.conn.ConnectionState()
}

// Close closes a connection.
// Note: TCP/IP stack does not guarantee delivery of messages before the connection is closed.
func (c *TLSConn) Close() error {
	return c.conn.Close() // todo: if conn.err is nil, send a close req and wait ack then close? (or even wait for everything else to finish?)
}

func makeHeaderBytes(h, size int) []byte {
	b := make([]byte, size)
	binary.LittleEndian.PutUint32(b, uint32(h))
	return b
}

func readHeaderBytes(h []byte) int {
	return int(binary.LittleEndian.Uint32(h))
}
