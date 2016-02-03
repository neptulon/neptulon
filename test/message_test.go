package test

import (
	"testing"
	"time"

	"golang.org/x/net/websocket"

	"github.com/neptulon/neptulon"
	"github.com/neptulon/neptulon/middleware"
	"github.com/neptulon/randstr"
)

type echoMsg struct {
	Message string `json:"message"`
}

var (
	msg1 = "Lorem ipsum dolor sit amet, consectetur adipiscing elit."
	msg2 = "In sit amet lectus felis, at pellentesque turpis."
	msg3 = "Nunc urna enim, cursus varius aliquet ac, imperdiet eget tellus."
	msg4 = randstr.Get(45 * 1000)       // 0.45 MB
	msg5 = randstr.Get(5 * 1000 * 1000) // 5.0 MB
)

func TestEchoWithoutTestHelpers(t *testing.T) {
	s := neptulon.NewServer("127.0.0.1:3001")
	go s.Start()
	time.Sleep(time.Millisecond * 10)
	defer s.Close()

	s.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		t.Log("Request received:", ctx.Method)
		ctx.Res = "response-wow!"
		return ctx.Next()
	})

	origin := "http://127.0.0.1"
	url := "ws://127.0.0.1:3001"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		t.Fatal(err)
	}
	if err := websocket.JSON.Send(ws, map[string]string{"id": "123", "method": "test"}); err != nil {
		t.Fatal(err)
	}
	var res interface{}
	if err := websocket.JSON.Receive(ws, &res); err != nil {
		t.Fatal(err)
	}
	t.Log("Got response:", res)

	if err := ws.Close(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Millisecond * 10)
}

func TestEcho(t *testing.T) {
	sh := NewServerHelper(t)
	rout := middleware.NewRouter()
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.Middleware(rout)
	rout.Request("echo", middleware.Echo)
	defer sh.Start().CloseWait()

	ch := sh.GetConnHelper()
	ch.Conn.MiddlewareFunc(middleware.Logger)
	defer ch.Connect().CloseWait()

	m := "Hello!"
	ch.SendRequest("echo", echoMsg{Message: m}, func(ctx *neptulon.ResCtx) error {
		var msg echoMsg
		if err := ctx.Result(&msg); err != nil {
			t.Fatal(err)
		}
		if msg.Message != m {
			t.Fatalf("expected: %v got: %v", m, msg.Message)
		}
		return nil
	})
}

func TestMessages(t *testing.T) {
	// todo: verify all message echoes from small to big
}

func TestBidirectional(t *testing.T) {
	// todo: test simultaneous read/writes
}

func TestTLS(t *testing.T) {
	// todo: ...
}
