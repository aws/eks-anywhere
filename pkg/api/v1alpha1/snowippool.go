package v1alpha1

import (
	"fmt"

	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

const (
	// SnowIPPoolKind is the object kind name for SnowIPPool.
	SnowIPPoolKind = "SnowIPPool"
)

// SnowIPPoolsSliceEqual compares and returns whether two snow IPPool objects are equal.
func SnowIPPoolsSliceEqual(a, b []snowv1.IPPool) bool {
	if len(a) != len(b) {
		return false
	}

	m := make(map[string]int, len(a))
	for _, v := range a {
		m[generateKeyForSnowIPPool(v)]++
	}
	for _, v := range b {
		k := generateKeyForSnowIPPool(v)
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

func generateKeyForSnowIPPool(pool snowv1.IPPool) string {
	return fmt.Sprintf("%s%s%s%s", keyFromStrPtr(pool.IPStart), keyFromStrPtr(pool.IPEnd), keyFromStrPtr(pool.Subnet), keyFromStrPtr(pool.Gateway))
}

func keyFromStrPtr(s *string) string {
	if s == nil {
		return "nil"
	}
	return *s
}
