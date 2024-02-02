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

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdUpgrade.Name, Namespace: constants.EksaSystemNamespace}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.Status.Ready).To(BeTrue())
}

func TestMDUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	mdUpgrade.Status = anywherev1.MachineDeploymentUpgradeStatus{
		Upgraded:       0,
		RequireUpgrade: 1,
	}
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdUpgrade.Status.Ready).To(BeFalse())
}

func TestMDUpgradeReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	now := metav1.Now()
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	mdUpgrade.DeletionTimestamp = &now
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrade.Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))
}

func TestMDUpgradeReconcileDeleteNodeUgradeAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	now := metav1.Now()
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	mdUpgrade.DeletionTimestamp = &now
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrade.Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeReconcileNodeUpgraderCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machine, node, mdUpgrade, _, md, ms := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	nodeUpgradeName := fmt.Sprintf("%s-node-upgrader", machine.Name)
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgradeName, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrade", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("machinedeploymentupgrades.anywhere.eks.amazonaws.com \"md-upgrade-request\" not found"))
}

func TestMDUpgradeReconcileUpdateMachineSet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	mdUpgrade.Status.Ready = true
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	ms = &clusterv1.MachineSet{}
	err = client.Get(ctx, types.NamespacedName{Name: "my-md-ms", Namespace: constants.EksaSystemNamespace}, ms)
	g.Expect(err).ToNot(HaveOccurred())
	if !strings.Contains(*ms.Spec.Template.Spec.Version, k8s128) {
		t.Fatalf("unexpected k8s version in capi machine: %s", *machine.Spec.Version)
	}
}

func TestMDUpgradeReconcileUpdateMachineSetError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	ms.Annotations[clusterv1.RevisionAnnotation] = "0"
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError(fmt.Sprintf("couldn't find machine set with revision version %v", md.Annotations[clusterv1.RevisionAnnotation])))
}

func TestMDObjectDoesNotExistError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machine, node, mdUpgrade, nodeUpgrade, _, ms := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrade", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, mdUpgrade, ms).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("getting MachineDeployment my-cluster-md: machinedeployments.cluster.x-k8s.io \"my-cluster-md\" not found"))
}

func getObjectsForMDUpgradeTest() (*clusterv1.Cluster, *clusterv1.Machine, *corev1.Node, *anywherev1.MachineDeploymentUpgrade, *anywherev1.NodeUpgrade, *clusterv1.MachineDeployment, *clusterv1.MachineSet) {
	cluster := generateCluster()
	node := generateNode()
	kubeadmConfig := generateKubeadmConfig()
	machine := generateMachine(cluster, node, kubeadmConfig)
	nodeUpgrade := generateNodeUpgrade(machine)
	mdUpgrade := generateMDUpgrade(cluster, machine)
	md := generateMachineDeployment(cluster)
	ms := generateMachineset(cluster)
	return cluster, machine, node, mdUpgrade, nodeUpgrade, md, ms
}

func mdUpgradeRequest(mdUpgrade *anywherev1.MachineDeploymentUpgrade) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      mdUpgrade.Name,
			Namespace: mdUpgrade.Namespace,
		},
	}
}

func generateMDUpgrade(cluster *clusterv1.Cluster, machine *clusterv1.Machine) *anywherev1.MachineDeploymentUpgrade {
	machineSpec := getMachineSpec(cluster)
	machineSpecJSON, _ := json.Marshal(machineSpec)
	machineSpecB64Encoded := base64.StdEncoding.EncodeToString(machineSpecJSON)
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
			MachinesRequireUpgrade: []corev1.ObjectReference{
				{
					Kind:      "Machine",
					Name:      machine.Name,
					Namespace: machine.Namespace,
				},
			},
			KubernetesVersion: k8s128,
			MachineSpecData:   machineSpecB64Encoded,
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
