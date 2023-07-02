package e2e

import (
	"fmt"
	"sync"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
)

// Default timeout for Tink tests to poll for hardware.
const hwPollingTimeout = 120 * time.Minute

type hardwareQueue struct {
	hws []*api.Hardware
	mu  sync.Mutex
}

func (hwQu *hardwareQueue) reserveHardware(count int) ([]*api.Hardware, error) {
	now := time.Now()
	after := now.Add(hwPollingTimeout)
	for {
		if now.After(after) {
			return nil, fmt.Errorf("hardware polling request timed out")
		}
		hwQu.mu.Lock()
		if count <= len(hwQu.hws) {
			hardwareReserved := hwQu.dequeueHw(count)
			hwQu.mu.Unlock()
			return hardwareReserved, nil
		}
		hwQu.mu.Unlock()
		time.Sleep(1 * time.Minute)
	}
}

func (hwQu *hardwareQueue) releaseHardware(hardwareToRelease []*api.Hardware) {
	hwQu.mu.Lock()
	hwQu.enqueueHw(hardwareToRelease)
	hwQu.mu.Unlock()
}

func (hwQu *hardwareQueue) enqueueHw(hws []*api.Hardware) {
	hwQu.hws = append(hwQu.hws, hws...)
}

func (hwQu *hardwareQueue) dequeueHw(count int) []*api.Hardware {
	hwsToRet := hwQu.hws[:count]
	hwQu.hws = hwQu.hws[count:]
	return hwsToRet
}

func newHardwareQueue(hws []*api.Hardware) *hardwareQueue {
	return &hardwareQueue{
		hws: hws,
		mu:  sync.Mutex{},
	}
}
