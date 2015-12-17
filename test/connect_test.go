package test

import "testing"

// func TestConnectTCP(t *testing.T) {
// 	s := NewTCPServerHelper(t)
// 	defer s.Close()
// 	c := s.GetTCPClient()
// 	defer c.Close()
// }

func TestConnectTLS(t *testing.T) {
	sh := NewTLSServerHelper(t)
	defer sh.Close()
	// sh.GetTLSClient(true)
	// defer ch.Close()

	// msg := []byte("test message")
	//
	// sh.Server.MiddlewareIn(middleware.Echo)
	// ch.Client.MiddlewareIn(func(ctx *client.Ctx) {
	// 	if !reflect.DeepEqual(ctx.Msg, msg) {
	// 		t.Fatalf("expected: '%s', got: '%s'", msg, ctx.Msg)
	// 	}
	// 	ctx.Next()
	// })
	//
	// ch.Client.Send(msg) // todo: use fail-safe ClientHelper.Send instead
}
