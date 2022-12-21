package v1alpha1

import (
	"fmt"
)

const (
	// SnowIPPoolKind is the object kind name for SnowIPPool.
	SnowIPPoolKind = "SnowIPPool"
)

// SnowIPPoolsSliceEqual compares and returns whether two snow IPPool objects are equal.
func SnowIPPoolsSliceEqual(a, b []IPPool) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[generateKeyForIPPool(v)]++
	}
	for _, v := range b {
		k := generateKeyForIPPool(v)
		if _, ok := m[k]; !ok {
			return false
		}
		m[k]--
		if m[k] == 0 {
			delete(m, k)
		}
	}
	return true
}

func generateKeyForIPPool(pool IPPool) string {
	return fmt.Sprintf("%s%s%s%s", pool.IPStart, pool.IPEnd, pool.Subnet, pool.Gateway)
}
