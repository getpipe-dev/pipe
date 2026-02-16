package runner

import (
	"time"
)

// Retry calls fn up to maxAttempts times. On failure it waits attempt*2 seconds
// before retrying. Returns the last error and the number of attempts made.
func Retry(maxAttempts int, fn func() error) (int, error) {
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	var err error
	for i := 1; i <= maxAttempts; i++ {
		err = fn()
		if err == nil {
			return i, nil
		}
		if i < maxAttempts {
			time.Sleep(time.Duration(i*2) * time.Second)
		}
	}
	return maxAttempts, err
}
