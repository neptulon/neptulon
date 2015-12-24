package middleware

import "github.com/neptulon/neptulon/client"

// Echo sends incoming messages back as is.
func Echo(ctx *client.Ctx) {
	ctx.Conn.Write(ctx.Msg)
	ctx.Next()
}
