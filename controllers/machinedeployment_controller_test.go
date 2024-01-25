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

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.mdUpgrade, mdObjs.md}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDReconcileComplete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.md.Spec.Replicas = pointer.Int32(1)
	mdObjs.md.Status.UpdatedReplicas = 1

	runtimeObjs := []runtime.Object{mdObjs.md}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDReconcileNotNeeded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	delete(mdObjs.md.Annotations, "machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed")

	runtimeObjs := []runtime.Object{mdObjs.md}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDReconcileCreateMachineDeploymentUpgrade(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md}
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
}

func TestMDReconcileMachineDeploymentUpgradeReady(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	mdObjs := getObjectsForMD()

	mdObjs.mdUpgrade.Status.Ready = true

	runtimeObjs := []runtime.Object{mdObjs.machine, mdObjs.md, mdObjs.mdUpgrade}
	client := fake.NewClientBuilder().WithRuntimeObjects(runtimeObjs...).Build()
	r := controllers.NewMachineDeploymentReconciler(client)
	req := mdRequest(mdObjs.md)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdObjs.mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).To(HaveOccurred())
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
	md := generateMD(cluster.Name)
	md.Name = cluster.Name
	md.TypeMeta = metav1.TypeMeta{
		APIVersion: clusterv1.GroupVersion.String(),
		Kind:       "MachineDeployment",
	}
	node := generateNode()
	machine := generateMachine(cluster, node)
	machine.Labels = map[string]string{
		"cluster.x-k8s.io/deployment-name": md.Name,
	}
	mdUpgrade := generateMDUpgrade(machine)
	mdUpgrade.Name = md.Name + "-md-upgrade"
	mdUpgrade.OwnerReferences = []metav1.OwnerReference{{
		APIVersion: md.APIVersion,
		Kind:       md.Kind,
		Name:       md.Name,
		UID:        md.UID,
	}}

	return mdObjects{
		machine:   machine,
		mdUpgrade: mdUpgrade,
		md:        md,
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

func generateMD(name string) *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: constants.EksaSystemNamespace,
			UID:       "test-uid",
			Annotations: map[string]string{
				"machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed": "true",
			},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Version: pointer.String("v1.28.3-eks-1-28-9"),
				},
			},
		},
	}
}
