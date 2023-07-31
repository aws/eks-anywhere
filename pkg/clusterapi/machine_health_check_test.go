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
	for _, timeout := range timeouts {
		tt := newApiBuilerTest(t)
		want := expectedMachineHealthCheckForControlPlane(timeout)
		tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
			NodeStartupTimeout: &metav1.Duration{
				Duration: timeout,
			},
			UnhealthyMachineTimeout: &metav1.Duration{
				Duration: timeout,
			},
		}
		got := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec.Cluster)
		tt.Expect(got).To(BeComparableTo(want))
	}
}

func expectedMachineHealthCheckForControlPlane(timeout time.Duration) *clusterv1.MachineHealthCheck {
	maxUnhealthy := intstr.Parse("100%")
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
	timeouts := []time.Duration{5 * time.Minute, time.Hour, 30 * time.Second}
	for _, timeout := range timeouts {
		tt := newApiBuilerTest(t)
		tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
		want := expectedMachineHealthCheckForWorkers(timeout)
		tt.clusterSpec.Cluster.Spec.MachineHealthCheck = &v1alpha1.MachineHealthCheck{
			NodeStartupTimeout: &metav1.Duration{
				Duration: timeout,
			},
			UnhealthyMachineTimeout: &metav1.Duration{
				Duration: timeout,
			},
		}
		got := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec.Cluster)
		tt.Expect(got).To(Equal(want))
	}
}

func expectedMachineHealthCheckForWorkers(timeout time.Duration) []*clusterv1.MachineHealthCheck {
	maxUnhealthy := intstr.Parse("40%")
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
