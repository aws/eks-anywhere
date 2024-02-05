package controllers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestMDUpgradeReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], mdUpgrade, nodeUpgrades[0], nodeUpgrades[1], md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.Status.RequireUpgrade).To(BeEquivalentTo(0))
	g.Expect(mdu.Status.Upgraded).To(BeEquivalentTo(2))
	g.Expect(mdu.Status.Ready).To(BeTrue())
}

func TestMDUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrades[1].Status = anywherev1.NodeUpgradeStatus{
		Completed: false,
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], mdUpgrade, nodeUpgrades[0], nodeUpgrades[1], md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.Status.RequireUpgrade).To(BeEquivalentTo(1))
	g.Expect(mdu.Status.Upgraded).To(BeEquivalentTo(1))
	g.Expect(mdUpgrade.Status.Ready).To(BeFalse())
}

func TestMDUpgradeReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	now := metav1.Now()
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	mdUpgrade.DeletionTimestamp = &now
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], mdUpgrade, nodeUpgrades[0], nodeUpgrades[1], md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))

	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[1].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine02-node-upgrader\" not found"))
}

func TestMDUpgradeReconcileDeleteNodeUpgradeAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	now := metav1.Now()
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	mdUpgrade.DeletionTimestamp = &now
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], mdUpgrade, nodeUpgrades[0], nodeUpgrades[1], md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))

	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[1].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine02-node-upgrader\" not found"))

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeReconcileNodeUpgraderCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, mdUpgrade, _, md, ms := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], mdUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	nodeUpgradeName := fmt.Sprintf("%s-node-upgrader", machines[0].Name)
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgradeName, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("machinedeploymentupgrades.anywhere.eks.amazonaws.com \"md-upgrade-request\" not found"))
}

func TestMDUpgradeReconcileUpdateMachineSet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	mdUpgrade.Status.Ready = true
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], mdUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	ms = &clusterv1.MachineSet{}
	err = client.Get(ctx, types.NamespacedName{Name: "my-md-ms", Namespace: constants.EksaSystemNamespace}, ms)
	g.Expect(err).ToNot(HaveOccurred())
	if !strings.Contains(*ms.Spec.Template.Spec.Version, k8s128) {
		t.Fatalf("unexpected k8s version in capi machine: %s", *machines[0].Spec.Version)
	}
}

func TestMDUpgradeReconcileUpdateMachineSetError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, md, ms := getObjectsForMDUpgradeTest()
	ms.Annotations[clusterv1.RevisionAnnotation] = "0"
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], mdUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(fmt.Sprintf("couldn't find machine set with revision version %v", md.Annotations[clusterv1.RevisionAnnotation])))
}

func TestMDObjectDoesNotExistError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, mdUpgrade, nodeUpgrades, _, ms := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], mdUpgrade, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("getting MachineDeployment my-cluster-md: machinedeployments.cluster.x-k8s.io \"my-cluster-md\" not found"))
}

func getObjectsForMDUpgradeTest() (*clusterv1.Cluster, []*clusterv1.Machine, []*corev1.Node, *anywherev1.MachineDeploymentUpgrade, []*anywherev1.NodeUpgrade, *clusterv1.MachineDeployment, *clusterv1.MachineSet) {
	cluster := generateCluster()
	node1 := generateNode()
	node2 := node1.DeepCopy()
	node2.ObjectMeta.Name = "node02"
	kubeadmConfig1 := generateKubeadmConfig()
	kubeadmConfig2 := generateKubeadmConfig()
	machine1 := generateMachine(cluster, node1, kubeadmConfig1)
	machine2 := generateMachine(cluster, node2, kubeadmConfig2)
	machine2.ObjectMeta.Name = "machine02"
	nodeUpgrade1 := generateNodeUpgrade(machine1)
	nodeUpgrade1.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	nodeUpgrade1.Name = fmt.Sprintf("%s-node-upgrader", machine1.Name)
	nodeUpgrade2 := generateNodeUpgrade(machine2)
	nodeUpgrade2.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	nodeUpgrade2.Name = fmt.Sprintf("%s-node-upgrader", machine2.Name)
	mdUpgrade := generateMDUpgrade(cluster, machine1, machine2)
	md := generateMachineDeployment(cluster)
	ms := generateMachineset(cluster)
	return cluster, []*clusterv1.Machine{machine1, machine2}, []*corev1.Node{node1, node2}, mdUpgrade, []*anywherev1.NodeUpgrade{nodeUpgrade1, nodeUpgrade2}, md, ms
}

func mdUpgradeRequest(mdUpgrade *anywherev1.MachineDeploymentUpgrade) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      mdUpgrade.Name,
			Namespace: mdUpgrade.Namespace,
		},
	}
}

func generateMDUpgrade(cluster *clusterv1.Cluster, machines ...*clusterv1.Machine) *anywherev1.MachineDeploymentUpgrade {
	machineSpec := getMachineSpec(cluster)
	machineSpecJSON, _ := json.Marshal(machineSpec)
	machineSpecB64Encoded := base64.StdEncoding.EncodeToString(machineSpecJSON)
	machinesRequireUpdate := []corev1.ObjectReference{}
	for i := range machines {
		machine := machines[i]
		machinesRequireUpdate = append(machinesRequireUpdate, corev1.ObjectReference{
			Kind:      "Machine",
			Name:      machine.Name,
			Namespace: machine.Namespace,
		})
	}

	return &anywherev1.MachineDeploymentUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "md-upgrade-request",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.MachineDeploymentUpgradeSpec{
			MachineDeployment: corev1.ObjectReference{
				Name:      fmt.Sprintf("%s-md", cluster.Name),
				Kind:      "MachineDeployment",
				Namespace: constants.EksaSystemNamespace,
			},
			MachinesRequireUpgrade: machinesRequireUpdate,
			KubernetesVersion:      k8s128,
			MachineSpecData:        machineSpecB64Encoded,
		},
	}
}

func generateMachineset(cluster *clusterv1.Cluster) *clusterv1.MachineSet {
	ms := getMachineSpec(cluster)
	ms.Version = ptr.String(k8s127)
	return &clusterv1.MachineSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-md-ms",
			Namespace:   constants.EksaSystemNamespace,
			Labels:      map[string]string{"cluster.x-k8s.io/deployment-name": fmt.Sprintf("%s-md", cluster.Name)},
			Annotations: map[string]string{clusterv1.RevisionAnnotation: "1"},
		},
		Spec: clusterv1.MachineSetSpec{
			ClusterName: cluster.Name,
			Replicas:    new(int32),
			Template: clusterv1.MachineTemplateSpec{
				Spec: *ms,
			},
		},
	}
}

func generateMachineDeployment(cluster *clusterv1.Cluster) *clusterv1.MachineDeployment {
	ms := getMachineSpec(cluster)
	return &clusterv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-md", cluster.Name),
			Namespace: constants.EksaSystemNamespace,
			Annotations: map[string]string{
				clusterv1.RevisionAnnotation:                                  "1",
				"machinedeployment.clusters.x-k8s.io/in-place-upgrade-needed": "true",
			},
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: cluster.Name,
			Replicas:    ptr.Int32(1),
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{"cluster.x-k8s.io/deployment-name": cluster.Name},
			},
			Template: clusterv1.MachineTemplateSpec{
				Spec: *ms,
			},
		},
	}
}

func getMachineSpec(cluster *clusterv1.Cluster) *clusterv1.MachineSpec {
	return &clusterv1.MachineSpec{
		ClusterName: cluster.Name,
		Version:     ptr.String(k8s128),
	}
}
