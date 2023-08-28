package clustermanager

import (
	"math"
	"testing"
	"time"

	. "github.com/onsi/gomega"
)

var (
	ClusterctlMoveRetryPolicy                      = clusterctlMoveRetryPolicy
	ClusterctlMoveWaitForInfrastructureRetryPolicy = clusterctlMoveWaitForInfrastructureRetryPolicy
)

func TestClusterManager_totalTimeoutForMachinesReadyWait(t *testing.T) {
	tests := []struct {
		name     string
		replicas int
		opts     []ClusterManagerOpt
		want     time.Duration
	}{
		{
			name:     "default timeouts with 1 replica",
			replicas: 1,
			want:     30 * time.Minute,
		},
		{
			name:     "default timeouts with 2 replicas",
			replicas: 2,
			want:     30 * time.Minute,
		},
		{
			name:     "default timeouts with 4 replicas",
			replicas: 4,
			want:     40 * time.Minute,
		},
		{
			name:     "no timeouts with 1 replica",
			replicas: 1,
			opts:     []ClusterManagerOpt{WithNoTimeouts()},
			want:     math.MaxInt64,
		},
		{
			name:     "no timeouts with 2 replicas",
			replicas: 2,
			opts:     []ClusterManagerOpt{WithNoTimeouts()},
			want:     math.MaxInt64,
		},
		{
			name:     "no timeouts with 1 replica",
			replicas: 1,
			opts:     []ClusterManagerOpt{WithNoTimeouts()},
			want:     math.MaxInt64,
		},
		{
			name:     "no timeouts with 0 replicas",
			replicas: 1,
			opts:     []ClusterManagerOpt{WithNoTimeouts()},
			want:     math.MaxInt64,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(nil, nil, nil, nil, nil, nil, nil, tt.opts...)
			g := NewWithT(t)
			g.Expect(c.totalTimeoutForMachinesReadyWait(tt.replicas)).To(Equal(tt.want))
		})
	}
}
