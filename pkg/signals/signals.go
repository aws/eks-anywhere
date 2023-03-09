package signals

import (
	"os"
	"os/signal"
	"sync"
)

// On registers a function to be called when one of sigs is received. If no sigs are provided On
// is a noop. It returns a cancellation func that can be used to unregister the signals. The
// cancellation func is idempotent.
func On(fn func(), sigs ...os.Signal) func() {
	if len(sigs) == 0 {
		return func() {}
	}

	receive := make(chan os.Signal, 1)
	done := make(chan struct{})
	signal.Notify(receive, sigs...)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-receive:
				fn()
			}
		}
	}()

	var once sync.Once
	return func() { once.Do(func() { close(done) }) }
}
