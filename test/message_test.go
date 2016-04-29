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
	go s.ListenAndServe()
	time.Sleep(time.Millisecond * 30)
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

	time.Sleep(time.Millisecond * 30)
}

func TestEcho(t *testing.T) {
	sh := NewServerHelper(t)
	rout := middleware.NewRouter()
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.Middleware(rout)
	rout.Request("echo", middleware.Echo)
	defer sh.ListenAndServe().CloseWait()

	ch := sh.GetConnHelper()
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

func TestError(t *testing.T) {
	sh := NewServerHelper(t)
	sh.Server.MiddlewareFunc(middleware.Logger)
	sh.Server.MiddlewareFunc(func(ctx *neptulon.ReqCtx) error {
		ctx.Err = &neptulon.ResError{
			Code:    1234,
			Message: "much error",
			Data:    map[string]string{"keykey": "valuevalue"},
		}
		return ctx.Next()
	})
	defer sh.ListenAndServe().CloseWait()

	ch := sh.GetConnHelper()
	defer ch.Connect().CloseWait()

	ch.SendRequest("testerror", nil, func(ctx *neptulon.ResCtx) error {
		var v map[string]string
		if ctx.Success {
			t.Error("expected to get error response")
		}
		if ctx.Result(&v) == nil {
			t.Error("did not expect to get any result for expected error response")
		}
		if ctx.ErrorCode != 1234 {
			t.Errorf("expected error code %v got %v", 1234, ctx.ErrorCode)
		}
		if ctx.ErrorMessage != "much error" {
			t.Errorf("expected error message %v got %v", "much error", ctx.ErrorMessage)
		}
		if ctx.ErrorData(&v) != nil || v["keykey"] != "valuevalue" {
			t.Errorf("expected error data %v got %v or errored during deserialization", "valuevalue", v["keykey"])
		}

		// todo: verify that conn is closed

		return nil
	})
}

func TestPanic(t *testing.T) {
	// todo: panic from inside a req handler and make sure that server/client does not crash and conn is closed
}
