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

var capiGitOpsResourceType = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)

func TestValidateGitopsForWorkloadClusters(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "SuccessNoGitops",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_gitops_response.json",
		},
		{
			name:               "FailureGitopsNameExists",
			wantErr:            errors.New("gitOpsConfig gitopstest already exists"),
			getClusterResponse: "testdata/gitops_name_exists.json",
		},
	}

	defaultGitOps := &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: v1alpha1.Flux{
				Github: v1alpha1.Github{
					Owner:               "owner",
					Repository:          "repo",
					FluxSystemNamespace: "flux-system",
					Branch:              "main",
					ClusterConfigPath:   "clusters/" + testclustername,
					Personal:            false,
				},
			},
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = testclustername
		s.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps
		s.SetManagedBy("management-cluster")
		// s.OIDCConfig = defaultOIDC
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(
				ctx, []string{
					"get", capiGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Return(*bytes.NewBufferString(fileContent), nil)

			err := createvalidations.ValidateGitOpsNameIsUnique(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestSkipValidateGitopsWithNoGitOpts(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "SuccessNoGitops",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_gitops_response.json",
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.SetManagedBy("management-cluster")
		s.Name = testclustername

		s.GitOpsConfig = nil
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

func TestValidateGitopsForSelfManagedCluster(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "Skip Validate GitOpsConfig name",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_gitops_response.json",
		},
	}

	defaultGitOps := &v1alpha1.GitOpsConfig{
		Spec: v1alpha1.GitOpsConfigSpec{
			Flux: v1alpha1.Flux{
				Github: v1alpha1.Github{
					Owner:               "owner",
					Repository:          "repo",
					FluxSystemNamespace: "flux-system",
					Branch:              "main",
					ClusterConfigPath:   "clusters/" + testclustername,
					Personal:            false,
				},
			},
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Name = testclustername
		s.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps

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
