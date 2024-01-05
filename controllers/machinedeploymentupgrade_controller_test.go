package controllers_test

import (
	"context"
	"fmt"
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
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestMDUpgradeReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, md, machine, node, mdUpgrade, nodeUpgrade := getObjectsForMDUpgradeTest()
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	nodeUpgrade.Status = anywherev1.NodeUpgradeStatus{
		Completed: true,
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, md, machine, node, mdUpgrade, nodeUpgrade).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	mdu := &anywherev1.MachineDeploymentUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: mdUpgrade.Name, Namespace: "eksa-system"}, mdu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(mdu.Status.Ready).To(BeTrue())
}

func TestMDUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, md, machine, node, mdUpgrade, nodeUpgrade := getObjectsForMDUpgradeTest()
	mdUpgrade.Status = anywherev1.MachineDeploymentUpgradeStatus{
		Upgraded:       0,
		RequireUpgrade: 1,
	}
	nodeUpgrade.Name = fmt.Sprintf("%s-node-upgrader", machine.Name)
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, md, machine, node, mdUpgrade, nodeUpgrade).Build()

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

	cluster, md, machine, node, mdUpgrade, nodeUpgrade := getObjectsForMDUpgradeTest()
	mdUpgrade.DeletionTimestamp = &now
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, md, machine, node, mdUpgrade).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrade.Name, Namespace: "eksa-system"}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"node-upgrade-request\" not found"))
}

func TestMDUpgradeReconcileDeleteNodeUgradeAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, md, machine, node, mdUpgrade, nodeUpgrade := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, md, machine, node, mdUpgrade).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrade.Name, Namespace: "eksa-system"}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"node-upgrade-request\" not found"))

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeReconcileNodeUpgraderCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, md, machine, node, mdUpgrade, _ := getObjectsForMDUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, md, machine, node, mdUpgrade).Build()

	r := controllers.NewMachineDeploymentUpgradeReconciler(client)
	req := mdUpgradeRequest(mdUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	nodeUpgradeName := fmt.Sprintf("%s-node-upgrader", machine.Name)
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgradeName, Namespace: "eksa-system"}, n)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestMDUpgradeReconcileErrors(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, md, machine, node, mdUpgrade, _ := getObjectsForMDUpgradeTest()

	tests := []struct {
		name   string
		objs   []runtime.Object
		errMsg string
	}{
		{
			name:   "machinedeploymentupgrade_missing",
			objs:   []runtime.Object{cluster, md, machine, node},
			errMsg: "machinedeploymentupgrades.anywhere.eks.amazonaws.com \"md-upgrade-request\" not found",
		},
		{
			name:   "kubeadmcontrolplane_missing",
			objs:   []runtime.Object{cluster, mdUpgrade, machine, node},
			errMsg: "getting MachineDeployment my-md: machinedeployments.cluster.x-k8s.io \"my-md\" not found",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := fake.NewClientBuilder().WithRuntimeObjects(test.objs...).Build()
			r := controllers.NewMachineDeploymentUpgradeReconciler(client)
			req := mdUpgradeRequest(mdUpgrade)
			_, err := r.Reconcile(ctx, req)
			g.Expect(err).To(MatchError(test.errMsg))
		})
	}
}

func getObjectsForMDUpgradeTest() (*clusterv1.Cluster, *clusterv1.MachineDeployment, *clusterv1.Machine, *corev1.Node, *anywherev1.MachineDeploymentUpgrade, *anywherev1.NodeUpgrade) {
	cluster := generateCluster()
	md := generateMD()
	node := generateNode()
	machine := generateMachine(cluster, node, map[string]string{"cluster.x-k8s.io/deployment-name": md.Name})
	nodeUpgrade := generateNodeUpgrade(machine)
	mdUpgrade := generateMDUpgrade(machine, cluster, md)
	return cluster, md, machine, node, mdUpgrade, nodeUpgrade
}

func mdUpgradeRequest(mdUpgrade *anywherev1.MachineDeploymentUpgrade) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      mdUpgrade.Name,
			Namespace: mdUpgrade.Namespace,
		},
	}
}

func generateMDUpgrade(machine *clusterv1.Machine, cluster *clusterv1.Cluster, md *clusterv1.MachineDeployment) *anywherev1.MachineDeploymentUpgrade {
	return &anywherev1.MachineDeploymentUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "md-upgrade-request",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.MachineDeploymentUpgradeSpec{
			MachineDeployment: corev1.ObjectReference{
				Name:      md.Name,
				Namespace: md.Namespace,
				Kind:      "MachineDeployment",
			},
		},
	}
}

func generateMD() *clusterv1.MachineDeployment {
	return &clusterv1.MachineDeployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-md",
			Namespace: "eksa-system",
		},
		Spec: clusterv1.MachineDeploymentSpec{
			Template: clusterv1.MachineTemplateSpec{
				Spec: clusterv1.MachineSpec{
					Version: ptr.String(k8s128),
				},
			},
		},
	}
}
