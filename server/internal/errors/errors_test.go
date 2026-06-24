package errors

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorWithMessage(t *testing.T) {
	e := BadRequest.WithMessage("invalid input")
	if e.Code != 40000 {
		t.Fatalf("Code mismatch: %d", e.Code)
	}
	if e.Message != "invalid input" {
		t.Fatalf("Message mismatch: %s", e.Message)
	}
}

func TestErrorWithCause(t *testing.T) {
	cause := errors.New("disk full")
	e := Internal.WithCause(cause)
	if e.cause != cause {
		t.Fatal("cause not stored")
	}
	if e.Error() == "" {
		t.Fatal("Error() returned empty")
	}
}

func TestAsReturnsTrue(t *testing.T) {
	be := BadRequest
	as, ok := As(be)
	if !ok {
		t.Fatal("As() should return true for *Error")
	}
	if as.Code != 40000 {
		t.Fatalf("Code mismatch: %d", as.Code)
	}

	as2, ok := As(errors.New("plain"))
	if ok || as2 != nil {
		t.Fatal("As() should return false for non-biz error")
	}
}

func TestNew(t *testing.T) {
	e := New(12345, "custom", http.StatusBadRequest)
	if e.Code != 12345 || e.HTTP != http.StatusBadRequest {
		t.Fatalf("unexpected: %+v", e)
	}
}
