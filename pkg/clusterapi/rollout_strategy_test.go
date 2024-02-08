package clusterapi_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/intstr"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestSetUpgradeRolloutStrategyInKubeadmControlPlane(t *testing.T) {
	tests := []struct {
		name            string
		rolloutStrategy *anywherev1.ControlPlaneUpgradeRolloutStrategy
		want            *controlplanev1.KubeadmControlPlane
	}{
		{
			name:            "no upgrade rollout strategy",
			rolloutStrategy: nil,
			want:            wantKubeadmControlPlane(),
		},
		{
			name: "with maxSurge",
			rolloutStrategy: &anywherev1.ControlPlaneUpgradeRolloutStrategy{
				RollingUpdate: &anywherev1.ControlPlaneRollingUpdateParams{
					MaxSurge: 1,
				},
			},
			want: wantKubeadmControlPlane(func(k *controlplanev1.KubeadmControlPlane) {
				maxSurge := intstr.FromInt(1)
				k.Spec.RolloutStrategy = &controlplanev1.RolloutStrategy{
					Type: controlplanev1.RollingUpdateStrategyType,
					RollingUpdate: &controlplanev1.RollingUpdate{
						MaxSurge: &maxSurge,
					},
				}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kcp := wantKubeadmControlPlane()
			clusterapi.SetUpgradeRolloutStrategyInKubeadmControlPlane(kcp, tt.rolloutStrategy)
			assert.Equal(t, tt.want, kcp)
		})
	}
}

func TestSetUpgradeRolloutStrategyInMachineDeployment(t *testing.T) {
	tests := []struct {
		name            string
		rolloutStrategy *anywherev1.WorkerNodesUpgradeRolloutStrategy
		want            *clusterv1.MachineDeployment
	}{
		{
			name:            "no upgrade rollout strategy",
			rolloutStrategy: nil,
			want:            wantMachineDeployment(),
		},
		{
			name: "with maxSurge and maxUnavailable",
			rolloutStrategy: &anywherev1.WorkerNodesUpgradeRolloutStrategy{
				RollingUpdate: &anywherev1.WorkerNodesRollingUpdateParams{
					MaxSurge:       1,
					MaxUnavailable: 0,
				},
			},
			want: wantMachineDeployment(func(m *clusterv1.MachineDeployment) {
				maxSurge := intstr.FromInt(1)
				maxUnavailable := intstr.FromInt(0)
				m.Spec.Strategy = &clusterv1.MachineDeploymentStrategy{
					Type: clusterv1.RollingUpdateMachineDeploymentStrategyType,
					RollingUpdate: &clusterv1.MachineRollingUpdateDeployment{
						MaxSurge:       &maxSurge,
						MaxUnavailable: &maxUnavailable,
					},
				}
			}),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := wantMachineDeployment()
			clusterapi.SetUpgradeRolloutStrategyInMachineDeployment(md, tt.rolloutStrategy)
			assert.Equal(t, tt.want, md)
		})
	}
}
