package snow_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/api/node/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/snow"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

func TestControlPlaneObjects(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.KubeadmControlPlaneName(g.clusterSpec),
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *controlplanev1.KubeadmControlPlane) error {
			obj.Spec.MachineTemplate.InfrastructureRef.Name = "test-cp-1"
			return nil
		})
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"test-cp-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *snowv1.AWSSnowMachineTemplate) error {
			mt.DeepCopyInto(obj)
			obj.SetName("test-cp-1")
			obj.Spec.Template.Spec.InstanceType = "updated-instance-type"
			return nil
		})

	wantMachineTemplateName := "test-cp-2"
	mt.SetName(wantMachineTemplateName)
	mt.Spec.Template.Spec.InstanceType = "sbe-c.large"
	kcp := wantKubeadmControlPlane()
	kcp.Spec.MachineTemplate.InfrastructureRef.Name = wantMachineTemplateName

	got, err := snow.ControlPlaneObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{wantCAPICluster(), wantSnowCluster(), kcp, mt}))
}

func TestControlPlaneObjectsOldControlPlaneNotExists(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.KubeadmControlPlaneName(g.clusterSpec),
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		Return(apierrors.NewNotFound(v1alpha1.Resource("foo"), "kind not found"))

	mt.SetName("test-cp-1")
	mt.Spec.Template.Spec.InstanceType = "sbe-c.large"

	got, err := snow.ControlPlaneObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{wantCAPICluster(), wantSnowCluster(), wantKubeadmControlPlane(), mt}))
}

func TestControlPlaneObjectsOldMachineTemplateNotExists(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.KubeadmControlPlaneName(g.clusterSpec),
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *controlplanev1.KubeadmControlPlane) error {
			obj.Spec.MachineTemplate.InfrastructureRef.Name = "test-cp-1"
			return nil
		})
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"test-cp-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		Return(apierrors.NewNotFound(v1alpha1.Resource("foo"), "kind not found"))

	mt.SetName("test-cp-1")
	mt.Spec.Template.Spec.InstanceType = "sbe-c.large"

	got, err := snow.ControlPlaneObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{wantCAPICluster(), wantSnowCluster(), wantKubeadmControlPlane(), mt}))
}

func TestControlPlaneObjectsGetOldControlPlaneError(t *testing.T) {
	g := newSnowTest(t)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.KubeadmControlPlaneName(g.clusterSpec),
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		Return(errors.New("get cp error"))

	_, err := snow.ControlPlaneObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).NotTo(Succeed())
}

func TestControlPlaneObjectsGetOldMachineTemplateError(t *testing.T) {
	g := newSnowTest(t)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.KubeadmControlPlaneName(g.clusterSpec),
			constants.EksaSystemNamespace,
			&controlplanev1.KubeadmControlPlane{},
		).
		Return(nil)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		Return(errors.New("get mt error"))

	_, err := snow.ControlPlaneObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).NotTo(Succeed())
}

func TestWorkersObjects(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.MachineDeploymentName(g.clusterSpec, g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]),
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *clusterv1.MachineDeployment) error {
			obj.Spec.Template.Spec.InfrastructureRef.Name = "test-wn-1"
			return nil
		})
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"test-wn-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *snowv1.AWSSnowMachineTemplate) error {
			mt.DeepCopyInto(obj)
			obj.SetName("test-wn-1")
			obj.Spec.Template.Spec.InstanceType = "updated-instance-type"
			return nil
		})

	wantMachineTemplateName := "test-wn-2"
	mt.SetName(wantMachineTemplateName)
	md := wantMachineDeployment()
	md.Spec.Template.Spec.InfrastructureRef.Name = wantMachineTemplateName

	got, err := snow.WorkersObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{md, wantKubeadmConfigTemplate(), mt}))
}

func TestWorkersObjectsOldMachineDeploymentNotExists(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.MachineDeploymentName(g.clusterSpec, g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]),
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(apierrors.NewNotFound(v1alpha1.Resource("foo"), "kind not found"))

	got, err := snow.WorkersObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{wantMachineDeployment(), wantKubeadmConfigTemplate(), mt}))
}

func TestWorkersObjectsOldMachineTemplateNotExists(t *testing.T) {
	g := newSnowTest(t)
	mt := wantSnowMachineTemplate()
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.MachineDeploymentName(g.clusterSpec, g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]),
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *clusterv1.MachineDeployment) error {
			obj.Spec.Template.Spec.InfrastructureRef.Name = "test-wn-1"
			return nil
		})
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"test-wn-1",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		Return(apierrors.NewNotFound(v1alpha1.Resource("foo"), "kind not found"))

	got, err := snow.WorkersObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal([]runtime.Object{wantMachineDeployment(), wantKubeadmConfigTemplate(), mt}))
}

func TestWorkersObjectsGetMachineDeploymentError(t *testing.T) {
	g := newSnowTest(t)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.MachineDeploymentName(g.clusterSpec, g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]),
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(errors.New("get md error"))

	_, err := snow.WorkersObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).NotTo(Succeed())
}

func TestWorkersObjectsGetMachineTemplateError(t *testing.T) {
	g := newSnowTest(t)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			clusterapi.MachineDeploymentName(g.clusterSpec, g.clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0]),
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(nil)
	g.kubeconfigClient.EXPECT().
		Get(
			g.ctx,
			"",
			constants.EksaSystemNamespace,
			&snowv1.AWSSnowMachineTemplate{},
		).
		Return(errors.New("get mt error"))

	_, err := snow.WorkersObjects(g.ctx, g.clusterSpec, g.kubeconfigClient)
	g.Expect(err).NotTo(Succeed())
}
