package test

import (
	"reflect"
	"sync"
	"testing"

	"github.com/neptulon/neptulon/client"
	"github.com/neptulon/neptulon/middleware"
)

func TestConnectTCP(t *testing.T) {
	sh := NewTCPServerHelper(t).MiddlewareIn(middleware.Echo).Start()
	defer sh.Close()

	var wg sync.WaitGroup
	msg := []byte("test message")

	ch := sh.GetTCPClientHelper().MiddlewareIn(func(ctx *client.Ctx) {
		defer wg.Done()
		if !reflect.DeepEqual(ctx.Msg, msg) {
			t.Fatalf("expected: '%s', got: '%s'", msg, ctx.Msg)
		}
		ctx.Next()
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

	ch := sh.GetTLSClientHelper().MiddlewareIn(func(ctx *client.Ctx) {
		defer wg.Done()
		if !reflect.DeepEqual(ctx.Msg, msg) {
			t.Fatalf("expected: '%s', got: '%s'", msg, ctx.Msg)
		}
		ctx.Next()
	}).Connect()
	defer ch.Close()

	wg.Add(1)
	ch.Send(msg)
	wg.Wait()
}
