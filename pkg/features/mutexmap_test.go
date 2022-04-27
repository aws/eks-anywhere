package features

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestMutexMapLoadAndStore(t *testing.T) {
	g := NewWithT(t)
	m := newMutexMap()

	key := "key"
	value := true
	v, ok := m.load(key)
	g.Expect(ok).To(BeFalse())
	g.Expect(v).To(BeFalse())

	m.store(key, value)
	v, ok = m.load(key)
	g.Expect(ok).To(BeTrue())
	g.Expect(v).To(Equal(value))
}

func TestMutexMapClear(t *testing.T) {
	g := NewWithT(t)
	m := newMutexMap()

	key := "key"
	value := true
	m.store(key, value)
	v, ok := m.load(key)
	g.Expect(ok).To(BeTrue())
	g.Expect(v).To(Equal(value))

	m.clear()
	v, ok = m.load(key)
	g.Expect(ok).To(BeFalse())
	g.Expect(v).To(BeFalse())
}
