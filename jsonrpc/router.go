package jsonrpc

// Router is a JSON-RPC message routing middleware.
type Router struct {
	jsonrpc        *App
	reqRoutes      map[string]func(ctx *ReqContext)
	notRoutes      map[string]func(ctx *NotContext)
	pendinRequests map[string]chan *Response // requests sent from the router that are pending responses from clients
}

// NewRouter creates a JSON-RPC router instance and registers it with the Neptulon JSON-RPC app.
func NewRouter(app *App) (*Router, error) {
	r := Router{
		jsonrpc:        app,
		reqRoutes:      make(map[string]func(ctx *ReqContext)),
		notRoutes:      make(map[string]func(ctx *NotContext)),
		pendinRequests: make(map[string]chan *Response),
	}

	app.ReqMiddleware(r.reqMiddleware)
	app.NotMiddleware(r.notMiddleware)
	app.ResMiddleware(r.resMiddleware)
	return &r, nil
}

// Request adds a new incoming request route registry.
func (r *Router) Request(route string, handler func(ctx *ReqContext)) {
	r.reqRoutes[route] = handler
}

// Notification adds a new incoming notification route registry.
func (r *Router) Notification(route string, handler func(ctx *NotContext)) {
	r.notRoutes[route] = handler
}

// SendRequest sends a JSON-RPC request throught the connection denoted by the connection ID.
func (r *Router) SendRequest(connID string, req *Request) chan<- *Response {
	r.jsonrpc.Send(connID, req)
	ch := make(chan *Response)
	r.pendinRequests[req.ID] = ch
	return ch
}

// SendNotification sends a JSON-RPC notification through the connection denoted by the connection ID.
func (r *Router) SendNotification(connID string, not *Notification) {
	r.jsonrpc.Send(connID, not)
}

func (r *Router) reqMiddleware(ctx *ReqContext) {
	if handler, ok := r.reqRoutes[ctx.Req.Method]; ok {
		handler(ctx)
	}
}

func (r *Router) notMiddleware(ctx *NotContext) {
	if handler, ok := r.notRoutes[ctx.Not.Method]; ok {
		handler(ctx)
	}
}

func (r *Router) resMiddleware(ctx *ResContext) {
	if ch, ok := r.pendinRequests[ctx.Res.ID]; ok {
		ch <- ctx.Res
		delete(r.pendinRequests, ctx.Res.ID)
	}
}
