package e2e

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

// Default timeout for Tink tests to poll for hardware.
const hwPollingTimeout = 120 * time.Minute

// hardwareCatalogue has a thread safe FIFO queue implementation to facilitate hardware reservation.
type hardwareCatalogue struct {
	hws []*api.Hardware
	mu  sync.Mutex
}

func (hwQu *hardwareCatalogue) reserveHardware(count int) ([]*api.Hardware, error) {
	now := time.Now()
	after := now.Add(hwPollingTimeout)
	for {
		if now.After(after) {
			return nil, fmt.Errorf("hardware polling request timed out")
		}
		hwQu.mu.Lock()
		if count <= len(hwQu.hws) {
			hardwareReserved := hwQu.acquireHw(count)
			hwQu.mu.Unlock()
			return hardwareReserved, nil
		}
		hwQu.mu.Unlock()
		time.Sleep(1 * time.Minute)
	}
}

func (hwQu *hardwareCatalogue) releaseHardware(hws []*api.Hardware) {
	hwQu.mu.Lock()
	hwQu.hws = append(hwQu.hws, hws...)
	hwQu.mu.Unlock()
}

func (hwQu *hardwareCatalogue) acquireHw(count int) []*api.Hardware {
	hwsToRet := hwQu.hws[:count]
	hwQu.hws = hwQu.hws[count:]
	return hwsToRet
}

func newHardwareCatalogue(hws []*api.Hardware) *hardwareCatalogue {
	return &hardwareCatalogue{
		hws: hws,
		mu:  sync.Mutex{},
	}
}
