package clusterapi_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clusterapi/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type fetchTest struct {
	*WithT
	ctx                   context.Context
	kubeClient            *mocks.MockKubeClient
	clusterSpec           *cluster.Spec
	workerNodeGroupConfig v1alpha1.WorkerNodeGroupConfiguration
	machineDeployment     *clusterv1.MachineDeployment
}

func newFetchTest(t *testing.T) fetchTest {
	ctrl := gomock.NewController(t)
	kubeClient := mocks.NewMockKubeClient(ctrl)
	wng := v1alpha1.WorkerNodeGroupConfiguration{
		Name: "md-0",
	}
	md := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							Name: "snow-test-md-0-1",
						},
					},
				},
			},
		},
	}
	return fetchTest{
		WithT:                 NewWithT(t),
		ctx:                   context.Background(),
		kubeClient:            kubeClient,
		clusterSpec:           givenClusterSpec(),
		workerNodeGroupConfig: wng,
		machineDeployment:     md,
	}
}

func TestMachineDeploymentInCluster(t *testing.T) {
	g := newFetchTest(t)
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0",
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *clusterv1.MachineDeployment) error {
			g.machineDeployment.DeepCopyInto(obj)
			return nil
		})

	got, err := clusterapi.MachineDeploymentInCluster(g.ctx, g.kubeClient, g.clusterSpec, g.workerNodeGroupConfig)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(g.machineDeployment))
}

func TestMachineDeploymentInClusterNotExists(t *testing.T) {
	g := newFetchTest(t)
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0",
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))

	got, err := clusterapi.MachineDeploymentInCluster(g.ctx, g.kubeClient, g.clusterSpec, g.workerNodeGroupConfig)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(BeNil())
}

func TestMachineDeploymentInClusterError(t *testing.T) {
	g := newFetchTest(t)
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0",
			constants.EksaSystemNamespace,
			&clusterv1.MachineDeployment{},
		).
		Return(errors.New("get md error"))

	got, err := clusterapi.MachineDeploymentInCluster(g.ctx, g.kubeClient, g.clusterSpec, g.workerNodeGroupConfig)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(BeNil())
}

func TestKubeadmConfigTemplateInCluster(t *testing.T) {
	g := newFetchTest(t)
	kct := &bootstrapv1.KubeadmConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "kct-1",
		},
	}
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0-1",
			constants.EksaSystemNamespace,
			&bootstrapv1.KubeadmConfigTemplate{},
		).
		DoAndReturn(func(_ context.Context, _, _ string, obj *bootstrapv1.KubeadmConfigTemplate) error {
			kct.DeepCopyInto(obj)
			return nil
		})

	got, err := clusterapi.KubeadmConfigTemplateInCluster(g.ctx, g.kubeClient, g.machineDeployment)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(Equal(kct))
}

func TestKubeadmConfigTemplateInClusterMachineDeploymentNil(t *testing.T) {
	g := newFetchTest(t)
	got, err := clusterapi.KubeadmConfigTemplateInCluster(g.ctx, g.kubeClient, nil)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(BeNil())
}

func TestKubeadmConfigTemplateInClusterNotExists(t *testing.T) {
	g := newFetchTest(t)
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0-1",
			constants.EksaSystemNamespace,
			&bootstrapv1.KubeadmConfigTemplate{},
		).
		Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))

	got, err := clusterapi.KubeadmConfigTemplateInCluster(g.ctx, g.kubeClient, g.machineDeployment)
	g.Expect(err).To(Succeed())
	g.Expect(got).To(BeNil())
}

func TestKubeadmConfigTemplateInClusterError(t *testing.T) {
	g := newFetchTest(t)
	g.kubeClient.EXPECT().
		Get(
			g.ctx,
			"snow-test-md-0-1",
			constants.EksaSystemNamespace,
			&bootstrapv1.KubeadmConfigTemplate{},
		).
		Return(errors.New("get kct error"))

	got, err := clusterapi.KubeadmConfigTemplateInCluster(g.ctx, g.kubeClient, g.machineDeployment)
	g.Expect(err).NotTo(Succeed())
	g.Expect(got).To(BeNil())
}
