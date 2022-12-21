package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestConfigureAutoscalingInMachineDeployment(t *testing.T) {
	replicas := int32(3)
	version := "v1.21.5-eks-1-21-9"
	tests := []struct {
		name              string
		autoscalingConfig *v1alpha1.AutoScalingConfiguration
		want              *clusterv1.MachineDeployment
	}{
		{
			name:              "no autoscaling config",
			autoscalingConfig: nil,
			want:              wantMachineDeployment(),
		},
		{
			name: "with autoscaling config",
			autoscalingConfig: &v1alpha1.AutoScalingConfiguration{
				MinCount: 1,
				MaxCount: 3,
			},
			want: &clusterv1.MachineDeployment{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "cluster.x-k8s.io/v1beta1",
					Kind:       "MachineDeployment",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster-wng-1",
					Namespace: "eksa-system",
					Labels: map[string]string{
						"cluster.x-k8s.io/cluster-name":                        "test-cluster",
						"cluster.anywhere.eks.amazonaws.com/cluster-name":      "test-cluster",
						"cluster.anywhere.eks.amazonaws.com/cluster-namespace": "my-namespace",
					},
					Annotations: map[string]string{
						"cluster.x-k8s.io/cluster-api-autoscaler-node-group-min-size": "1",
						"cluster.x-k8s.io/cluster-api-autoscaler-node-group-max-size": "3",
					},
				},
				Spec: clusterv1.MachineDeploymentSpec{
					ClusterName: "test-cluster",
					Selector: metav1.LabelSelector{
						MatchLabels: map[string]string{},
					},
					Template: clusterv1.MachineTemplateSpec{
						ObjectMeta: clusterv1.ObjectMeta{
							Labels: map[string]string{
								"cluster.x-k8s.io/cluster-name": "test-cluster",
							},
						},
						Spec: clusterv1.MachineSpec{
							Bootstrap: clusterv1.Bootstrap{
								ConfigRef: &v1.ObjectReference{
									APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
									Kind:       "KubeadmConfigTemplate",
									Name:       "md-0",
								},
							},
							ClusterName: "test-cluster",
							InfrastructureRef: v1.ObjectReference{
								APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
								Kind:       "ProviderMachineTemplate",
								Name:       "provider-template",
							},
							Version: &version,
						},
					},
					Replicas: &replicas,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantMachineDeployment()
			clusterapi.ConfigureAutoscalingInMachineDeployment(got, tt.autoscalingConfig)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
