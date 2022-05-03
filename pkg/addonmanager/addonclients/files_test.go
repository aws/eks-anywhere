package addonclients_test

import (
	"context"
	"io/ioutil"
	"path"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

type filesTest struct {
	*WithT
	ctx         context.Context
	currentSpec *cluster.Spec
	newSpec     *cluster.Spec
	fluxConfig  v1alpha1.FluxConfig
}

func newFilesTest(t *testing.T) *filesTest {
	fluxConfigSpec := v1alpha1.FluxConfigSpec{
		SystemNamespace:   "flux-system",
		ClusterConfigPath: "clusters/management-cluster",
		Branch:            "testBranch",
		Github: &v1alpha1.GithubProviderConfig{
			Owner:      "mFowler",
			Repository: "testRepo",
			Personal:   true,
		},
	}
	currentSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &v1alpha1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "management-cluster",
			},
		}
		s.FluxConfig = &v1alpha1.FluxConfig{
			Spec: fluxConfigSpec,
		}
	})

	return &filesTest{
		WithT:       NewWithT(t),
		ctx:         context.Background(),
		currentSpec: currentSpec,
		newSpec:     currentSpec.DeepCopy(),
		fluxConfig:  v1alpha1.FluxConfig{Spec: fluxConfigSpec},
	}
}

func TestUpdateLegacyFileStructureNoGitOpsConfig(t *testing.T) {
	tt := newFilesTest(t)
	f, _, _ := newAddonClient(t)
	tt.newSpec.FluxConfig = nil

	tt.Expect(f.UpdateLegacyFileStructure(tt.ctx, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpdateLegacyFileStructureNoChanges(t *testing.T) {
	tt := newFilesTest(t)
	f, m, g := newAddonClient(t)
	_, err := g.Writer.WithDir("clusters/management-cluster/flux-system")
	if err != nil {
		t.Errorf("failed to create test flux-system directory: %v", err)
	}
	_, err = g.Writer.WithDir("clusters/management-cluster/management-cluster/eksa-system")
	if err != nil {
		t.Errorf("failed to create test eksa-system directory: %v", err)
	}

	m.gitClient.EXPECT().Clone(tt.ctx).Return(nil)
	m.gitClient.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)

	tt.Expect(f.UpdateLegacyFileStructure(tt.ctx, tt.currentSpec, tt.newSpec)).To(BeNil())
}

func TestUpdateLegacyFileStructureSuccess(t *testing.T) {
	tt := newFilesTest(t)
	f, m, g := newAddonClient(t)
	_, err := g.Writer.WithDir("clusters/management-cluster/flux-system")
	if err != nil {
		t.Errorf("failed to create test flux-system directory: %v", err)
	}
	w, err := g.Writer.WithDir("clusters/management-cluster/eksa-system")
	if err != nil {
		t.Errorf("failed to create test eksa-system directory: %v", err)
	}
	eksaContent, err := ioutil.ReadFile("./testdata/cluster-config-default-path-management.yaml")
	if err != nil {
		t.Fatalf("File [%s] reading error in test: %v", "cluster-config-default-path-management.yaml", err)
	}
	_, err = w.Write(defaultEksaClusterConfigFileName, eksaContent, filewriter.PersistentFile)
	if err != nil {
		t.Fatalf("failed to write eksa-cluster.yaml in test: %v", err)
	}
	kustomizationContent, err := ioutil.ReadFile("./testdata/kustomization.yaml")
	if err != nil {
		t.Fatalf("File [%s] reading error in test: %v", "kustomization.yaml", err)
	}
	_, err = w.Write(defaultKustomizationManifestFileName, kustomizationContent, filewriter.PersistentFile)
	if err != nil {
		t.Fatalf("failed to write kustomization.yaml in test: %v", err)
	}

	m.gitClient.EXPECT().Clone(tt.ctx).Return(nil)
	m.gitClient.EXPECT().Branch(tt.fluxConfig.Spec.Branch).Return(nil)
	m.gitClient.EXPECT().Add(tt.fluxConfig.Spec.ClusterConfigPath).Return(nil)
	m.gitClient.EXPECT().Commit(test.OfType("string")).Return(nil)
	m.gitClient.EXPECT().Push(tt.ctx).Return(nil)
	m.gitClient.EXPECT().Remove("clusters/management-cluster/eksa-system").Return(nil)

	tt.Expect(f.UpdateLegacyFileStructure(tt.ctx, tt.currentSpec, tt.newSpec)).To(BeNil())

	expectedEksaClusterConfigPath := path.Join(g.Writer.Dir(), tt.fluxConfig.Spec.ClusterConfigPath, tt.newSpec.Cluster.GetClusterName(), "eksa-system", defaultEksaClusterConfigFileName)
	test.AssertFilesEquals(t, expectedEksaClusterConfigPath, "./testdata/cluster-config-default-path-management.yaml")

	expectedEksaKustomizationPath := path.Join(g.Writer.Dir(), tt.fluxConfig.Spec.ClusterConfigPath, tt.newSpec.Cluster.GetClusterName(), "eksa-system", defaultKustomizationManifestFileName)
	test.AssertFilesEquals(t, expectedEksaKustomizationPath, "./testdata/kustomization.yaml")
}
