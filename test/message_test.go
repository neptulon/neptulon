package test

import (
	"sync"
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/neptulon/jsonrpc"
	"github.com/neptulon/jsonrpc/middleware"
	"github.com/neptulon/neptulon"
)

type echoMsg struct {
	Message string `json:"message"`
}

func TestEcho(t *testing.T) {
	sh := NewServerHelper(t).Start()
	defer sh.Close()

	rout := sh.GetRouter()
	rout.Request("echo", middleware.Echo)

	ch := sh.GetClientHelper().Connect()
	defer ch.Close()

	ch.SendRequest("echo", echoMsg{Message: "Hello!"}, func(ctx *jsonrpc.ResCtx) error {
		var msg echoMsg
		if err := ctx.Result(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Message != "Hello!" {
			t.Fatalf("expected: %v got: %v", "Hello!", msg.Message)
		}
		return ctx.Next()
	})
}

func TestEcho(t *testing.T) {
	s := neptulon.NewServer("127.0.0.1:3010")
	go s.Start()
	defer s.Close()
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

func TestTLS(t *testing.T) {
	// todo: client cert etc.
}
