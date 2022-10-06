package clusterapi_test

import (
	"context"
	"errors"
	"testing"

	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

type dockerControlPlane = clusterapi.ControlPlane[*dockerv1.DockerCluster, *dockerv1.DockerMachineTemplate]

func TestControlPlaneObjects(t *testing.T) {
	tests := []struct {
		name         string
		controlPlane *dockerControlPlane
		want         []clusterapi.Object
	}{
		{
			name: "stacked etcd",
			controlPlane: &dockerControlPlane{
				Cluster:                     capiCluster(),
				ProviderCluster:             dockerCluster(),
				KubeadmControlPlane:         kubeadmControlPlane(),
				ControlPlaneMachineTemplate: dockerMachineTemplate(),
			},
			want: []clusterapi.Object{
				capiCluster(),
				dockerCluster(),
				kubeadmControlPlane(),
				dockerMachineTemplate(),
			},
		},
		{
			name: "unstacked etcd",
			controlPlane: &dockerControlPlane{
				Cluster:                     capiCluster(),
				ProviderCluster:             dockerCluster(),
				KubeadmControlPlane:         kubeadmControlPlane(),
				ControlPlaneMachineTemplate: dockerMachineTemplate(),
				EtcdCluster:                 etcdCluster(),
				EtcdMachineTemplate:         dockerMachineTemplate(),
			},
			want: []clusterapi.Object{
				capiCluster(),
				dockerCluster(),
				kubeadmControlPlane(),
				dockerMachineTemplate(),
				etcdCluster(),
				dockerMachineTemplate(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.controlPlane.Objects()).To(ConsistOf(tt.want))
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

func TestControlPlaneUpdateImmutableObjectNamesNoKubeadmControlPlane(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := test.NewFakeKubeClient()
	cp := controlPlaneStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName

	g.Expect(cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare)).To(Succeed())
	g.Expect(cp.ControlPlaneMachineTemplate.Name).To(Equal(originalCPMachineTemplateName))
}

func TestControlPlaneUpdateImmutableObjectNamesErrorReadingControlPlane(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneStackedEtcd()
	client := test.NewFakeKubeClientAlwaysError()

	g.Expect(
		cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).To(
		MatchError(ContainSubstring("reading current kubeadm control plane from API")),
	)
}

func TestControlPlaneUpdateImmutableObjectNamesErrorUpdatingName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(cp.Objects())...)

	g.Expect(
		cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, withChangesCompare),
	).To(
		MatchError(ContainSubstring("incrementing name for DockerMachineTemplate eksa-system/my-machine-template")),
	)
}

func TestControlPlaneUpdateImmutableObjectNamesSuccessStackedEtcd(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(cp.Objects())...)

	g.Expect(cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare)).To(Succeed())
}

func TestControlPlaneUpdateImmutableObjectNamesNoEtcdCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(cp.Objects())...)
	cp.EtcdCluster = etcdCluster()

	g.Expect(cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare)).To(Succeed())
}

func TestControlPlaneUpdateImmutableObjectNamesErrorReadingEtcdCluster(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneUnStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	scheme := runtime.NewScheme()
	g.Expect(controlplanev1.AddToScheme(scheme)).To(Succeed())
	client := test.NewKubeClient(
		fake.NewClientBuilder().WithScheme(scheme).WithObjects(cp.KubeadmControlPlane).Build(),
	)

	g.Expect(
		cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare),
	).To(
		MatchError(ContainSubstring("reading current etcdadm cluster from API")),
	)
}

func TestControlPlaneUpdateImmutableObjectNamesErrorUpdatingEtcdName(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneUnStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	originalEtcdMachineTemplateName := "my-etcd-machine-template"
	cp.EtcdMachineTemplate.Name = originalEtcdMachineTemplateName
	cp.EtcdCluster.Spec.InfrastructureTemplate.Name = originalEtcdMachineTemplateName
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(cp.Objects())...)

	g.Expect(
		cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, withChangesCompare),
	).To(
		MatchError(ContainSubstring("incrementing name for DockerMachineTemplate eksa-system/my-etcd-machine-template")),
	)
}

func TestControlPlaneUpdateImmutableObjectNamesSuccessUnstackedEtcd(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cp := controlPlaneUnStackedEtcd()
	originalCPMachineTemplateName := "my-machine-template-1"
	cp.ControlPlaneMachineTemplate.Name = originalCPMachineTemplateName
	cp.KubeadmControlPlane.Spec.MachineTemplate.InfrastructureRef.Name = originalCPMachineTemplateName
	originalEtcdMachineTemplateName := "my-etcd-machine-template-2"
	cp.EtcdMachineTemplate.Name = originalEtcdMachineTemplateName
	cp.EtcdCluster.Spec.InfrastructureTemplate.Name = originalEtcdMachineTemplateName
	client := test.NewFakeKubeClient(clientutil.ObjectsToClientObjects(cp.Objects())...)

	g.Expect(cp.UpdateImmutableObjectNames(ctx, client, dummyRetriever, noChangesCompare)).To(Succeed())
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

func capiCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{}
}

func dockerCluster() *dockerv1.DockerCluster {
	return &dockerv1.DockerCluster{}
}

func kubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{}
}

func dockerMachineTemplate() *dockerv1.DockerMachineTemplate {
	return &dockerv1.DockerMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			Kind: "DockerMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.EksaSystemNamespace,
		},
	}
}

func etcdCluster() *etcdv1.EtcdadmCluster {
	return &etcdv1.EtcdadmCluster{}
}

func controlPlaneStackedEtcd() *dockerControlPlane {
	return &dockerControlPlane{
		Cluster:                     capiCluster(),
		ProviderCluster:             dockerCluster(),
		KubeadmControlPlane:         kubeadmControlPlane(),
		ControlPlaneMachineTemplate: dockerMachineTemplate(),
	}
}

func controlPlaneUnStackedEtcd() *dockerControlPlane {
	cp := controlPlaneStackedEtcd()
	cp.EtcdCluster = etcdCluster()
	cp.EtcdMachineTemplate = dockerMachineTemplate()

	return cp
}
