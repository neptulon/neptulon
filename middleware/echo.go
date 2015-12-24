package middleware

import "github.com/neptulon/neptulon/client"

// Echo sends incoming messages back as is.
func Echo(ctx *client.Ctx) error {
	ctx.Res = ctx.Msg
	return ctx.Next()
}
