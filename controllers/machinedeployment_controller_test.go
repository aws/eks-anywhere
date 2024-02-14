package controllers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

type mdObjects struct {
	machine   *clusterv1.Machine
	mdUpgrade *anywherev1.MachineDeploymentUpgrade
	md        *clusterv1.MachineDeployment
	mhc       *clusterv1.MachineHealthCheck
}

func TestMDSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewMachineDeploymentReconciler(client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestMDReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.mdUpgrade, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileComplete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.md.Spec.Replicas = pointer.Int32(1)
	mdObjs.md.Status.UpdatedReplicas = 1

	runtimeObjs := []runtime.Object{mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	md := &clusterv1.MachineDeployment{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.md.Name, Namespace: constants.EksaSystemNamespace}, md)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(md.Annotations).ToNot(HaveKey("machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed"))

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Eventually(func(g Gomega) error {
		func(g Gomega) {
			g.Expect(mhc.Annotations).To(HaveKey("cluster.x-k8s.io/paused"))
		}(g)

		return nil
	})
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileNotNeeded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	delete(mdObjs.md.Annotations, "machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed")

	runtimeObjs := []runtime.Object{mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileCreateMachineDeploymentUpgrade(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.OwnerReferences).To(BeEquivalentTo(mdObjs.mdUpgrade.OwnerReferences))
	g.Expect(len(mdu.Spec.MachinesRequireUpgrade)).To(BeEquivalentTo(1))
	g.Expect(mdu.Spec.KubernetesVersion).To(BeEquivalentTo(mdObjs.mdUpgrade.Spec.KubernetesVersion))

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).To(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileMachineDeploymentUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.mdUpgrade.Status.Ready = true

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mdUpgrade, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).To(HaveOccurred())

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileMDAndMachineDeploymentUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.mdUpgrade.Status.Ready = true
	mdObjs.md.Status.UpdatedReplicas = *mdObjs.md.Spec.Replicas

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mdUpgrade, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError("machinedeploymentupgrades.anywhere.eks.amazonaws.com \"my-cluster-md-upgrade\" not found"))

	md := &clusterv1.MachineDeployment{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.md.Name, Namespace: constants.EksaSystemNamespace}, md)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(md.Annotations).ToNot(HaveKey("machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed"))

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileMDReadyAndMachineDeploymentUpgradeAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.md.Status.UpdatedReplicas = *mdObjs.md.Spec.Replicas

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	// verify the in-place-upgrade-needed annotation is removed even when the MachineDeploymentUpgrade object is not found
	md := &clusterv1.MachineDeployment{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.md.Name, Namespace: constants.EksaSystemNamespace}, md)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(md.Annotations).ToNot(HaveKey("machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed"))

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey("cluster.x-k8s.io/paused"))
}

func TestMDReconcileNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	client := fake.NewClientBuilder().WithRuntimeObjects().Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("machinedeployments.cluster.x-k8s.io \"my-cluster\" not found"))
}

func TestMDReconcileMHCNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("machinehealthchecks.cluster.x-k8s.io \"my-cluster-worker-unhealthy\" not found"))
}

func TestMDReconcileVersionMissing(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.md.Spec.Template.Spec.Version = nil

	runtimeObjs := []runtime.Object{mdObjs.md}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("unable to retrieve kubernetes version from MachineDeployment \"my-cluster\""))
}

func getObjectsForMD() mdObjects {
	cluster := generateCluster()
	md := generateMachineDeployment(cluster)
	md.Name = cluster.Name
	md.TypeMeta = metav1.TypeMeta{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "MachineDeployment",
	}
	node := generateNode()
	kubeadmConfig := generateKubeadmConfig()
	machine := generateMachine(cluster, node, kubeadmConfig)
	machine.Labels = map[string]string{
		"cluster.x-k8s.io/deployment-name": md.Name,
	}
	mdUpgrade := generateMDUpgrade(cluster, machine)
	mdUpgrade.Name = md.Name + "-md-upgrade"
	mdUpgrade.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: md.APIVersion,
		Kind:       md.Kind,
		Name:       md.Name,
		UID:        md.UID,
	}}
	mhc := generateMHCforMD(md.Name)

	return mdObjects{
		machine:   machine,
		mdUpgrade: mdUpgrade,
		md:        md,
		mhc:       mhc,
	}
}

func mdRequest(md *clusterv1.MachineDeployment) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      md.Name,
			Namespace: md.Namespace,
		},
	}
}

func generateMHCforMD(mdName string) *clusterv1.MachineHealthCheck {
	return &clusterv1.MachineHealthCheck{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-worker-unhealthy", mdName),
			Namespace: "eksa-system",
		},
		Spec: clusterv1.MachineHealthCheckSpec{
			NodeStartupTimeout: &metav1.Duration{
				Duration: 20 * time.Minute,
			},
		},
	}
}
