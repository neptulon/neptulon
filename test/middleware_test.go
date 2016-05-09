package test

import (
	"log"
	"testing"
	"time"

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

	gotRes := make(chan bool)
	ch.Conn.SendRequest("echo", echoMsg{Message: "just testing"}, func(ctx *neptulon.ResCtx) error {
		gotRes <- true
		return nil
	})

	select {
	case <-gotRes:
		log.Fatal("expected no response, got one")
	case <-time.After(time.Millisecond * 25):
	}

	// todo: verify that the server is still up and functional
}

func TestMiddlewareErrorReturn(t *testing.T) {

}
