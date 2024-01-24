package controllers_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestCPUpgradeReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}

	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: "eksa-system"}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileEarly(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	cpUpgrade.Status.Ready = true
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: "eksa-system"}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeNotUpgraded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: false,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: "eksa-system"}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeUpgradeError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, _ := getObjectsForCPUpgradeTest()

	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError("getting node upgrader for machine machine02: nodeupgrades.anywhere.eks.amazonaws.com \"machine02-node-upgrader\" not found"))
}

func TestCPUpgradeReconcileNodeUpgraderCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: "eksa-system"}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	cpUpgrade.Status = anywherev1.ControlPlaneUpgradeStatus{
		Upgraded:       0,
		RequireUpgrade: 2,
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cpUpgrade.Status.Ready).To(BeFalse())
}

func TestCPUpgradeReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	now := metav1.Now()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	cpUpgrade.DeletionTimestamp = &now
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Name, Namespace: "eksa-system"}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))
}

func TestCPUpgradeObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("controlplaneupgrades.anywhere.eks.amazonaws.com \"cp-upgrade-request\" not found"))
}

func TestCPUpgradeReconcileUpdateCapiMachineVersion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], nodeUpgrades[1]}
	nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	machine := &clusterv1.Machine{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Spec.Machine.Name, Namespace: "eksa-system"}, machine)
	g.Expect(err).ToNot(HaveOccurred())
	if !strings.Contains(*machine.Spec.Version, "v1.28.3-eks-1-28-9") {
		t.Fatalf("unexpected k8s version in capi machine: %s", *machine.Spec.Version)
	}
}

func getObjectsForCPUpgradeTest() (*clusterv1.Cluster, []*clusterv1.Machine, []*corev1.Node, *anywherev1.ControlPlaneUpgrade, []*anywherev1.NodeUpgrade) {
	cluster := generateCluster()
	node1 := generateNode()
	node2 := node1.DeepCopy()
	node2.ObjectMeta.Name = "node02"
	machine1 := generateMachine(cluster, node1)
	machine2 := generateMachine(cluster, node2)
	machine2.ObjectMeta.Name = "machine02"
	nodeUpgrade1 := generateNodeUpgrade(machine1)
	nodeUpgrade2 := generateNodeUpgrade(machine2)
	nodeUpgrade2.ObjectMeta.Name = "node-upgrade-request-2"
	machines := []*clusterv1.Machine{machine1, machine2}
	nodes := []*corev1.Node{node1, node2}
	nodeUpgrades := []*anywherev1.NodeUpgrade{nodeUpgrade1, nodeUpgrade2}
	cpUpgrade := generateCPUpgrade(machines)
	return cluster, machines, nodes, cpUpgrade, nodeUpgrades
}

func cpUpgradeRequest(cpUpgrade *anywherev1.ControlPlaneUpgrade) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cpUpgrade.Name,
			Namespace: cpUpgrade.Namespace,
		},
	}
}

func generateCPUpgrade(machine []*clusterv1.Machine) *anywherev1.ControlPlaneUpgrade {
	etcdVersion := "v3.5.9-eks-1-28-9"
	return &anywherev1.ControlPlaneUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cp-upgrade-request",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.ControlPlaneUpgradeSpec{
			ControlPlane: corev1.ObjectReference{
				Name:      "my-cp",
				Namespace: "eksa-system",
				Kind:      "KubeadmControlPlane",
			},
			MachinesRequireUpgrade: []corev1.ObjectReference{
				{
					Kind:      "Machine",
					Name:      machine[0].Name,
					Namespace: machine[0].Namespace,
				},
				{
					Kind:      "Machine",
					Name:      machine[1].Name,
					Namespace: machine[1].Namespace,
				},
			},
			KubernetesVersion: "v1.28.3-eks-1-28-9",
			EtcdVersion:       etcdVersion,
		},
	}
}
