package upgradevalidations_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidateVersionSkew(t *testing.T) {
	tests := []struct {
		name           string
		wantErr        error
		upgradeVersion anywherev1.KubernetesVersion
		oldVersion     anywherev1.KubernetesVersion
	}{
		{
			name:           "FailureTwoMinorVersions",
			wantErr:        fmt.Errorf("only +1 minor version skew is supported"),
			upgradeVersion: anywherev1.Kube120,
			oldVersion:     anywherev1.Kube118,
		},
		{
			name:           "FailureMinusOneMinorVersion",
			wantErr:        fmt.Errorf("kubernetes version downgrade is not supported (%s) -> (%s)", anywherev1.Kube120, anywherev1.Kube119),
			upgradeVersion: anywherev1.Kube119,
			oldVersion:     anywherev1.Kube120,
		},
		{
			name:           "SuccessSameVersion",
			wantErr:        nil,
			upgradeVersion: anywherev1.Kube119,
			oldVersion:     anywherev1.Kube119,
		},
		{
			name:           "SuccessOneMinorVersion",
			wantErr:        nil,
			upgradeVersion: anywherev1.Kube120,
			oldVersion:     anywherev1.Kube119,
		},
	}

	mockCtrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(mockCtrl)
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			newCluster := baseCluster()
			newCluster.Spec.KubernetesVersion = tc.upgradeVersion

			oldCluster := baseCluster()
			oldCluster.Spec.KubernetesVersion = tc.oldVersion

			cluster := &types.Cluster{KubeconfigFile: "test.kubeconfig"}

			k.EXPECT().GetEksaCluster(ctx, cluster, newCluster.Name).Return(oldCluster, nil)

			err := upgradevalidations.ValidateServerVersionSkew(ctx, newCluster, cluster, cluster, k)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateWorkerVersionSkew(t *testing.T) {
	kube119 := anywherev1.KubernetesVersion("1.19")
	kube121 := anywherev1.KubernetesVersion("1.21")
	tests := []struct {
		name                 string
		wantErr              error
		upgradeVersion       anywherev1.KubernetesVersion
		oldVersion           anywherev1.KubernetesVersion
		upgradeWorkerVersion *anywherev1.KubernetesVersion
		oldWorkerVersion     *anywherev1.KubernetesVersion
	}{
		{
			name:                 "FailureTwoMinorVersions",
			wantErr:              fmt.Errorf("only +1 minor version skew is supported"),
			upgradeVersion:       anywherev1.Kube121,
			oldVersion:           anywherev1.Kube121,
			upgradeWorkerVersion: &kube121,
			oldWorkerVersion:     &kube119,
		},
	}

	mockCtrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(mockCtrl)
	ctx := context.Background()

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			newCluster := baseCluster()
			newCluster.Spec.KubernetesVersion = tc.upgradeVersion
			newCluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = tc.upgradeWorkerVersion

			oldCluster := baseCluster()
			oldCluster.Spec.KubernetesVersion = tc.oldVersion
			oldCluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = tc.oldWorkerVersion

			cluster := &types.Cluster{KubeconfigFile: "test.kubeconfig"}

			k.EXPECT().GetEksaCluster(ctx, cluster, newCluster.Name).Return(oldCluster, nil)

			err := upgradevalidations.ValidateWorkerServerVersionSkew(ctx, newCluster, cluster, cluster, k)
			if err != nil && !strings.Contains(err.Error(), tc.wantErr.Error()) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func baseCluster() *anywherev1.Cluster {
	c := &anywherev1.Cluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       anywherev1.ClusterKind,
			APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "mgmt",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube121,
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 3,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:       "bundles-1",
				Namespace:  constants.EksaSystemNamespace,
				APIVersion: anywherev1.SchemeBuilder.GroupVersion.String(),
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{{
				Name:  "md-0",
				Count: ptr.Int(1),
				MachineGroupRef: &anywherev1.Ref{
					Kind: anywherev1.VSphereMachineConfigKind,
					Name: "eksa-unit-test",
				},
			}},
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{Cilium: &anywherev1.CiliumConfig{}},
				Pods: anywherev1.Pods{
					CidrBlocks: []string{"192.168.0.0/16"},
				},
				Services: anywherev1.Services{
					CidrBlocks: []string{"10.96.0.0/12"},
				},
			},
			DatacenterRef: anywherev1.Ref{
				Kind: anywherev1.VSphereDatacenterKind,
				Name: "eksa-unit-test",
			},
		},
	}

	return c
}
