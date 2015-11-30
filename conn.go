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
	"github.com/neptulon/shortid"
)

// Conn is a full-duplex bidirectional connection.
type Conn struct {
	ID      string     // Randomly generated unique connection ID.
	Session *cmap.CMap // Thread-safe data store for storing arbitrary data for this connection session.
	Conn    net.Conn   // Inner connection object.

	headerSize         int
	maxMsgSize         int
	readDeadline       time.Duration
	debug              bool
	err                error
	clientDisconnected bool // hack: Whether the client disconnected from server before server closed connection
}

// newTLSConn creates a new Conn object which wraps the given tls.Conn object.
// Default values for headerSize, maxMsgSize, and readDeadline are 4 bytes, 4294967295 bytes (4GB), and 300 seconds, respectively.
// Debug mode logs all raw TCP communication.
func newTLSConn(conn *tls.Conn, headerSize, maxMsgSize, readDeadline int, debug bool) (*Conn, error) {
	if headerSize == 0 {
		headerSize = 4
	}
	if maxMsgSize == 0 {
		maxMsgSize = 4294967295
	}
	if readDeadline == 0 {
		readDeadline = 300
	}

	id, err := shortid.UUID()
	if err != nil {
		return nil, err
	}

	return &Conn{
		ID:           id,
		Session:      cmap.New(),
		Conn:         conn,
		headerSize:   headerSize,
		maxMsgSize:   maxMsgSize,
		readDeadline: time.Second * time.Duration(readDeadline),
		debug:        debug,
	}, nil
}

// DialTLS creates a new client side connection to a server at the given network address,
// with optional CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func DialTLS(addr string, ca []byte, clientCert []byte, clientCertKey []byte, debug bool) (*Conn, error) {
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

	// todo: dial timeout like that of net.Conn.DialTimeout
	c, err := tls.Dial("tcp", addr, &tls.Config{RootCAs: cas, Certificates: certs})
	if err != nil {
		return nil, err
	}

	return newTLSConn(c, 0, 0, 0, debug)
}

// SetReadDeadline set the read deadline for the connection in seconds.
func (c *Conn) SetReadDeadline(seconds int) {
	c.readDeadline = time.Second * time.Duration(seconds)
}

// Read waits for and reads the next incoming message from the connection.
func (c *Conn) Read() (msg []byte, err error) {
	if err = c.Conn.SetReadDeadline(time.Now().Add(c.readDeadline)); err != nil {
		return
	}

	// read the content length header
	h := make([]byte, c.headerSize)
	var n int
	n, err = c.Conn.Read(h)
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
		i, err := c.Conn.Read(msg[total:])
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
func (c *Conn) Write(msg []byte) error {
	l := len(msg)
	h := makeHeaderBytes(l, c.headerSize)

	// write the header
	n, err := c.Conn.Write(h)
	if err != nil {
		return err
	}
	if n != c.headerSize {
		err = fmt.Errorf("expected to write %v bytes but only wrote %v bytes", l, n)
	}

	// write the body
	n, err = c.Conn.Write(msg)
	if err != nil {
		return err
	}
	if n != l {
		err = fmt.Errorf("expected to write %v bytes but only wrote %v bytes", l, n)
	}

	return nil
}

// RemoteAddr returns the remote network address.
func (c *Conn) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

// ConnectionState returns basic TLS details about the connection.
func (c *Conn) ConnectionState() tls.ConnectionState {
	return c.Conn.(*tls.Conn).ConnectionState()
}

// Close closes a connection.
// Note: TCP/IP stack does not guarantee delivery of messages before the connection is closed.
func (c *Conn) Close() error {
	return c.Conn.Close() // todo: if conn.err is nil, send a close req and wait ack then close? (or even wait for everything else to finish?)
}

func makeHeaderBytes(h, size int) []byte {
	b := make([]byte, size)
	binary.LittleEndian.PutUint32(b, uint32(h))
	return b
}

func readHeaderBytes(h []byte) int {
	return int(binary.LittleEndian.Uint32(h))
}
