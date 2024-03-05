package controllers_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const mdInPlaceAnnotation = "machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed"

type mdObjects struct {
	machine   *clusterv1.Machine
	mdUpgrade *anywherev1.MachineDeploymentUpgrade
	md        *clusterv1.MachineDeployment
	mhc       *clusterv1.MachineHealthCheck
}

func TestMDSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewMachineDeploymentReconciler(client, client)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager())).To(Succeed())
}

func TestMDReconcileNotNeeded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	delete(mdObjs.md.Annotations, mdInPlaceAnnotation)

	runtimeObjs := []runtime.Object{mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestMDReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.mdUpgrade, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestMDReconcileCreateMachineDeploymentUpgrade(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
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
	g.Expect(mhc.Annotations).To(HaveKey(capiPausedAnnotation))
}

func TestMDReconcileMDAndMachineDeploymentUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.mdUpgrade.Status.Ready = true
	mdObjs.machine.Spec.Version = &mdObjs.mdUpgrade.Spec.KubernetesVersion

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mdUpgrade, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
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
	g.Expect(md.Annotations).ToNot(HaveKey(mdInPlaceAnnotation))

	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestMDReconcileFullFlow(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mhc}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	// Expect MachineDeploymentUpgrade object to be created and not ready
	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.Status.Ready).To(BeFalse())

	// Expect KCP to still have in-place annotation
	md := &clusterv1.MachineDeployment{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.md.Name, Namespace: constants.EksaSystemNamespace}, md)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(md.Annotations).To(HaveKey(mdInPlaceAnnotation))

	// Expect MHC for KCP to be paused
	mhc := &clusterv1.MachineHealthCheck{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).To(HaveKey(capiPausedAnnotation))

	machine := &clusterv1.Machine{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.machine.Name, Namespace: constants.EksaSystemNamespace}, machine)
	g.Expect(err).ToNot(HaveOccurred())

	// Mark MachineDeploymentUpgrade as ready and update Machine K8s version
	mdu.Status.Ready = true
	err = client.Update(ctx, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	machine.Spec.Version = &mdu.Spec.KubernetesVersion
	err = client.Update(ctx, machine)
	g.Expect(err).ToNot(HaveOccurred())

	// trigger another reconcile loop
	req = mdRequest(md)
	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	// Expect MachineDeploymentUpgrade object to be deleted
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).To(HaveOccurred())
	g.Expect(apierrors.IsNotFound(err)).To(BeTrue())

	// Expect MD to no longer have in-place annotation
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.md.Name, Namespace: constants.EksaSystemNamespace}, md)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(md.Annotations).ToNot(HaveKey(mdInPlaceAnnotation))

	// Expect MHC for MD to not be paused
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mhc.Name, Namespace: constants.EksaSystemNamespace}, mhc)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mhc.Annotations).ToNot(HaveKey(capiPausedAnnotation))
}

func TestMDReconcileNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	client := fake.NewClientBuilder().WithRuntimeObjects().Build()
	r := controllers.NewMachineDeploymentReconciler(client, client)
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
	r := controllers.NewMachineDeploymentReconciler(client, client)
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
	r := controllers.NewMachineDeploymentReconciler(client, client)
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
