package jsonrpc

import "errors"

// Router is a JSON-RPC message routing middleware.
type Router struct {
	jsonrpc   *Server
	reqRoutes map[string]func(ctx *ReqCtx) // method name -> handler
	notRoutes map[string]func(ctx *NotCtx) // method name -> handler
	resRoutes map[string]func(ctx *ResCtx) // message ID -> handler : requests sent from the router that are pending responses from clients
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
		resRoutes: make(map[string]func(ctx *ResCtx)),
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
func (r *Router) SendRequest(connID string, req *Request, resHandler func(ctx *ResCtx)) {
	r.jsonrpc.Send(connID, req)
	r.resRoutes[req.ID] = resHandler
}

// SendNotification sends a JSON-RPC notification through the connection denoted by the connection ID.
func (r *Router) SendNotification(connID string, not *Notification) {
	r.jsonrpc.Send(connID, not)
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
	if handler, ok := r.resRoutes[ctx.id]; ok {
		handler(ctx)
		delete(r.resRoutes, ctx.id)
	}
}
