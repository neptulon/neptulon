package jsonrpc

// Router is a JSON-RPC request routing middleware.
type Router struct {
	requestRoutes      map[string]func(ctx *ReqContext)
	notificationRoutes map[string]func(ctx *NotContext)
}

// NewRouter creates a JSON-RPC router instance and registers it with the Neptulon JSON-RPC app.
func NewRouter(app *App) (*Router, error) {
	r := Router{
		requestRoutes:      make(map[string]func(ctx *ReqContext)),
		notificationRoutes: make(map[string]func(ctx *NotContext)),
	}

	app.Middleware(r.middleware)
	return &r, nil
}

// Request adds a new request route registry.
func (r *Router) Request(route string, handler func(ctx *ReqContext)) {
	r.requestRoutes[route] = handler
}

// Notification adds a new notification route registry.
func (r *Router) Notification(route string, handler func(ctx *NotContext)) {
	r.notificationRoutes[route] = handler
}

func (r *Router) middleware(ctx *Context) {
	// if not request or notification don't handle it
	if ctx.Msg.Method == "" {
		return
	}

	// if request
	if ctx.Msg.ID != "" {
		if handler, ok := r.requestRoutes[ctx.Msg.Method]; ok {
			rctx := ReqContext{Conn: ctx.Conn, Req: &Request{ID: ctx.Msg.ID, Method: ctx.Msg.Method, Params: ctx.Msg.Params}}
			if handler(&rctx); rctx.Res != nil || rctx.ResErr != nil {
				ctx.Res = rctx.Res
				ctx.ResErr = rctx.ResErr
			}
		}
	} else { // if notification
		if handler, ok := r.notificationRoutes[ctx.Msg.Method]; ok {
			ctx := NotContext{conn: ctx.Conn, not: &Notification{Method: ctx.Msg.Method, Params: ctx.Msg.Params}}
			handler(&ctx)
		}
	}
}
