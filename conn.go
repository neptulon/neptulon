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
)

// Conn is a full-duplex bidirectional connection.
// Default values for header size, maximum message size, and read/write deadlines are
// 4 bytes, 2^([header size]4*8)-1 = 4294967295 bytes (4GB), and 300 seconds, respectively.
type Conn struct {
	Conn net.Conn // Inner connection object.

	tls        bool
	headerSize int
	maxMsgSize int
	deadline   time.Duration
	debug      bool
}

// SetDeadline set the read/write deadlines for the connection, in seconds.
func (c *Conn) SetDeadline(seconds int) {
	c.deadline = time.Second * time.Duration(seconds)
}

// Read waits for and reads the next incoming message from the connection.
func (c *Conn) Read() ([]byte, error) {
	if err := c.Conn.SetReadDeadline(time.Now().Add(c.deadline)); err != nil {
		return nil, err
	}

	// read the content length header
	h := make([]byte, c.headerSize)
	n, err := c.Conn.Read(h)
	if err != nil {
		return nil, err
	}

	if n != c.headerSize {
		return nil, fmt.Errorf("expected to read header size %v bytes but instead read %v bytes", c.headerSize, n)
	}

	// calculate the content length
	n = readHeaderBytes(h)
	if n > c.maxMsgSize {
		return nil, fmt.Errorf("size of message to be read (%v) is bigger than maxMsgSize (%v)", n, c.maxMsgSize)
	}

	// read the message content
	msg := make([]byte, n)
	total := 0
	for total < n {
		// todo: log here in case it gets stuck, or there is a dos attack, pumping up cpu usage!
		i, err := c.Conn.Read(msg[total:])
		if err != nil {
			err = fmt.Errorf("error while reading incoming message: %v", err)
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

	return msg, nil
}

// Write writes given message to the connection.
func (c *Conn) Write(msg []byte) error {
	if err := c.Conn.SetWriteDeadline(time.Now().Add(c.deadline)); err != nil {
		return err
	}

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
func (c *Conn) ConnectionState() (tls.ConnectionState, error) {
	if !c.tls {
		return tls.ConnectionState{}, errors.New("not a TLS connection")
	}

	return c.Conn.(*tls.Conn).ConnectionState(), nil
}

// Close closes a connection.
// Note: TCP/IP stack does not guarantee delivery of messages before the connection is closed.
func (c *Conn) Close() error {
	return c.Conn.Close() // todo: if conn.err is nil, send a close req and wait ack then close? (or even wait for everything else to finish?)
}

// DialTCP creates a new client side TCP connection to a server at the given network address.
func dialTCP(addr string, debug bool) (*Conn, error) {
	c, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newConn(c, false, debug)
}

// DialTLS creates a new client side TLS connection to a server at the given network address,
// with optional server CA and/or a client certificate (PEM encoded X.509 cert/key).
func dialTLS(addr string, ca, clientCert, clientCertKey []byte, debug bool) (*Conn, error) {
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

	c, err := tls.Dial("tcp", addr, &tls.Config{RootCAs: cas, Certificates: certs})
	if err != nil {
		return nil, err
	}

	return newConn(c, true, debug)
}

func newConn(conn net.Conn, tls, debug bool) (*Conn, error) {
	if conn == nil {
		return nil, errors.New("connection object cannot be nil")
	}

	return &Conn{
		Conn:       conn,
		tls:        tls,
		headerSize: 4,
		maxMsgSize: 4294967295,
		deadline:   time.Second * time.Duration(300),
		debug:      debug,
	}, nil
}

// makeHeaderBytes takes the size of a message in bytes and puts it into a header block in little endian encoding.
// i.e. message size 4294967295 bytes and 4 byte header block will generate header: [255 255 255 255]
// l = message size in bytes
// h = header size in bytes
func makeHeaderBytes(m, h int) []byte {
	b := make([]byte, h)
	binary.LittleEndian.PutUint32(b, uint32(m))
	return b
}

// readHeaderBytes does reverse of what makeHeaderBytes does and reads the message size out of the given header block.
func readHeaderBytes(h []byte) int {
	return int(binary.LittleEndian.Uint32(h))
}
