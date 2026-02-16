package runner

import (
	"errors"
	"testing"
)

func TestRetry_ImmediateSuccess(t *testing.T) {
	attempts, err := Retry(3, func() error { return nil })
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_FailThenSucceed(t *testing.T) {
	call := 0
	attempts, err := Retry(2, func() error {
		call++
		if call == 1 {
			return errors.New("fail once")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Fatalf("expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_AllExhausted(t *testing.T) {
	attempts, err := Retry(1, func() error {
		return errors.New("always fail")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_ClampsToOne(t *testing.T) {
	attempts, err := Retry(0, func() error {
		return errors.New("fail")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts != 1 {
		t.Fatalf("expected 1 attempt (clamped), got %d", attempts)
	}
}

func TestRetry_ReturnsLastError(t *testing.T) {
	call := 0
	_, err := Retry(1, func() error {
		call++
		return errors.New("error from final attempt")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "error from final attempt" {
		t.Fatalf("expected last error, got %q", err.Error())
	}
}
