// Package jsonrpc implements JSON-RPC 2.0 protocol for Neptulon framework.
package jsonrpc

import (
	"encoding/json"
	"log"

	"github.com/nbusy/neptulon"
)

// App is a Neptulon JSON-RPC app.
type App struct {
	neptulon      *neptulon.App
	reqMiddleware []func(ctx *ReqContext)
	notMiddleware []func(ctx *NotContext)
	resMiddleware []func(ctx *ResContext)
}

// NewApp creates a Neptulon JSON-RPC app.
func NewApp(n *neptulon.App) (*App, error) {
	a := App{neptulon: n}
	n.Middleware(a.neptulonMiddleware)
	return &a, nil
}

// ReqMiddleware registers a new request middleware to handle incoming requests.
func (a *App) ReqMiddleware(reqMiddleware func(ctx *ReqContext)) {
	a.reqMiddleware = append(a.reqMiddleware, reqMiddleware)
}

// NotMiddleware registers a new notification middleware to handle incoming notifications.
func (a *App) NotMiddleware(notMiddleware func(ctx *NotContext)) {
	a.notMiddleware = append(a.notMiddleware, notMiddleware)
}

// ResMiddleware registers a new response middleware to handle incoming responses.
func (a *App) ResMiddleware(resMiddleware func(ctx *ResContext)) {
	a.resMiddleware = append(a.resMiddleware, resMiddleware)
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
	var m message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Fatalln("Cannot deserialize incoming message:", err)
	}

	// if incoming message is a request or response
	if m.ID != "" {
		// if incoming message is a request
		if m.Method != "" {
			ctx := ReqContext{Conn: conn, Req: &Request{ID: m.ID, Method: m.Method, Params: m.Params}}
			for _, mid := range a.reqMiddleware {
				mid(&ctx)
				if ctx.Res == nil && ctx.ResErr == nil {
					continue
				}

				data, err := json.Marshal(Response{ID: m.ID, Result: ctx.Res, Error: ctx.ResErr})
				if err != nil {
					log.Fatalln("Errored while serializing JSON-RPC response:", err)
				}

				return data
			}
		}

		// if incoming message is a response
		ctx := ResContext{Conn: conn, Res: &Response{ID: m.ID, Result: m.Result, Error: m.Error}}
		for _, mid := range a.resMiddleware {
			mid(&ctx)
		}
	}

	// if incoming message is a notification
	if m.Method != "" {
		ctx := NotContext{Conn: conn, Not: &Notification{Method: m.Method, Params: m.Params}}
		for _, mid := range a.notMiddleware {
			mid(&ctx)
		}
	}

	// if incoming message is none of the above
	data, err := json.Marshal(Notification{Method: "system.invalid.message"})
	if err != nil {
		log.Fatalln("Errored while serializing JSON-RPC response:", err)
	}

	return data
	// todo: close conn
}
