package createvalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

var vsphereDatacenterResourceType = fmt.Sprintf("vspheredatacenterconfigs.%s", v1alpha1.GroupVersion.Group)

func TestValidateVSphereDatacenterForWorkloadClusters(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "SuccessNoVSphereDatacenter",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_vspheredatacenter_response.json",
		},
		{
			name:               "FailureVSphereDatacenterNameExists",
			wantErr:            errors.New("VSphereDatacenter vspheredatacenter already exists"),
			getClusterResponse: "testdata/vspheredatacenter_name_exists.json",
		},
	}

	defaultDatacenter := &v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "datacenter",
			Network:    "network",
			Server:     "server.com",
			Thumbprint: "test",
			Insecure:   false,
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		nb := false
		s.Spec.Management = &nb
		s.Name = testclustername
		s.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vspheredatacenter",
		}
		s.DatacenterConfig = &defaultDatacenter.ObjectMeta
		s.SetManagedBy("management-cluster")
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(
				ctx, []string{
					"get", vsphereDatacenterResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Namespace,
					"--field-selector=metadata.name=vspheredatacenter",
				}).Return(*bytes.NewBufferString(fileContent), nil)

			err := createvalidations.ValidateDatacenterNameIsUnique(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateVSphereDatacenterConfigForSelfManagedCluster(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "Skip Validate VSpheredatacenterConfig name",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_vspheredatacenter_response.json",
		},
	}

	defaultDatacenter := &v1alpha1.VSphereDatacenterConfig{
		Spec: v1alpha1.VSphereDatacenterConfigSpec{
			Datacenter: "datacenter",
			Network:    "network",
			Server:     "server.com",
			Thumbprint: "test",
			Insecure:   false,
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		nb := false
		s.Spec.Management = &nb
		s.Name = testclustername
		s.Spec.DatacenterRef = v1alpha1.Ref{
			Kind: v1alpha1.VSphereDatacenterKind,
			Name: "vspheredatacenter",
		}
		s.DatacenterConfig = &defaultDatacenter.ObjectMeta

		s.SetSelfManaged()
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", capiGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Times(0)

			err := createvalidations.ValidateGitOpsNameIsUnique(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
