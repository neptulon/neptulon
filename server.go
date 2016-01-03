// Package neptulon is a RPC framework with middleware support.
package neptulon

import (
	"net/http"

	"golang.org/x/net/websocket"
)

// Server is a Neptulon server.
type Server struct {
}

// NewServer creates a new Neptulon server.
func NewServer(addr string) *Server {
	return &Server{}
}

// This example demonstrates a trivial echo server.
func Start() {
	http.Handle("/", websocket.Handler(EchoServer))
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

type T struct {
	Msg   string
	Count int
}

// Echo the data received on the WebSocket.
func EchoServer(ws *websocket.Conn) {
	// receive JSON type T
	var data T
	websocket.JSON.Receive(ws, &data)

	// send JSON type T
	websocket.JSON.Send(ws, data)
}
