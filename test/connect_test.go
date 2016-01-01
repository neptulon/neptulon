package test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
)

func TestConnectTCP(t *testing.T) {
	sh := NewTCPServerHelper(t).MiddlewareIn(middleware.Echo).Start()
	defer sh.Close()

	var wg sync.WaitGroup
	msg := []byte("test message")

	ch := sh.GetTCPClientHelper().MiddlewareIn(func(ctx *neptulon.Ctx) error {
		defer wg.Done()
		if !reflect.DeepEqual(ctx.Msg, msg) {
			t.Fatalf("expected: '%s', got: '%s'", msg, ctx.Msg)
		}
		return ctx.Next()
	}).Connect()
	defer ch.Close()

	wg.Add(1)
	ch.Send(msg)
	wg.Wait()
}

func TestConnectTLS(t *testing.T) {
	sh := NewTLSServerHelper(t).MiddlewareIn(middleware.Echo).Start()
	defer sh.Close()

	var wg sync.WaitGroup
	msg := []byte("test message")

	ch := sh.GetTLSClientHelper().MiddlewareIn(func(ctx *neptulon.Ctx) error {
		defer wg.Done()
		if !reflect.DeepEqual(ctx.Msg, msg) {
			t.Fatalf("expected: '%s', got: '%s'", msg, ctx.Msg)
		}
		return ctx.Next()
	}).Connect()
	defer ch.Close()

	wg.Add(1)
	ch.Send(msg)
	wg.Wait()
}
