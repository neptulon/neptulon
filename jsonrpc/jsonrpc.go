// Package jsonrpc implements JSON-RPC 2.0 protocol for Neptulon framework.
package jsonrpc

import (
	"encoding/json"
	"log"

	"github.com/nbusy/neptulon"
)

// App is a Neptulon JSON-RPC app.
type App struct {
	inMiddleware []func(conn *neptulon.Conn, msg *Message) (result interface{}, resErr *ResError)
	// outMiddleware []func(conn *neptulon.Conn, msg *Message)
}

// NewApp creates a Neptulon JSON-RPC app.
func NewApp(n *neptulon.App) (*App, error) {
	a := App{}
	n.Middleware(a.neptulonMiddleware)
	return &a, nil
}

// Middleware registers a new middleware to handle incoming messages.
func (a *App) Middleware(middleware func(conn *neptulon.Conn, msg *Message) (result interface{}, resErr *ResError)) {
	a.inMiddleware = append(a.inMiddleware, middleware)
}

func (a *App) neptulonMiddleware(conn *neptulon.Conn, msg []byte) []byte {
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Fatalln("Cannot deserialize incoming message:", err)
	}

	for _, mid := range a.inMiddleware {
		res, resErr := mid(conn, &m)
		if res == nil && resErr == nil {
			continue
		}

		if m.Method == "" || m.ID == "" {
			log.Fatalln("Cannot return a response to a non request")
		}

		data, err := json.Marshal(Response{ID: m.ID, Result: res, Error: resErr})
		if err != nil {
			log.Fatalln("Errored while serializing JSON-RPC response:", err)
		}

		return data
	}

	return nil
}
