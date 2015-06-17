package jsonrpc

import "github.com/nbusy/neptulon"

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

func (r *Router) middleware(conn *neptulon.Conn, msg *Message) (result interface{}, err *ResError) {
	// if not request or notification don't handle it
	if msg.Method == "" {
		return nil, nil
	}

	// if request
	if msg.ID != "" {
		if handler, ok := r.requestRoutes[msg.Method]; ok {
			ctx := ReqContext{Conn: conn, Req: &Request{ID: msg.ID, Method: msg.Method, Params: msg.Params}}
			if handler(&ctx); ctx.Res != nil || ctx.ResErr != nil {
				return ctx.Res, ctx.ResErr
			}
		}
	} else { // if notification
		if handler, ok := r.notificationRoutes[msg.Method]; ok {
			ctx := NotContext{conn: conn, not: &Notification{Method: msg.Method, Params: msg.Params}}
			handler(&ctx)
			// todo: need to return something to prevent deeper handlers to further handle this request (i.e. not found handler logging not found warning)
		}
	}

	return nil, nil
}
