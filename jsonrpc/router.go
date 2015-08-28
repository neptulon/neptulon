package jsonrpc

// Router is a JSON-RPC request routing middleware.
type Router struct {
	reqRoutes map[string]func(ctx *ReqContext)
	notRoutes map[string]func(ctx *NotContext)
}

// NewRouter creates a JSON-RPC router instance and registers it with the Neptulon JSON-RPC app.
func NewRouter(app *App) (*Router, error) {
	r := Router{
		reqRoutes: make(map[string]func(ctx *ReqContext)),
		notRoutes: make(map[string]func(ctx *NotContext)),
	}

	app.ReqMiddleware(r.reqMiddleware)
	app.NotMiddleware(r.notMiddleware)
	return &r, nil
}

// Request adds a new request route registry.
// Optionally, you can pass in a data structure that the returned JSON-RPC response result data will be serialized into. Otherwise json.Unmarshal defaults apply.
func (r *Router) Request(route string, resultData interface{}, handler func(ctx *ReqContext)) {
	r.reqRoutes[route] = handler
}

// Notification adds a new notification route registry.
func (r *Router) Notification(route string, handler func(ctx *NotContext)) {
	r.notRoutes[route] = handler
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
