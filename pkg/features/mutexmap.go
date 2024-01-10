package features

import "sync"

func newMutexMap() *mutexMap {
	return &mutexMap{
		internal: make(map[string]bool),
	}
}

type mutexMap struct {
	internal map[string]bool
	sync.RWMutex
}

func (m *mutexMap) load(key string) (value, ok bool) {
	m.RLock()
	result, ok := m.internal[key]
	m.RUnlock()
	return result, ok
}

func (m *mutexMap) store(key string, value bool) {
	m.Lock()
	m.internal[key] = value
	m.Unlock()
}

func (m *mutexMap) clear() {
	m.Lock()
	m.internal = make(map[string]bool)
	m.Unlock()
}
