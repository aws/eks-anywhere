package controllers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	kcpInPlaceAnnotation = "controlplane.clusters.x-k8s.io/in-place-upgrade-needed"
	capiPausedAnnotation = "cluster.x-k8s.io/paused"
)

type kcpObjects struct {
	machines  []*clusterv1beta2.Machine
	cpUpgrade *anywherev1.ControlPlaneUpgrade
	kcp       *controlplanev1beta2.KubeadmControlPlane
	mhc       *clusterv1beta2.MachineHealthCheck
}

func TestKCPSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestKCPReconcileNotNeeded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	delete(kcpObjs.kcp.Annotations, kcpInPlaceAnnotation)

	runtimeObjs := []runtime.Object{kcpObjs.kcp, kcpObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mhc := &clusterv1beta2.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestKCPReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.cpUpgrade, kcpObjs.kcp, kcpObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestKCPReconcileCreateControlPlaneUpgrade(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.kcp, kcpObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
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
	kcpSpec, err := json.Marshal(kcpObjs.kcp.Spec)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cpu.Spec.ControlPlaneSpecData).To(BeEquivalentTo(base64.StdEncoding.EncodeToString(kcpSpec)))

	mhc := &clusterv1beta2.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).To(HaveKey(capiPausedAnnotation))
}

func TestKCPReconcileKCPAndControlPlaneUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.kcp.Status.Version = kcpObjs.kcp.Spec.Version
	kcpObjs.cpUpgrade.Status.Ready = true

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.cpUpgrade, kcpObjs.kcp, kcpObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError("controlplaneupgrades.anywhere.eks.amazonaws.com \"my-cluster-cp-upgrade\" not found"))

	kcp := &controlplanev1beta2.KubeadmControlPlane{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.kcp.Name, Namespace: constants.EksaSystemNamespace}, kcp)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(kcp.Annotations).ToNot(HaveKey(kcpInPlaceAnnotation))

	mhc := &clusterv1beta2.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestKCPReconcileControlPlaneUpgradeVersionAliases(t *testing.T) {
	const target = "v1.34.12-eks-1-34-25"
	tests := []struct {
		name        string
		observed    string
		desired     string
		upgrade     string
		ready       bool
		wantDeleted bool
	}{
		{"EKS-D source alias", "v1.34.12-eks-abcdef1", target, target, true, true},
		{"same release label", target, target, target, true, true},
		{"not ready", "v1.34.12-eks-abcdef1", target, target, false, false},
		{"different patch", "v1.34.11-eks-abcdef1", target, target, true, false},
		{"different minor", "v1.33.12-eks-abcdef1", target, target, true, false},
		{"invalid observed", "not-a-version", target, target, true, false},
		{"invalid identical versions", "not-a-version", "not-a-version", "not-a-version", true, false},
		{"unknown prerelease", "v1.34.12-rc.1", target, target, true, false},
		{"malformed source alias", "v1.34.12-eks-not-a-hash", target, target, true, false},
		{"wrong release branch", "v1.34.12-eks-abcdef1", "v1.34.12-eks-1-33-25", "v1.34.12-eks-1-33-25", true, false},
		{"desired target changed", "v1.34.12-eks-abcdef1", "v1.34.12-eks-1-34-26", target, true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			objs := getObjectsForKCP()
			objs.kcp.Spec.Version = tt.desired
			objs.kcp.Status.Version = tt.observed
			objs.cpUpgrade.Spec.KubernetesVersion = tt.upgrade
			objs.cpUpgrade.Status.Ready = tt.ready
			objs.mhc.Annotations = map[string]string{capiPausedAnnotation: ""}
			client := fake.NewClientBuilder().WithRuntimeObjects(objs.kcp, objs.mhc, objs.cpUpgrade).Build()
			r := controllers.NewKubeadmControlPlaneReconciler(client, client)
			_, err := r.Reconcile(ctx, kcpRequest(objs.kcp))
			g.Expect(err).ToNot(HaveOccurred())
			cpu := &anywherev1.ControlPlaneUpgrade{}
			err = client.Get(ctx, types.NamespacedName{Name: objs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
			if tt.wantDeleted {
				g.Expect(apierrors.IsNotFound(err)).To(BeTrue())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
			kcp := &controlplanev1beta2.KubeadmControlPlane{}
			g.Expect(client.Get(ctx, types.NamespacedName{Name: objs.kcp.Name, Namespace: constants.EksaSystemNamespace}, kcp)).To(Succeed())
			mhc := &clusterv1beta2.MachineHealthCheck{}
			g.Expect(client.Get(ctx, types.NamespacedName{Name: objs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)).To(Succeed())
			if tt.wantDeleted {
				g.Expect(kcp.Annotations).ToNot(HaveKey(kcpInPlaceAnnotation))
				g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
			} else {
				g.Expect(kcp.Annotations).To(HaveKey(kcpInPlaceAnnotation))
				g.Expect(mhc.Annotations).To(HaveKey(capiPausedAnnotation))
			}
		})
	}
}

func TestKCPReconcileFullFlow(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.kcp, kcpObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	// Expect ControlPlaneUpgrade object to be created and not ready
	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cpu.Status.Ready).To(BeFalse())

	// Expect KCP to still have in-place annotation
	kcp := &controlplanev1beta2.KubeadmControlPlane{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.kcp.Name, Namespace: constants.EksaSystemNamespace}, kcp)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(kcp.Annotations).To(HaveKey(kcpInPlaceAnnotation))

	// Expect MHC for KCP to be paused
	mhc := &clusterv1beta2.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).To(HaveKey(capiPausedAnnotation))

	// Mark ControlPlaneUpgrade as ready and update KCP status K8s version
	cpu.Status.Ready = true
	err = client.Update(ctx, cpu)
	g.Expect(err).ToNot(HaveOccurred())
	kcp.Status.Version = kcp.Spec.Version
	err = client.Update(ctx, kcp)
	g.Expect(err).ToNot(HaveOccurred())

	// trigger another reconcile loop
	req = kcpRequest(kcp)
	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	// Expect ControlPlaneUpgrade object to be deleted
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).To(HaveOccurred())
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())

	// Expect KCP to no longer have in-place annotation
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.kcp.Name, Namespace: constants.EksaSystemNamespace}, kcp)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(kcp.Annotations).ToNot(HaveKey(kcpInPlaceAnnotation))

	// Expect MHC for KCP to not be paused
	err = client.Get(ctx, types.NamespacedName{Name: kcpObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestKCPReconcileNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	client := fake.NewClientBuilder().WithRuntimeObjects().Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("kubeadmcontrolplanes.controlplane.cluster.x-k8s.io \"my-cluster\" not found"))
}

func TestKCPReconcileMHCNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	runtimeObjs := []runtime.Object{kcpObjs.machines[0], kcpObjs.machines[1], kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("machinehealthchecks.cluster.x-k8s.io \"my-cluster-kcp-unhealthy\" not found"))
}

func TestKCPReconcileClusterConfigurationMissing(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.kcp.Spec.KubeadmConfigSpec.ClusterConfiguration = bootstrapv1beta2.ClusterConfiguration{}

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("local etcd configuration is missing for KubeadmControlPlane \"my-cluster\""))
}

func TestKCPReconcileStackedEtcdMissing(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	kcpObjs := getObjectsForKCP()

	kcpObjs.kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.Etcd.Local = bootstrapv1beta2.LocalEtcd{}

	runtimeObjs := []runtime.Object{kcpObjs.kcp}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewKubeadmControlPlaneReconciler(client, client)
	req := kcpRequest(kcpObjs.kcp)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("local etcd configuration is missing for KubeadmControlPlane \"my-cluster\""))
}

func getObjectsForKCP() kcpObjects {
	cluster := generateCluster()
	kcp := generateKCP(cluster.Name)
	kcp.Name = cluster.Name
	kcp.TypeMeta = metav1.TypeMeta{
		APIVersion: controlplanev1beta2.GroupVersion.String(),
		Kind:       "KubeadmControlPlane",
	}
	node1 := generateNode()
	node2 := node1.DeepCopy()
	node2.ObjectMeta.Name = "node02"
	kubeadmConfig1 := generateKubeadmConfig()
	kubeadmConfig2 := generateKubeadmConfig()
	machine1 := generateMachine(cluster, node1, kubeadmConfig1)
	machine1.Labels = map[string]string{
		"cluster.x-k8s.io/control-plane-name": kcp.Name,
	}
	machine2 := generateMachine(cluster, node2, kubeadmConfig2)
	machine2.ObjectMeta.Name = "machine02"
	machine2.Labels = map[string]string{
		"cluster.x-k8s.io/control-plane-name": kcp.Name,
	}
	machines := []*clusterv1beta2.Machine{machine1, machine2}
	cpUpgrade := generateCPUpgrade(machines)
	cpUpgrade.Name = kcp.Name + "-cp-upgrade"
	cpUpgrade.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: kcp.APIVersion,
		Kind:       kcp.Kind,
		Name:       kcp.Name,
		UID:        kcp.UID,
	}}
	mhc := generateMHCforKCP(kcp.Name)

	return kcpObjects{
		machines:  machines,
		cpUpgrade: cpUpgrade,
		kcp:       kcp,
		mhc:       mhc,
	}
}

func kcpRequest(kcp *controlplanev1beta2.KubeadmControlPlane) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      kcp.Name,
			Namespace: kcp.Namespace,
		},
	}
}

func generateKCP(name string) *controlplanev1beta2.KubeadmControlPlane {
	return &controlplanev1beta2.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			UID:       "test-uid",
			Annotations: map[string]string{
				kcpInPlaceAnnotation: "true",
			},
		},
		Spec: controlplanev1beta2.KubeadmControlPlaneSpec{
			KubeadmConfigSpec: bootstrapv1beta2.KubeadmConfigSpec{
				ClusterConfiguration: bootstrapv1beta2.ClusterConfiguration{
					Etcd: bootstrapv1beta2.Etcd{
						Local: bootstrapv1beta2.LocalEtcd{
							ImageTag: etcd129,
						},
					},
				},
			},
			Replicas: ptr.To(int32(3)),
			Version:  k8s129,
		},
	}
}

func generateMHCforKCP(kcpName string) *clusterv1beta2.MachineHealthCheck {
	return &clusterv1beta2.MachineHealthCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-kcp-unhealthy", kcpName),
			Namespace: "eksa-system",
		},
		Spec: clusterv1beta2.MachineHealthCheckSpec{
			Checks: clusterv1beta2.MachineHealthCheckChecks{
				NodeStartupTimeoutSeconds: ptr.To(int32(1200)), // 20 minutes
			},
		},
	}
}
