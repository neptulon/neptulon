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
	reqMiddleware []func(ctx *ReqCtx)
	notMiddleware []func(ctx *NotCtx)
	resMiddleware []func(ctx *ResCtx)
}

// NewApp creates a Neptulon JSON-RPC app.
func NewApp(n *neptulon.App) (*App, error) {
	a := App{neptulon: n}
	n.Middleware(a.neptulonMiddleware)
	return &a, nil
}

// ReqMiddleware registers a new request middleware to handle incoming requests.
func (a *App) ReqMiddleware(reqMiddleware func(ctx *ReqCtx)) {
	a.reqMiddleware = append(a.reqMiddleware, reqMiddleware)
}

// NotMiddleware registers a new notification middleware to handle incoming notifications.
func (a *App) NotMiddleware(notMiddleware func(ctx *NotCtx)) {
	a.notMiddleware = append(a.notMiddleware, notMiddleware)
}

// ResMiddleware registers a new response middleware to handle incoming responses.
func (a *App) ResMiddleware(resMiddleware func(ctx *ResCtx)) {
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
		log.Fatalln("Errored while sending JSON-RPC message:", err)
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
			ctx := ReqCtx{Conn: conn, id: m.ID, method: m.Method, params: m.Params}
			for _, mid := range a.reqMiddleware {
				mid(&ctx)
				if ctx.Done || ctx.Res != nil || ctx.Err != nil {
					break
				}
			}

			if ctx.Res != nil || ctx.Err != nil {
				data, err := json.Marshal(Response{ID: m.ID, Result: ctx.Res, Error: ctx.Err})
				if err != nil {
					log.Fatalln("Errored while serializing JSON-RPC response:", err)
				}

				return data
			}

			return nil
		}

		// if incoming message is a response
		ctx := ResCtx{Conn: conn, id: m.ID, result: m.Result, code: m.Error.Code, message: m.Error.Message, data: m.Error.Data}
		for _, mid := range a.resMiddleware {
			mid(&ctx)
			if ctx.Done {
				break
			}
		}

		return nil
	}

	// if incoming message is a notification
	if m.Method != "" {
		ctx := NotCtx{Conn: conn, method: m.Method, params: m.Params}
		for _, mid := range a.notMiddleware {
			mid(&ctx)
			if ctx.Done {
				break
			}
		}

		return nil
	}

	// if incoming message is none of the above
	data, err := json.Marshal(Notification{Method: "invalidMessage"})
	if err != nil {
		log.Fatalln("Errored while serializing JSON-RPC response:", err)
	}

	return data
	// todo: close conn
}
