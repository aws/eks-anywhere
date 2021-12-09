package executables

import "sync"

type syncSlice struct {
	internal []string
	sync.RWMutex
}

func newSyncSlice() *syncSlice {
	return &syncSlice{
		internal: []string{},
	}
}

func (s *syncSlice) append(v ...string) {
	s.Lock()
	defer s.Unlock()
	s.internal = append(s.internal, v...)
}

func (s *syncSlice) iterate() <-chan string {
	c := make(chan string)

	go func() {
		s.RLock()
		defer s.RUnlock()
		defer close(c)
		for _, v := range s.internal {
			c <- v
		}
	}()

	return c
}
