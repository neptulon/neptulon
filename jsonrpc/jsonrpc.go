// Package jsonrpc implements JSON-RPC 2.0 protocol for Neptulon framework.
package jsonrpc

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/nbusy/neptulon"
)

// Server is a Neptulon JSON-RPC server.
type Server struct {
	neptulon      *neptulon.Server
	reqMiddleware []func(ctx *ReqCtx)
	notMiddleware []func(ctx *NotCtx)
	resMiddleware []func(ctx *ResCtx)
}

// NewServer creates a Neptulon JSON-RPC server.
func NewServer(n *neptulon.Server) (*Server, error) {
	if n == nil {
		return nil, errors.New("Given Neptulon server instance is nil.")
	}

	s := Server{neptulon: n}
	n.Middleware(s.neptulonMiddleware)
	return &s, nil
}

// ReqMiddleware registers a new request middleware to handle incoming requests.
func (s *Server) ReqMiddleware(reqMiddleware func(ctx *ReqCtx)) {
	s.reqMiddleware = append(s.reqMiddleware, reqMiddleware)
}

// NotMiddleware registers a new notification middleware to handle incoming notifications.
func (s *Server) NotMiddleware(notMiddleware func(ctx *NotCtx)) {
	s.notMiddleware = append(s.notMiddleware, notMiddleware)
}

// ResMiddleware registers a new response middleware to handle incoming responses.
func (s *Server) ResMiddleware(resMiddleware func(ctx *ResCtx)) {
	s.resMiddleware = append(s.resMiddleware, resMiddleware)
}

// Send sends a message throught the connection denoted by the connection ID.
func (s *Server) Send(connID string, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Fatalln("Errored while serializing JSON-RPC response:", err)
	}

	err = s.neptulon.Send(connID, data)
	if err != nil {
		log.Fatalln("Errored while sending JSON-RPC message:", err)
	}
}

func (s *Server) neptulonMiddleware(conn *neptulon.Conn, msg []byte) []byte {
	var m message
	if err := json.Unmarshal(msg, &m); err != nil {
		log.Fatalln("Cannot deserialize incoming message:", err)
	}

	// if incoming message is a request or response
	if m.ID != "" {
		// if incoming message is a request
		if m.Method != "" {
			ctx := ReqCtx{Conn: conn, id: m.ID, method: m.Method, params: m.Params}
			for _, mid := range s.reqMiddleware {
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
		for _, mid := range s.resMiddleware {
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
		for _, mid := range s.notMiddleware {
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
