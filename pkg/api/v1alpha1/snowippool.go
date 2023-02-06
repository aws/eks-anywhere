package v1alpha1

import (
	"bytes"
	"fmt"
	"net"
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

func validateSnowIPPool(pool *SnowIPPool) error { //nolint:gocyclo
	for index, ipPool := range pool.Spec.Pools {
		if len(ipPool.IPStart) == 0 {
			return fmt.Errorf("SnowIPPool Pools[%d].IPStart can not be empty", index)
		}

		ipStart := net.ParseIP(ipPool.IPStart)
		if ipStart == nil {
			return fmt.Errorf("SnowIPPool Pools[%d].IPStart is invalid", index)
		}

		if len(ipPool.IPEnd) == 0 {
			return fmt.Errorf("SnowIPPool Pools[%d].IPEnd can not be empty", index)
		}

		ipEnd := net.ParseIP(ipPool.IPEnd)
		if ipEnd == nil {
			return fmt.Errorf("SnowIPPool Pools[%d].IPEnd is invalid", index)
		}

		if len(ipPool.Gateway) == 0 {
			return fmt.Errorf("SnowIPPool Pools[%d].Gateway can not be empty", index)
		}

		gateway := net.ParseIP(ipPool.Gateway)
		if gateway == nil {
			return fmt.Errorf("SnowIPPool Pools[%d].Gateway is invalid", index)
		}

		if bytes.Compare(ipStart, ipEnd) >= 0 {
			return fmt.Errorf("SnowIPPool Pools[%d].IPStart should be smaller than IPEnd", index)
		}

		if len(ipPool.Subnet) == 0 {
			return fmt.Errorf("SnowIPPool Pools[%d].Subnet can not be empty", index)
		}

		_, ipNet, err := net.ParseCIDR(ipPool.Subnet)
		if err != nil {
			return fmt.Errorf("SnowIPPool Pools[%d].Subnet is invalid: %v", index, err)
		}

		if !ipNet.Contains(ipStart) {
			return fmt.Errorf("SnowIPPool Pools[%d].IPStart should be within the subnet range %s", index, ipPool.Subnet)
		}

		if !ipNet.Contains(ipEnd) {
			return fmt.Errorf("SnowIPPool Pools[%d].IPEnd should be within the subnet range %s", index, ipPool.Subnet)
		}
	}

	return nil
}
