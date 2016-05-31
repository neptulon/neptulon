package middleware

import (
	"bytes"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/neptulon/cmap"
	"github.com/neptulon/neptulon"
)

func captureOutput(f func()) string {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	f()
	log.SetOutput(os.Stderr)
	return buf.String()
}

func TestResLog(t *testing.T) {
	ctx := &neptulon.ReqCtx{
		Session: cmap.New(),
		ID:      "1234",
		Method:  "wow.method",
		Res:     "my response",
		Err:     &neptulon.ResError{Code: 98765},
	}

	out := captureOutput(func() {
		err := Logger(ctx)
		if err != nil {
			t.Fatal("didn't expect error from logger")
		}
	})

	if !strings.Contains(out, "my response") || strings.Contains(out, "98765") {
		log.Fatalf("malformed log output: %v", out)
	}
}

func TestErrResLog(t *testing.T) {
	ctx := &neptulon.ReqCtx{
		Session: cmap.New(),
		ID:      "1234",
		Method:  "wow.method",
		Err:     &neptulon.ResError{Code: 98765},
	}

	out := captureOutput(func() {
		err := Logger(ctx)
		if err != nil {
			t.Fatal("didn't expect error from logger")
		}
	})

	if strings.Contains(out, "my response") || !strings.Contains(out, "98765") {
		log.Fatalf("malformed log output: %v", out)
	}
}

func TestCustLog(t *testing.T) {
	s := cmap.New()
	s.Set(CustResLogDataKey, "custom log output")

	ctx := &neptulon.ReqCtx{
		Session: s,
		ID:      "1234",
		Method:  "wow.method",
		Res:     "my response",
		Err:     &neptulon.ResError{Code: 98765},
	}

	out := captureOutput(func() {
		err := Logger(ctx)
		if err != nil {
			t.Fatal("didn't expect error from logger")
		}
	})

	if strings.Contains(out, "my response") || strings.Contains(out, "98765") {
		log.Fatalf("malformed log output: %v", out)
	}
}
