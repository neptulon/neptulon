package test

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/neptulon/neptulon"
)

func TestEcho(t *testing.T) {
	s := neptulon.NewServer("127.0.0.1:3010")
	go s.Start()
	time.Sleep(time.Millisecond)

	var wg sync.WaitGroup
	s.Middleware(func(ctx *neptulon.ReqCtx) error {
		defer wg.Done()
		t.Log("Request received:", ctx.Method)
		ctx.Res = "response-wow!"
		return ctx.Next()
	})

	wg.Add(1)

	origin := "http://127.0.0.1"
	url := "ws://127.0.0.1:3010"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	if err := websocket.JSON.Send(ws, neptulon.Request{ID: "123", Method: "test"}); err != nil {
		t.Fatal(err)
	}
	var res neptulon.Response
	if err := websocket.JSON.Receive(ws, &res); err != nil {
		t.Fatal(err)
	}
	t.Log("Got response:", res)

	wg.Wait()
}