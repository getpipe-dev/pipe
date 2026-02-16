package cache

import (
	"fmt"
	"strings"
	"time"
)

// ParseExpiry computes the concrete expiration time from an expireAfter string.
// Supported formats:
//   - Duration: "10m", "1h", "30s" — expires at cachedAt + duration
//   - Absolute: "18:10 UTC", "18:10" — next occurrence of that wall-clock time
//   - Empty: returns zero time (no expiry, cache is permanent)
func ParseExpiry(expireAfter string, cachedAt time.Time) (time.Time, error) {
	if expireAfter == "" {
		return time.Time{}, nil
	}

	// Try duration first (e.g. "10m", "1h", "30s")
	if d, err := time.ParseDuration(expireAfter); err == nil {
		return cachedAt.Add(d), nil
	}

	// Try absolute time (e.g. "18:10 UTC", "18:10")
	return parseAbsoluteTime(expireAfter, cachedAt)
}

// parseAbsoluteTime handles "HH:MM UTC" and "HH:MM" (local time).
func parseAbsoluteTime(s string, cachedAt time.Time) (time.Time, error) {
	s = strings.TrimSpace(s)

	var loc *time.Location
	var timePart string

	if strings.HasSuffix(s, " UTC") {
		loc = time.UTC
		timePart = strings.TrimSuffix(s, " UTC")
	} else {
		loc = cachedAt.Location()
		timePart = s
	}

	parsed, err := time.Parse("15:04", timePart)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid expiry %q: expected duration (e.g. 1h) or time (e.g. 18:10 UTC)", s)
	}

	// Build today's occurrence of that time in the target timezone
	y, m, d := cachedAt.In(loc).Date()
	candidate := time.Date(y, m, d, parsed.Hour(), parsed.Minute(), 0, 0, loc)

	// If the time has already passed, push to tomorrow
	if !candidate.After(cachedAt) {
		candidate = candidate.Add(24 * time.Hour)
	}

	return candidate, nil
}
