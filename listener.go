package neptulon

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
)

// listener accepts connections from devices.
type listener struct {
	debug    bool
	listener net.Listener
}

// listenTLS creates a TLS listener with the given PEM encoded X.509 certificate and the private key on the local network address laddr.
// Debug mode logs all server activity.
func listenTLS(cert, privKey, clientCACert []byte, laddr string, debug bool) (*listener, error) {
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

	l, err := tls.Listen("tcp", laddr, &conf)
	if err != nil {
		return nil, fmt.Errorf("failed to create TLS listener on network address %v with error: %v", laddr, err)
	}

	log.Printf("TLS listener created: %v\n", laddr)

	return &listener{
		debug:    debug,
		listener: l,
	}, nil
}

// Accept waits for incoming connections and forwards the client connect events to provided handler.
// This function blocks and never returns, unless there is an error while accepting a new connection.
func (l *listener) Accept(connHandler func(c net.Conn) error) error {
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

		log.Println("Client connected:", conn.RemoteAddr())

		return connHandler(conn)
	}
}

// Close closes the listener.
func (l *listener) Close() error {
	return l.listener.Close()
}
