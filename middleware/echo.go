package middleware

import "github.com/neptulon/neptulon"

// Echo sends incoming messages back as is.
func Echo(ctx *neptulon.ReqCtx) error {
	var msg interface{}
	if err := ctx.Params(&msg); err != nil {
		return err
	}

	ctx.Res = msg
	return ctx.Next()
}
