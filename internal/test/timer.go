package test

import "time"

var (
	defaultNow = time.Unix(0, 1234567890000*int64(time.Millisecond))
	newNow     = time.Unix(0, 987654321000*int64(time.Millisecond))
)

func FakeNow() time.Time {
	return defaultNow
}

// NewFakeNow sets a dummy value for time.Now in unit tests.
// This is particularly useful when testing upgrade operations.
func NewFakeNow() time.Time {
	return newNow
}
