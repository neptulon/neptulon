package test

import (
	"io"
	"net"
	"sync"
	"testing"
	"time"
)

// ListenerHelper is a net.Listener wrapper for testing.
// All the functions are wrapped with proper test runner error logging.
type ListenerHelper struct {
	Listener net.Listener
	Addr     string

	testing    *testing.T
	listenerWG sync.WaitGroup
}

// NewListenerHelper creates a new listener helper object.
func NewListenerHelper(t *testing.T) *ListenerHelper {
	if testing.Short() {
		t.Skip("Skipping integration test in short testing mode")
	}

	lh := &ListenerHelper{
		Addr:    "127.0.0.1:3001",
		testing: t,
	}

	l, err := net.Listen("tcp", lh.Addr)
	if err != nil {
		t.Fatalf("listener: failed to start on address %v: %v", lh.Addr, err)
	}

	lh.listenerWG.Add(1)
	go func() {
		defer lh.listenerWG.Done()
		t.Log("listener: started accepting connection on:", lh.Addr)

		conn, err := l.Accept()
		if err != nil {
			t.Fatal("listener: failed to accept connection:", err)
		}

		if _, err := io.Copy(conn, conn); err != nil {
			t.Fatal("listener: failed to read or write message from connection:", err)
		}

		if err := conn.Close(); err != nil {
			t.Fatal("listener: failed to close connection:", err)
		}

		if err := l.Close(); err != nil {
			t.Fatal("listener: failed to close listener:", err)
		}
	}()

	time.Sleep(time.Millisecond) // give Accept() goroutine cycles to initiate

	lh.Listener = l
	return lh
}

// GetClientPair creates and returns a new client connection.
func (lh *ListenerHelper) GetClientPair() *ClientHelper {
	return NewClientHelper(lh.testing, lh.Addr).Connect()
}

// func TestRead(t *testing.T) {
// 	msg1 := "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
// 	msg2 := "In sit amet lectus felis, at pellentesque turpis."
// 	msg3 := "Nunc urna enim, cursus varius aliquet ac, imperdiet eget tellus."
// 	msg4 := randstr.Get(45000)   // 0.45 MB
// 	msg5 := randstr.Get(5000000) // 5.0 MB
//
// 	host := "127.0.0.1:3010"
// 	certChain, err := ca.GenCertChain("FooBar", "127.0.0.1", "127.0.0.1", time.Hour, 512)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	l, err := ListenTLS(certChain.ServerCert, certChain.ServerKey, certChain.IntCACert, host, false)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	var listenerWG sync.WaitGroup
// 	listenerWG.Add(1)
// 	go func() {
// 		defer listenerWG.Done()
// 		l.Accept(func(conn *Conn) {},
// 			func(conn *Conn, msg []byte) {
// 				m := string(msg)
// 				if m == "close" {
// 					conn.Close()
// 					return
// 				}
//
// 				connstate, _ := conn.ConnectionState()
// 				certs := connstate.PeerCertificates
// 				if len(certs) > 0 {
// 					t.Logf("Client connected with client certificate subject: %v\n", certs[0].Subject)
// 				}
//
// 				if m != msg1 && m != msg2 && m != msg3 && m != msg4 && m != msg5 {
// 					t.Fatal("Sent and incoming messages did not match! Sent message was message:", m)
// 				}
// 			}, func(conn *Conn) {})
// 	}()
//
// 	roots := x509.NewCertPool()
// 	ok := roots.AppendCertsFromPEM(certChain.IntCACert)
// 	if !ok {
// 		panic("failed to parse root certificate")
// 	}
//
// 	tlsConf := &tls.Config{RootCAs: roots}
// 	conn, err := tls.Dial("tcp", host, tlsConf)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	newconn, _ := NewTLSConn(conn, 0, 0, 0, false)
//
// 	send(t, newconn, msg1)
// 	send(t, newconn, msg1)
// 	send(t, newconn, msg2)
// 	send(t, newconn, msg3)
// 	send(t, newconn, msg4)
// 	send(t, newconn, msg1)
// 	send(t, newconn, msg5)
// 	send(t, newconn, msg1)
// 	send(t, newconn, "close")
//
// 	l.reqWG.Wait()
//
// 	l.connWG.Wait()
// 	newconn.Close()
//
// 	l.Close()
// 	listenerWG.Wait()
//
// 	// t.Logf("\nconn:\n%+v\n\n", conn)
// 	// t.Logf("\nconn.ConnectionState():\n%+v\n\n", conn.ConnectionState())
// 	// t.Logf("\ntls.Config:\n%+v\n\n", tlsConf)
// }

// func TestClientDisconnect(t *testing.T) {
// 	// todo: we need to verify that events occur in the order that we want them (either via event hooks or log analysis)
// 	// this seems like a listener test than a integration test from a client perspective
// 	s := getServer(t)
// 	c := getClientConnWithClientCert(t)
// 	if err := c.Close(); err != nil {
// 		t.Fatal("Failed to close the client connection:", err)
// 	}
// 	if err := s.Stop(); err != nil {
// 		t.Fatal("Failed to stop the server:", err)
// 	}
// 	wg.Wait()
// }

// // closeGraceful waits for all request then connection handler goroutines to return then closes the listener. This method is meant for testing.
// func closeGraceful(l *Listener) error {
// 	// todo: more proper way is to do TCPConn.CloseRead()/reqWG.Wait()/TCPConn.CloseWrite()/listener.Close()
// 	// but that requires using net.TCPListener/TCPConn and then upgrading to TLS (which is also good when supporting UnixSocket)
// 	l.reqWG.Wait()
// 	l.connWG.Wait()
// 	return l.listener.Close()
// }
