package jsonrpc

import (
	"errors"

	"github.com/neptulon/cmap"
	"github.com/neptulon/neptulon"
)

// Router is a JSON-RPC message routing middleware.
type Router struct {
	jsonrpc   *Server
	reqRoutes map[string]func(ctx *ReqCtx) // method name -> handler func(ctx *ReqCtx)
	notRoutes map[string]func(ctx *NotCtx) // method name -> handler func(ctx *NotCtx)
	resRoutes *cmap.CMap                   // message ID (string) -> handler func(ctx *ResCtx) : requests sent from the router that are pending responses from clients
}

// NewRouter creates a JSON-RPC router instance and registers it with the Neptulon JSON-RPC server.
func NewRouter(s *Server) (*Router, error) {
	if s == nil {
		return nil, errors.New("Given Neptulon server instance is nil.")
	}

	r := Router{
		jsonrpc:   s,
		reqRoutes: make(map[string]func(ctx *ReqCtx)),
		notRoutes: make(map[string]func(ctx *NotCtx)),
		resRoutes: cmap.New(),
	}

	s.ReqMiddleware(r.reqMiddleware)
	s.NotMiddleware(r.notMiddleware)
	s.ResMiddleware(r.resMiddleware)
	return &r, nil
}

// Request adds a new incoming request route registry.
func (r *Router) Request(route string, handler func(ctx *ReqCtx)) {
	r.reqRoutes[route] = handler
}

// Notification adds a new incoming notification route registry.
func (r *Router) Notification(route string, handler func(ctx *NotCtx)) {
	r.notRoutes[route] = handler
}

// SendRequest sends a JSON-RPC request throught the connection denoted by the connection ID.
// resHandler is called when a response is returned.
func (r *Router) SendRequest(connID string, method string, params interface{}, resHandler func(ctx *ResCtx)) error {
	id, err := neptulon.GenID()
	if err != nil {
		return err
	}

	req := Request{ID: id, Method: method, Params: params}
	if err = r.jsonrpc.send(connID, req); err != nil {
		return err
	}

	r.resRoutes.Set(req.ID, resHandler)
	return nil
}

// SendNotification sends a JSON-RPC notification through the connection denoted by the connection ID.
func (r *Router) SendNotification(connID string, method string, params interface{}) error {
	return r.jsonrpc.send(connID, Notification{Method: method, Params: params})
}

func (r *Router) reqMiddleware(ctx *ReqCtx) {
	if handler, ok := r.reqRoutes[ctx.method]; ok {
		handler(ctx)
	}
}

func (r *Router) notMiddleware(ctx *NotCtx) {
	if handler, ok := r.notRoutes[ctx.method]; ok {
		handler(ctx)
	}
}

func (r *Router) resMiddleware(ctx *ResCtx) {
	if handler, ok := r.resRoutes.GetOk(ctx.id); ok {
		handler.(func(ctx *ResCtx))(ctx)
		r.resRoutes.Delete(ctx.id)
	}
}
