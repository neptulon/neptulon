package client

import (
	"crypto/tls"
	"reflect"
	"testing"
)

func TestMakeHeaderBytes(t *testing.T) {
	if h := makeHeaderBytes(1, 4); !reflect.DeepEqual(h, []byte{1, 0, 0, 0}) {
		t.Fatal("expected 1000 got", h)
	}

	if h := makeHeaderBytes(858993459, 4); !reflect.DeepEqual(h, []byte{51, 51, 51, 51}) {
		t.Fatal("expected 51515151 got", h)
	}

	if h := makeHeaderBytes(4294967295, 4); !reflect.DeepEqual(h, []byte{255, 255, 255, 255}) {
		t.Fatal("expected 255255255255 got", h)
	}
}

func TestReadHeaderBytes(t *testing.T) {
	if m := readHeaderBytes([]byte{1, 0, 0, 0}); m != 1 {
		t.Fatal("expected 1 got", m)
	}

	if m := readHeaderBytes([]byte{51, 51, 51, 51}); m != 858993459 {
		t.Fatal("expected 858993459 got", m)
	}
}

func TestNewConn(t *testing.T) {
	conn := &tls.Conn{}
	c, err := newConn(conn, true, true)
	if err != nil {
		t.Fatal(err)
	}

	if c.Conn != conn ||
		!c.tls ||
		!c.debug ||
		c.headerSize == 0 ||
		c.maxMsgSize == 0 ||
		c.readDeadline == 0 {
		t.Fatal("Conn object was misconfigured")
	}
}

func TestReadWrite(t *testing.T) {
	// addr := "127.0.0.1:3001"
	// var listenerWG sync.WaitGroup
	//
	// l, err := net.Listen("tcp", addr)
	// if err != nil {
	// 	t.Fatalf("listener: failed to start on address %v: %v", addr, err)
	// }
	//
	// listenerWG.Add(1)
	// go func() {
	// 	defer listenerWG.Done()
	// 	t.Log("listener: started accepting connection on:", addr)
	//
	// 	conn, err := l.Accept()
	// 	if err != nil {
	// 		t.Fatal("listener: failed to accept connection:", err)
	// 	}
	//
	// 	if _, err := io.Copy(conn, conn); err != nil {
	// 		t.Fatal("listener: failed to read or write message from connection:", err)
	// 	}
	//
	// 	if err := conn.Close(); err != nil {
	// 		t.Fatal("listener: failed to close connection:", err)
	// 	}
	//
	// 	if err := l.Close(); err != nil {
	// 		t.Fatal("listener: failed to close listener:", err)
	// 	}
	// }()
	//
	// time.Sleep(time.Millisecond) // give Accept() goroutine cycles to initiate

	// lh := test.NewListenerHelper(t)
	// c, err := dialTCP(lh.Addr)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

func TestSimultaneousWrite(t *testing.T) {
	// t.Fatalln("Failed to receive simultaneously written and TCP interleaved messages in the correct order:", err)
	// if this is the case, we might need to merge header+payload before sending instead of sending them in two step order
}
