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

var eksaGitOpsResourceType = fmt.Sprintf("gitopsconfigs.%s", v1alpha1.GroupVersion.Group)

func TestValidateGitopsForWorkloadClustersPath(t *testing.T) {
	tests := []struct {
		name    string
		wantErr error
		github  v1alpha1.Github
	}{
		{
			name:    "Success",
			wantErr: nil,
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/management-gitops",
				Personal:            false,
			},
		},
		{
			name:    "Failure, path diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, branch diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "dev",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, owner owner",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "janedoe",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, repo diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "diffrepo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, namespace diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "diff-ns",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            false,
			},
		},
		{
			name:    "Failure, personal diff",
			wantErr: errors.New("workload cluster gitOpsConfig is invalid: expected gitOpsConfig to be the same between management and its workload clusters"),
			github: v1alpha1.Github{
				Owner:               "owner",
				Repository:          "repo",
				FluxSystemNamespace: "flux-system",
				Branch:              "main",
				ClusterConfigPath:   "clusters/" + testclustername,
				Personal:            true,
			},
		},
	}

	gitOpsListContent := test.ReadFile(t, "testdata/empty_get_gitops_response.json")
	eksaClusterContent := test.ReadFile(t, "testdata/eksa_cluster_exists.json")
	mgmtGitOpsContent := test.ReadFile(t, "testdata/management_gitops_config.json")

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			defaultGitOps := &v1alpha1.GitOpsConfig{
				Spec: v1alpha1.GitOpsConfigSpec{
					Flux: v1alpha1.Flux{
						Github: tc.github,
					},
				},
			}

			clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster.Name = testclustername
				s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
					Kind: v1alpha1.GitOpsConfigKind,
					Name: "gitopstest",
				}
				s.GitOpsConfig = defaultGitOps
				s.Cluster.SetManagedBy("management-cluster")
			})
			k, ctx, cluster, e := validations.NewKubectl(t)
			cluster.Name = "management-cluster"

			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Return(*bytes.NewBufferString(gitOpsListContent), nil)

			e.EXPECT().Execute(
				ctx, []string{
					"get", "clusters.anywhere.eks.amazonaws.com", "-A", "-o", "jsonpath={.items[0]}", "--kubeconfig",
					cluster.KubeconfigFile,
					"--field-selector=metadata.name=management-cluster",
				}).Return(*bytes.NewBufferString(eksaClusterContent), nil)

			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "management-gitops", "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
				}).Return(*bytes.NewBufferString(mgmtGitOpsContent), nil)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

func TestValidateGitopsForWorkloadClustersFailure(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		getClusterResponse string
	}{
		{
			name:               "FailureGitOpsNameExists",
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
		s.Cluster.Name = testclustername
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps
		s.Cluster.SetManagedBy("management-cluster")
		// s.OIDCConfig = defaultOIDC
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Return(*bytes.NewBufferString(fileContent), nil)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec)
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
		getClusterResponse string
	}{
		{
			name:               "SuccessNoGitops",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_gitops_response.json",
		},
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.SetManagedBy("management-cluster")
		s.Cluster.Name = testclustername

		s.GitOpsConfig = nil
	})

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Times(0)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec)
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
		s.Cluster.Name = testclustername
		s.Cluster.Spec.GitOpsRef = &v1alpha1.Ref{
			Kind: v1alpha1.GitOpsConfigKind,
			Name: "gitopstest",
		}
		s.GitOpsConfig = defaultGitOps

		s.Cluster.SetSelfManaged()
	})
	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			e.EXPECT().Execute(
				ctx, []string{
					"get", eksaGitOpsResourceType, "-o", "json", "--kubeconfig",
					cluster.KubeconfigFile, "--namespace", clusterSpec.Cluster.Namespace,
					"--field-selector=metadata.name=gitopstest",
				}).Times(0)

			err := createvalidations.ValidateGitOps(ctx, k, cluster, clusterSpec)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
