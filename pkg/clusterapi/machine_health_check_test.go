package clusterapi_test

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestMachineHealthCheckForControlPlane(t *testing.T) {
	timeouts := []time.Duration{5 * time.Minute, time.Hour, 30 * time.Second}
	maxUnhealthy := intstr.Parse("80%")
	for _, timeout := range timeouts {
		tt := newApiBuilerTest(t)
		want := expectedMachineHealthCheckForControlPlane(timeout, maxUnhealthy)
		tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
			NodeStartupTimeout: &metav1.Duration{
				Duration: timeout,
			},
			UnhealthyMachineTimeout: &metav1.Duration{
				Duration: timeout,
			},
			MaxUnhealthy: &maxUnhealthy,
		}
		got := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec.Cluster)
		tt.Expect(got).To(BeComparableTo(want))
	}
}

func TestMachineHealthCheckForControlPlaneWithTimeoutOverride(t *testing.T) {
	defaultTimeout := 30 * time.Minute
	cpTimeout := 60 * time.Minute
	maxUnhealthy := intstr.Parse("100%")

	tt := newApiBuilerTest(t)
	want := expectedMachineHealthCheckForControlPlane(cpTimeout, maxUnhealthy)
	tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: cpTimeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: cpTimeout,
		},
	}
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: defaultTimeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: defaultTimeout,
		},
		MaxUnhealthy: &maxUnhealthy,
	}
	got := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec.Cluster)
	tt.Expect(got).To(BeComparableTo(want))
}

func TestMachineHealthCheckForControlPlaneWithMaxUnhealthyOverride(t *testing.T) {
	timeout := 30 * time.Minute
	defaultMaxUnhealthy := intstr.Parse("40%")
	cpMaxUnhealthyOverride := intstr.Parse("100%")

	tt := newApiBuilerTest(t)
	want := expectedMachineHealthCheckForControlPlane(timeout, cpMaxUnhealthyOverride)
	tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		MaxUnhealthy: &cpMaxUnhealthyOverride,
	}
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: timeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: timeout,
		},
		MaxUnhealthy: &defaultMaxUnhealthy,
	}
	got := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec.Cluster)
	tt.Expect(got).To(BeComparableTo(want))
}

func expectedMachineHealthCheckForControlPlane(timeout time.Duration, maxUnhealthy intstr.IntOrString) *clusterv1.MachineHealthCheck {
	return &clusterv1.MachineHealthCheck{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "MachineHealthCheck",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-kcp-unhealthy",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: clusterv1.MachineHealthCheckSpec{
			ClusterName: "test-cluster",
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{
					"cluster.x-k8s.io/control-plane": "",
				},
			},
			MaxUnhealthy: &maxUnhealthy,
			NodeStartupTimeout: &metav1.Duration{
				Duration: timeout,
			},
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionUnknown,
					Timeout: metav1.Duration{
						Duration: timeout,
					},
				},
				{
					Type:   corev1.NodeReady,
					Status: corev1.ConditionFalse,
					Timeout: metav1.Duration{
						Duration: timeout,
					},
				},
			},
		},
	}
}

func TestMachineHealthCheckForWorkers(t *testing.T) {
	maxUnhealthy := intstr.Parse("40%")
	timeouts := []time.Duration{5 * time.Minute, time.Hour, 30 * time.Second}
	for _, timeout := range timeouts {
		tt := newApiBuilerTest(t)
		want := expectedMachineHealthCheckForWorkers(timeout, maxUnhealthy)
		tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
		tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
			NodeStartupTimeout: &metav1.Duration{
				Duration: timeout,
			},
			UnhealthyMachineTimeout: &metav1.Duration{
				Duration: timeout,
			},
			MaxUnhealthy: &maxUnhealthy,
		}
		got := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec.Cluster)
		tt.Expect(got).To(Equal(want))
	}
}

func TestMachineHealthCheckForWorkersWithTimeoutOverride(t *testing.T) {
	defaultTimeout := 30 * time.Minute
	workerTimeout := 60 * time.Minute
	maxUnhealthy := intstr.Parse("40%")

	tt := newApiBuilerTest(t)
	want := expectedMachineHealthCheckForWorkers(workerTimeout, maxUnhealthy)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: workerTimeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: workerTimeout,
		},
	}
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: defaultTimeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: defaultTimeout,
		},
		MaxUnhealthy: &maxUnhealthy,
	}
	got := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec.Cluster)
	tt.Expect(got).To(Equal(want))
}

func TestMachineHealthCheckForWorkersWithMaxUnhealthyOverride(t *testing.T) {
	timeout := 30 * time.Minute
	defaultMaxUnhealthy := intstr.Parse("40%")
	workerMaxUnhealthyOverride := intstr.Parse("100%")

	tt := newApiBuilerTest(t)
	want := expectedMachineHealthCheckForWorkers(timeout, workerMaxUnhealthyOverride)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		MaxUnhealthy: &workerMaxUnhealthyOverride,
	}
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: timeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: timeout,
		},
		MaxUnhealthy: &defaultMaxUnhealthy,
	}
	got := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec.Cluster)
	tt.Expect(got).To(Equal(want))
}

func expectedMachineHealthCheckForWorkers(timeout time.Duration, maxUnhealthy intstr.IntOrString) []*clusterv1.MachineHealthCheck {
	return []*clusterv1.MachineHealthCheck{
		{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cluster.x-k8s.io/v1beta1",
				Kind:       "MachineHealthCheck",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-wng-1-worker-unhealthy",
				Namespace: constants.EksaSystemNamespace,
			},
			Spec: clusterv1.MachineHealthCheckSpec{
				ClusterName: "test-cluster",
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{
						"cluster.x-k8s.io/deployment-name": "test-cluster-wng-1",
					},
				},
				MaxUnhealthy:       &maxUnhealthy,
				NodeStartupTimeout: &metav1.Duration{Duration: timeout},
				UnhealthyConditions: []clusterv1.UnhealthyCondition{
					{
						Type:    corev1.NodeReady,
						Status:  corev1.ConditionUnknown,
						Timeout: metav1.Duration{Duration: timeout},
					},
					{
						Type:    corev1.NodeReady,
						Status:  corev1.ConditionFalse,
						Timeout: metav1.Duration{Duration: timeout},
					},
				},
			},
		},
	}
}

func TestMachineHealthCheckObjects(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
	timeout := 5 * time.Minute
	tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
		NodeStartupTimeout: &metav1.Duration{
			Duration: timeout,
		},
		UnhealthyMachineTimeout: &metav1.Duration{
			Duration: timeout,
		},
	}

	wantWN := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec.Cluster)
	wantCP := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec.Cluster)

	got := clusterapi.MachineHealthCheckObjects(tt.clusterSpec.Cluster)
	tt.Expect(got).To(Equal([]kubernetes.Object{wantWN[0], wantCP}))
}
