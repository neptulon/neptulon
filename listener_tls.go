package neptulon

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

// TLSListener accepts connections from devices.
type TLSListener struct {
	debug        bool
	listener     net.Listener
	readDeadline int
	connWG       sync.WaitGroup
	reqWG        sync.WaitGroup
}

// Listen creates a TLS listener with the given PEM encoded X.509 certificate and the private key on the local network address laddr.
// Debug mode logs all server activity.
func Listen(cert, privKey, clientCACert []byte, laddr string, debug bool) (*TLSListener, error) {
	tlsCert, err := tls.X509KeyPair(cert, privKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the server certificate or the private key: %v", err)
	}

	c, _ := pem.Decode(cert)
	if tlsCert.Leaf, err = x509.ParseCertificate(c.Bytes); err != nil {
		return nil, fmt.Errorf("failed to parse the server certificate: %v", err)
	}

	pool := x509.NewCertPool()
	ok := pool.AppendCertsFromPEM(clientCACert)
	if !ok {
		return nil, fmt.Errorf("failed to parse the CA certificate: %v", err)
	}

	conf := tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		ClientCAs:    pool,
		ClientAuth:   tls.VerifyClientCertIfGiven,
	}

	l, err := tls.Listen("tcp", laddr, &conf) // todo: returned type is TCPListener due to "tcp" so store it as such
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS listener on network address %v with error: %v", laddr, err)
	}
	log.Printf("TLS listener created: %v\n", laddr)

	return &TLSListener{
		debug:    debug,
		listener: l,
	}, nil
}

// SetReadDeadline sets the read deadline for connections.
// If not set, default deadline of Conn struct is used.
func (l *TLSListener) SetReadDeadline(seconds int) {
	l.readDeadline = seconds
}

// Accept waits for incoming connections and forwards the client connect/message/disconnect events to provided handlers in a new goroutine.
// This function blocks and never returns, unless there is an error while accepting a new connection.
func (l *TLSListener) Accept(handleConn func(conn Conn), handleMsg func(conn Conn, msg []byte), handleDisconn func(conn Conn)) error {
	defer log.Println("Listener closed:", l.listener.Addr())
	for {
		conn, err := l.listener.Accept()
		if err != nil {
			if operr, ok := err.(*net.OpError); ok && operr.Op == "accept" && operr.Err.Error() == "use of closed network connection" {
				return nil
			}
			return fmt.Errorf("error while accepting a new connection from a client: %v", err)
			// todo: it might not be appropriate to break the loop on recoverable errors (like client disconnect during handshake)
			// the underlying fd.accept() does some basic recovery though we might need more: http://golang.org/src/net/fd_unix.go
		}

		// todo: this casting early on doesn't seem necessary and can be removed in a futur iteration not to cause any side effects
		tlsconn, ok := conn.(*tls.Conn)
		if !ok {
			conn.Close()
			return errors.New("cannot cast net.Conn interface to tls.Conn type")
		}

		l.connWG.Add(1)
		log.Println("Client connected:", conn.RemoteAddr())

		c, err := newTLSConn(tlsconn, 0, 0, l.readDeadline, l.debug)
		if err != nil {
			return err
		}

		// client, err := newTLSClient

		go handleClient(l, c, handleConn, handleMsg, handleDisconn)
	}
}

// handleClient waits for messages from the connected client and forwards the client message/disconnect
// events to provided handlers in a new goroutine.
// This function never returns, unless there is an error while reading from the channel or the client disconnects.
func handleClient(l *TLSListener, conn Conn, handleConn func(conn Conn), handleMsg func(conn Conn, msg []byte), handleDisconn func(conn Conn)) error {
	handleConn(conn)
	tlsConn := conn.(*TLSConn) // hack: till we clean all the below hacks

	defer func() {
		tlsConn.err = tlsConn.Close() // todo: handle close error, store the error in conn object and return it to handleMsg/handleErr/handleDisconn or one level up (to server)
		if tlsConn.clientDisconnected {
			log.Println("Client disconnected:", tlsConn.RemoteAddr())
		} else {
			log.Println("Closed client connection:", tlsConn.RemoteAddr())
		}
		handleDisconn(tlsConn)
		l.connWG.Done()
	}()

	for {
		if tlsConn.err != nil {
			return tlsConn.err // todo: should we send error message to user, log the error, and close the conn and return instead?
		}

		msg, err := tlsConn.Read()
		if err != nil {
			if err == io.EOF {
				tlsConn.clientDisconnected = true
				break
			}
			if operr, ok := err.(*net.OpError); ok && operr.Op == "read" && operr.Err.Error() == "use of closed network connection" {
				tlsConn.clientDisconnected = true
				break
			}
			log.Fatalln("Errored while reading:", err)
		}

		l.reqWG.Add(1)
		go func() {
			defer l.reqWG.Done()
			handleMsg(conn, msg)
		}()
	}

	return tlsConn.err
}

// Close closes the listener.
func (l *TLSListener) Close() error {
	return l.listener.Close()
}