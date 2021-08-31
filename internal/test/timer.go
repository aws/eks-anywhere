package test

import "time"

var defaultNow = time.Unix(0, 1234567890000*int64(time.Millisecond))

func FakeNow() time.Time {
	return defaultNow
}
