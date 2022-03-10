package v1alpha1

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSetClusterDefaults(t *testing.T) {
	tests := []struct {
		name            string
		in, wantCluster *Cluster
		wantErr         string
	}{
		{
			name: "worker node group - no name specified",
			in: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{
							Cilium: &CiliumConfig{},
						},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ControlPlaneConfiguration: ControlPlaneConfiguration{
						Count: 3,
						Endpoint: &Endpoint{
							Host: "test-ip",
						},
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					},
					WorkerNodeGroupConfigurations: []WorkerNodeGroupConfiguration{{
						Name:  "md-0",
						Count: 3,
						MachineGroupRef: &Ref{
							Kind: VSphereMachineConfigKind,
							Name: "eksa-unit-test",
						},
					}},
					DatacenterRef: Ref{
						Kind: VSphereDatacenterKind,
						Name: "eksa-unit-test",
					},
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{
							Cilium: &CiliumConfig{},
						},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: "",
		},
		{
			name: "cni plugin - old format in input, set new format",
			in: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ClusterNetwork: ClusterNetwork{
						CNI: Cilium,
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantCluster: &Cluster{
				TypeMeta: metav1.TypeMeta{
					Kind:       ClusterKind,
					APIVersion: SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: ClusterSpec{
					KubernetesVersion: Kube119,
					ClusterNetwork: ClusterNetwork{
						CNIConfig: &CNIConfig{
							Cilium: &CiliumConfig{},
						},
						Pods: Pods{
							CidrBlocks: []string{"192.168.0.0/16"},
						},
						Services: Services{
							CidrBlocks: []string{"10.96.0.0/12"},
						},
					},
				},
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			gotErr := setClusterDefaults(tt.in)
			if tt.wantErr == "" {
				g.Expect(gotErr).To(BeNil())
			} else {
				g.Expect(gotErr).To(MatchError(ContainSubstring(tt.wantErr)))
			}

			g.Expect(tt.in).To(Equal(tt.wantCluster))
		})
	}
}
