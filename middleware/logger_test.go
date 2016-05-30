package middleware

import (
	"testing"

	"github.com/neptulon/cmap"
	"github.com/neptulon/neptulon"
)

func TestResLog(t *testing.T) {
	err := Logger(&neptulon.ReqCtx{
		Session: cmap.New(),
		ID:      "1234",
		Method:  "wow.method",
		Res:     "my response",
	})

	if err != nil {
		t.Fatal("unexpected error return value")
	}
}

func TestErrResLog(t *testing.T) {
	err := Logger(&neptulon.ReqCtx{
		Session: cmap.New(),
		ID:      "1234",
		Method:  "wow.method",
		Err:     &neptulon.ResError{Code: 98765},
	})

	if err != nil {
		t.Fatal("unexpected error return value")
	}
}

func TestCustLog(t *testing.T) {
	s := cmap.New()
	s.Set(CustResLogDataKey, "custom log output")

	err := Logger(&neptulon.ReqCtx{
		Session: s,
		ID:      "1234",
		Method:  "wow.method",
		Res:     "my response",
	})

	if err != nil {
		t.Fatal("unexpected error return value")
	}
}
