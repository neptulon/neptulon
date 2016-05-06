package test

import (
	"testing"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
)

func TestMiddlewarePanic(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		panic("much panic")
	})
	sh.Server.MiddlewareFunc(middleware.Echo)
	defer sh.ListenAndServe().CloseWait()

	ch := sh.GetConnHelper().Connect()
	defer ch.CloseWait()

	// ch.SendRequest("echo", echoMsg{Message: m}, func(ctx *neptulon.ResCtx) error {
	// 	var msg echoMsg
	// 	if err := ctx.Result(&msg); err != nil {
	// 		t.Fatal(err)
	// 	}
	// 	if msg.Message != m {
	// 		t.Fatalf("expected: %v got: %v", m, msg.Message)
	// 	}
	// 	return nil
	// })
}

func TestMiddlewareErrorReturn(t *testing.T) {

}
