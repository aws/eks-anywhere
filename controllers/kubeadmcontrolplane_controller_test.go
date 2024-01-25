package controllers_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type kcpObjects struct {
	machines  []*clusterv1.Machine
	cpUpgrade *anywherev1.ControlPlaneUpgrade
	kcp       *controlplanev1.KubeadmControlPlane
}

func TestKCPSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewKubeadmControlPlaneReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestKCPReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.cpUpgrade, kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestKCPReconcileComplete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	count := int32(len(kcpObjs.machines))
	kcpObjs.kcp.Spec.Replicas = pointer.Int32(count)
	kcpObjs.kcp.Status.UpdatedReplicas = count

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	kcp := &controlplanev1.KubeadmControlPlane{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.kcp.Name, Namespace: constants.EksaSystemNamespace}, kcp)
	g.Expect(err).ToNot(HaveOccurred())
	_, ok := kcp.Annotations["controlplane.clusters.x-k8s.io/in-place-upgrade-needed"]
	g.Expect(ok).To(BeFalse())
}

func TestKCPReconcileNotNeeded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	delete(kcpObjs.kcp.Annotations, "controlplane.clusters.x-k8s.io/in-place-upgrade-needed")

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestKCPReconcileCreateControlPlaneUpgrade(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cpu.OwnerReferences).To(BeEquivalentTo(kcpObjs.cpUpgrade.OwnerReferences))
	g.Expect(len(cpu.Spec.MachinesRequireUpgrade)).To(BeEquivalentTo(len(kcpObjs.cpUpgrade.Spec.MachinesRequireUpgrade)))
	g.Expect(cpu.Spec.EtcdVersion).To(BeEquivalentTo(kcpObjs.cpUpgrade.Spec.EtcdVersion))
	g.Expect(cpu.Spec.KubernetesVersion).To(BeEquivalentTo(kcpObjs.cpUpgrade.Spec.KubernetesVersion))
}

func TestKCPReconcileControlPlaneUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.cpUpgrade.Status.Ready = true

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.cpUpgrade, kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).To(HaveOccurred())
}

func TestKCPReconcileNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	client := fake.NewClientBuilder().WithRuntimeObjects().Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("kubeadmcontrolplanes.controlplane.cluster.x-k8s.io \"my-cluster\" not found"))
}

func TestKCPReconcileClusterConfigurationMissing(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.kcp.Spec.KubeadmConfigSpec.ClusterConfiguration = nil

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("ClusterConfiguration not set for KubeadmControlPlane \"my-cluster\", unable to retrieve etcd information"))
}

func TestKCPReconcileStackedEtcdMissing(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local = nil

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("local etcd configuration is missing"))
}

func getObjectsForKCP() kcpObjects {
	cluster := generateCluster()
	kcp := generateKCP(cluster.Name)
	kcp.Name = cluster.Name
	kcp.TypeMeta = metav1.TypeMeta{
		APIVersion: controlplanev1.GroupVersion.String(),
		Kind:       "KubeadmControlPlane",
	}
	node1 := generateNode()
	node2 := node1.DeepCopy()
	node2.ObjectMeta.Name = "node02"
	machine1 := generateMachine(cluster, node1)
	machine1.Labels = map[string]string{
		"cluster.x-k8s.io/control-plane-name": kcp.Name,
	}
	machine2 := generateMachine(cluster, node2)
	machine2.ObjectMeta.Name = "machine02"
	machine2.Labels = map[string]string{
		"cluster.x-k8s.io/control-plane-name": kcp.Name,
	}
	machines := []*clusterv1.Machine{machine1, machine2}
	cpUpgrade := generateCPUpgrade(machines)
	cpUpgrade.Name = kcp.Name + "-cp-upgrade"
	cpUpgrade.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: kcp.APIVersion,
		Kind:       kcp.Kind,
		Name:       kcp.Name,
		UID:        kcp.UID,
	}}

	return kcpObjects{
		machines:  machines,
		cpUpgrade: cpUpgrade,
		kcp:       kcp,
	}
}

func kcpRequest(kcp *controlplanev1.KubeadmControlPlane) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kcp.Name,
			Namespace: kcp.Namespace,
		},
	}
}

func generateKCP(name string) *controlplanev1.KubeadmControlPlane {
	return &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			UID:       "test-uid",
			Annotations: map[string]string{
				"controlplane.clusters.x-k8s.io/in-place-upgrade-needed": "true",
			},
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					Etcd: bootstrapv1.Etcd{
						Local: &bootstrapv1.LocalEtcd{
							ImageMeta: bootstrapv1.ImageMeta{
								ImageTag: "v3.5.9-eks-1-28-9",
							},
						},
					},
				},
			},
			Version: "v1.28.3-eks-1-28-9",
		},
	}
}
