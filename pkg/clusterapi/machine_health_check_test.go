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
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestMachineHealthCheckForControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	maxUnhealthy := intstr.Parse("100%")
	want := &clusterv1.MachineHealthCheck{
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
			UnhealthyConditions: []clusterv1.UnhealthyCondition{
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionUnknown,
					Timeout: metav1.Duration{Duration: 5 * time.Minute},
				},
				{
					Type:    corev1.NodeReady,
					Status:  corev1.ConditionFalse,
					Timeout: metav1.Duration{Duration: 5 * time.Minute},
				},
			},
		},
	}
	got := clusterapi.MachineHealthCheckForControlPlane(tt.clusterSpec)
	tt.Expect(got).To(Equal(want))
}

func TestMachineHealthCheckForWorkers(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations = []v1alpha1.WorkerNodeGroupConfiguration{*tt.workerNodeGroupConfig}
	maxUnhealthy := intstr.Parse("40%")
	want := map[string]*clusterv1.MachineHealthCheck{
		"wng-1": {
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
				MaxUnhealthy: &maxUnhealthy,
				UnhealthyConditions: []clusterv1.UnhealthyCondition{
					{
						Type:    corev1.NodeReady,
						Status:  corev1.ConditionUnknown,
						Timeout: metav1.Duration{Duration: 5 * time.Minute},
					},
					{
						Type:    corev1.NodeReady,
						Status:  corev1.ConditionFalse,
						Timeout: metav1.Duration{Duration: 5 * time.Minute},
					},
				},
			},
		},
	}

	got := clusterapi.MachineHealthCheckForWorkers(tt.clusterSpec)
	tt.Expect(got).To(Equal(want))
}
