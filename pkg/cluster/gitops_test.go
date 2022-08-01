package cluster_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
)

const (
	owner             = "janedoe"
	repository        = "flux-fleet"
	fluxNamespace     = "test-ns"
	branch            = "test-branch"
	clusterConfigPath = "test-path"
	personal          = false
)

func TestGitOpsToFluxConversionProcessing(t *testing.T) {
	tests := []struct {
		name           string
		wantConfigPath string
		wantFluxSpec   anywherev1.FluxConfigSpec
	}{
		{
			name:           "workload cluster with GitOpsConfig",
			wantConfigPath: "testdata/cluster_gitops_1_21.yaml",
			wantFluxSpec: anywherev1.FluxConfigSpec{
				SystemNamespace:   fluxNamespace,
				ClusterConfigPath: clusterConfigPath,
				Branch:            branch,
				Github: &anywherev1.GithubProviderConfig{
					Owner:      owner,
					Repository: repository,
					Personal:   personal,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			config, err := cluster.ParseConfigFromFile(tt.wantConfigPath)
			if err != nil {
				t.Fatal("cluster.ParseConfigFromFile error != nil, want nil", err)
			}
			g.Expect(config.FluxConfig.Spec).To(Equal(tt.wantFluxSpec))
		})
	}
}

func TestDefaultConfigClientBuilderGitOpsConfig(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			GitOpsRef: &anywherev1.Ref{
				Kind: anywherev1.GitOpsConfigKind,
				Name: "my-gitops",
			},
		},
	}
	gitopsConfig := &anywherev1.GitOpsConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-gitops",
			Namespace: "default",
		},
	}

	client.EXPECT().Get(ctx, "my-gitops", "default", &anywherev1.GitOpsConfig{}).DoAndReturn(
		func(ctx context.Context, name, namespace string, obj runtime.Object) error {
			c := obj.(*anywherev1.GitOpsConfig)
			c.ObjectMeta = gitopsConfig.ObjectMeta
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(config.GitOpsConfig).To(Equal(gitopsConfig))
}
