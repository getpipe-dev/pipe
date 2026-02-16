package cache

import (
	"testing"
	"time"
)

func TestParseExpiry_Empty(t *testing.T) {
	exp, err := ParseExpiry("", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exp.IsZero() {
		t.Fatalf("expected zero time, got %v", exp)
	}
}

func TestParseExpiry_Duration(t *testing.T) {
	cachedAt := time.Date(2026, 2, 17, 10, 0, 0, 0, time.UTC)
	exp, err := ParseExpiry("1h", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := cachedAt.Add(time.Hour)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_DurationMinutes(t *testing.T) {
	cachedAt := time.Date(2026, 2, 17, 10, 0, 0, 0, time.UTC)
	exp, err := ParseExpiry("30m", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := cachedAt.Add(30 * time.Minute)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_DurationSeconds(t *testing.T) {
	cachedAt := time.Date(2026, 2, 17, 10, 0, 0, 0, time.UTC)
	exp, err := ParseExpiry("30s", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := cachedAt.Add(30 * time.Second)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_AbsoluteUTC_Future(t *testing.T) {
	// Cached at 10:00 UTC, expires at 18:10 UTC same day
	cachedAt := time.Date(2026, 2, 17, 10, 0, 0, 0, time.UTC)
	exp, err := ParseExpiry("18:10 UTC", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 2, 17, 18, 10, 0, 0, time.UTC)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_AbsoluteUTC_Past(t *testing.T) {
	// Cached at 20:00 UTC, 18:10 already passed → tomorrow
	cachedAt := time.Date(2026, 2, 17, 20, 0, 0, 0, time.UTC)
	exp, err := ParseExpiry("18:10 UTC", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 2, 18, 18, 10, 0, 0, time.UTC)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_AbsoluteLocal(t *testing.T) {
	loc := time.FixedZone("Test", 5*3600) // UTC+5
	cachedAt := time.Date(2026, 2, 17, 10, 0, 0, 0, loc)
	exp, err := ParseExpiry("15:00", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 2, 17, 15, 0, 0, 0, loc)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_AbsoluteLocal_Past(t *testing.T) {
	loc := time.FixedZone("Test", 5*3600)
	cachedAt := time.Date(2026, 2, 17, 16, 0, 0, 0, loc)
	exp, err := ParseExpiry("15:00", cachedAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := time.Date(2026, 2, 18, 15, 0, 0, 0, loc)
	if !exp.Equal(want) {
		t.Fatalf("expected %v, got %v", want, exp)
	}
}

func TestParseExpiry_Invalid(t *testing.T) {
	_, err := ParseExpiry("not-a-time", time.Now())
	if err == nil {
		t.Fatal("expected error for invalid expiry")
	}
}
