package test

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
)

func TestMiddlewarePanics(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		panic("much panic")
	})
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

	// todo: rather than waiting 25 millisecs, add a Conn.Disconnected handler and check that it is called
	// todo: verify that the server is still up and functional
}

func TestMiddlewareRetursError(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		return errors.New("much error")
	})
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
}

func TestErrorHandlerMiddleware(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.MiddlewareFunc(middleware.Error)
	sh.Server.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		return errors.New("much error")
	})
	defer sh.ListenAndServe().CloseWait()

	ch := sh.GetConnHelper().Connect()
	defer ch.CloseWait()

	ch.SendRequestSync("echo", echoMsg{Message: "just testing"}, func(ctx *neptulon.ResCtx) error {
		return nil
	})
}
