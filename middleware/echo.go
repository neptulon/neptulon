package middleware

import "github.com/neptulon/client"

// Echo sends incoming messages back as is.
func Echo(ctx *client.Ctx) {
	ctx.Client.Send(ctx.Msg)
	ctx.Next()
}
