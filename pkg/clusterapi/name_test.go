package clusterapi_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestIncrementName(t *testing.T) {
	tests := []struct {
		name    string
		oldName string
		want    string
		wantErr string
	}{
		{
			name:    "valid",
			oldName: "cluster-1",
			want:    "cluster-2",
			wantErr: "",
		},
		{
			name:    "invalid format",
			oldName: "cluster-1a",
			want:    "",
			wantErr: "invalid format of name",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got, err := clusterapi.IncrementName(tt.oldName)
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
				g.Expect(got).To(Equal(tt.want))
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestIncrementNameWithFallbackDefault(t *testing.T) {
	tests := []struct {
		name        string
		oldName     string
		defaultName string
		want        string
	}{
		{
			name:        "valid",
			oldName:     "cluster-1",
			defaultName: "default",
			want:        "cluster-2",
		},
		{
			name:        "invalid format",
			oldName:     "cluster-1a",
			defaultName: "default",
			want:        "default",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			got := clusterapi.IncrementNameWithFallbackDefault(tt.oldName, tt.defaultName)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestObjectName(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		version int
		want    string
	}{
		{
			name:    "cluster-1",
			base:    "cluster",
			version: 1,
			want:    "cluster-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(clusterapi.ObjectName(tt.base, tt.version)).To(Equal(tt.want))
		})
	}
}

func TestDefaultObjectName(t *testing.T) {
	tests := []struct {
		name string
		base string
		want string
	}{
		{
			name: "cluster-1",
			base: "cluster",
			want: "cluster-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(clusterapi.DefaultObjectName(tt.base)).To(Equal(tt.want))
		})
	}
}

func TestKubeadmControlPlaneName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test cluster",
			want: "test-cluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.KubeadmControlPlaneName(g.clusterSpec.Cluster)).To(Equal(tt.want))
		})
	}
}

func TestMachineDeploymentName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "wng 1",
			want: "test-cluster-wng-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.MachineDeploymentName(g.clusterSpec.Cluster, *g.workerNodeGroupConfig)).To(Equal(tt.want))
		})
	}
}

func TestDefaultKubeadmConfigTemplateName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "wng 1",
			want: "test-cluster-wng-1-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.DefaultKubeadmConfigTemplateName(g.clusterSpec, *g.workerNodeGroupConfig)).To(Equal(tt.want))
		})
	}
}

func TestControlPlaneMachineTemplateName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test cluster",
			want: "test-cluster-control-plane-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.ControlPlaneMachineTemplateName(g.clusterSpec.Cluster)).To(Equal(tt.want))
		})
	}
}

func TestEtcdMachineTemplateName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "test cluster",
			want: "test-cluster-etcd-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.EtcdMachineTemplateName(g.clusterSpec.Cluster)).To(Equal(tt.want))
		})
	}
}

func TestWorkerMachineTemplateName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "wng 1",
			want: "test-cluster-wng-1-1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.WorkerMachineTemplateName(g.clusterSpec, *g.workerNodeGroupConfig)).To(Equal(tt.want))
		})
	}
}

func TestControlPlaneMachineHealthCheckName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "cp",
			want: "test-cluster-kcp-unhealthy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.ControlPlaneMachineHealthCheckName(g.clusterSpec.Cluster)).To(Equal(tt.want))
		})
	}
}

func TestWorkerMachineHealthCheckName(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{
			name: "wng 1",
			want: "test-cluster-wng-1-worker-unhealthy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.Expect(clusterapi.WorkerMachineHealthCheckName(g.clusterSpec.Cluster, *g.workerNodeGroupConfig)).To(Equal(tt.want))
		})
	}
}

func TestEnsureNewNameIfChangedObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	originalName := "my-machine-template-1"
	mt := dockerMachineTemplate()
	mt.Name = originalName
	client := test.NewFakeKubeClient()

	g.Expect(clusterapi.EnsureNewNameIfChanged(ctx, client, notFoundRetriever, withChangesCompare, mt)).To(Succeed())
	g.Expect(mt.Name).To(Equal(originalName))
}

func TestEnsureNewNameIfChangedErrorReadingObject(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mt := dockerMachineTemplate()
	mt.Name = "my-machine-template"
	client := test.NewFakeKubeClient()

	g.Expect(
		clusterapi.EnsureNewNameIfChanged(ctx, client, errorRetriever, withChangesCompare, mt),
	).To(
		MatchError(ContainSubstring("reading DockerMachineTemplate eksa-system/my-machine-template from API")),
	)
}

func TestEnsureNewNameIfChangedErrorIncrementingName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mt := dockerMachineTemplate()
	mt.Name = "my-machine-template"
	client := test.NewFakeKubeClient()

	g.Expect(
		clusterapi.EnsureNewNameIfChanged(ctx, client, dummyRetriever, withChangesCompare, mt),
	).To(
		MatchError(ContainSubstring("incrementing name for DockerMachineTemplate eksa-system/my-machine-template")),
	)
}

func TestEnsureNewNameIfChangedObjectNeedsNewName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mt := dockerMachineTemplate()
	mt.Name = "my-machine-template-1"
	client := test.NewFakeKubeClient()

	g.Expect(clusterapi.EnsureNewNameIfChanged(ctx, client, dummyRetriever, withChangesCompare, mt)).To(Succeed())
	g.Expect(mt.Name).To(Equal("my-machine-template-2"))
}

func TestEnsureNewNameIfChangedObjectHasNotChanged(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	originalName := "my-machine-template-1"
	mt := dockerMachineTemplate()
	mt.Name = originalName
	client := test.NewFakeKubeClient()

	g.Expect(clusterapi.EnsureNewNameIfChanged(ctx, client, dummyRetriever, noChangesCompare, mt)).To(Succeed())
	g.Expect(mt.Name).To(Equal(originalName))
}

func TestClusterCASecretName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(clusterapi.ClusterCASecretName("my-cluster")).To(Equal("my-cluster-ca"))
}

func TestClusterKubeconfigSecretName(t *testing.T) {
	g := NewWithT(t)
	g.Expect(clusterapi.ClusterKubeconfigSecretName("my-cluster")).To(Equal("my-cluster-kubeconfig"))
}

func TestInitialTemplateNamesForWorkers(t *testing.T) {
	tests := []struct {
		name         string
		wantTNames   map[string]string
		wantKCTNames map[string]string
	}{
		{
			name: "wng 1",
			wantTNames: map[string]string{
				"wng-1": "test-cluster-wng-1-1",
			},
			wantKCTNames: map[string]string{
				"wng-1": "test-cluster-wng-1-1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			spec := g.clusterSpec.DeepCopy()
			spec.Cluster.Spec.WorkerNodeGroupConfigurations = append(spec.Cluster.Spec.WorkerNodeGroupConfigurations, *g.workerNodeGroupConfig)
			workloadTemplateNames, kubeadmConfigTemplateNames := clusterapi.InitialTemplateNamesForWorkers(spec)
			g.Expect(workloadTemplateNames).To(Equal(tt.wantTNames))
			g.Expect(kubeadmConfigTemplateNames).To(Equal(tt.wantKCTNames))
		})
	}
}

func dummyRetriever(_ context.Context, _ kubernetes.Client, _, _ string) (*dockerv1.DockerMachineTemplate, error) {
	return dockerMachineTemplate(), nil
}

func errorRetriever(_ context.Context, _ kubernetes.Client, _, _ string) (*dockerv1.DockerMachineTemplate, error) {
	return nil, errors.New("reading object")
}

func notFoundRetriever(_ context.Context, _ kubernetes.Client, _, _ string) (*dockerv1.DockerMachineTemplate, error) {
	return nil, apierrors.NewNotFound(schema.GroupResource{}, "")
}

func noChangesCompare(_, _ *dockerv1.DockerMachineTemplate) bool {
	return true
}

func withChangesCompare(_, _ *dockerv1.DockerMachineTemplate) bool {
	return false
}
