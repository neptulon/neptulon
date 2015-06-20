// Package jsonrpc implements JSON-RPC 2.0 protocol for Neptulon framework.
package jsonrpc

import (
	"encoding/json"
	"log"

	"github.com/nbusy/neptulon"
)

// App is a Neptulon JSON-RPC app.
type App struct {
	neptulon   *neptulon.App
	middleware []func(ctx *Context)
}

// NewApp creates a Neptulon JSON-RPC app.
func NewApp(n *neptulon.App) (*App, error) {
	a := App{neptulon: n}
	n.Middleware(a.neptulonMiddleware)
	return &a, nil
}

// Middleware registers a new middleware to handle incoming messages.
func (a *App) Middleware(middleware func(ctx *Context)) {
	a.middleware = append(a.middleware, middleware)
}

// Send sends a message throught the connection denoted by the connection ID.
func (a *App) Send(connID string, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("Errored while serializing JSON-RPC response:", err)
	}

	err = a.neptulon.Send(connID, data)
	if err != nil {
		log.Fatalln("Errored sending JSON-RPC message:", err)
	}
}

func (a *App) neptulonMiddleware(conn *neptulon.Conn, msg []byte) []byte {
	var m Message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Fatalln("Cannot deserialize incoming message:", err)
	}

	for _, mid := range a.middleware {
		ctx := Context{Conn: conn, Msg: &m}
		mid(&ctx)
		if ctx.ResMsg.Result == nil && ctx.ResMsg.Error == nil {
			continue
		}

		if m.Method == "" || m.ID == "" {
			log.Fatalln("Cannot return a response to a non request")
		}

		data, err := json.Marshal(Response{ID: m.ID, Result: ctx.ResMsg.Result, Error: ctx.ResMsg.Error})
		if err != nil {
			log.Fatalln("Errored while serializing JSON-RPC response:", err)
		}

		return data
	}

	return nil
}
