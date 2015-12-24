package middleware

import "github.com/neptulon/neptulon/client"

// Echo sends incoming messages back as is.
func Echo(ctx *client.Ctx) error {
	if err := ctx.Client.Send(ctx.Msg); err != nil {
		return err
	}

	return ctx.Next()
}
