package controllers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/rand"
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

const (
	etcd128 = "v3.5.9-eks-1-28-9"
	etcd129 = "v3.5.10-eks-1-29-0"
	k8s127  = "v1.27.1-eks-1-27-1"
	k8s128  = "v1.28.3-eks-1-28-9"
	k8s129  = "v1.29.0-eks-1-29-0"
)

type cpUpgradeObjects struct {
	cluster        *clusterv1.Cluster
	machines       []*clusterv1.Machine
	nodes          []*corev1.Node
	cpUpgrade      *anywherev1.ControlPlaneUpgrade
	nodeUpgrades   []*anywherev1.NodeUpgrade
	kubeadmConfigs []*bootstrapv1.KubeadmConfig
	infraMachines  []*tinkerbellv1.TinkerbellMachine
}

func TestCPUpgradeReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}

	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileEarly(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	testObjs.cpUpgrade.Status.Ready = true
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeNotUpgraded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: false,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeUpgradeEnsureStatusUpdated(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()

	objs := []runtime.Object{
		testObjs.cluster, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1], testObjs.cpUpgrade,
		testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(cpu.Status.RequireUpgrade).To(BeEquivalentTo(2))
	g.Expect(cpu.Status.Upgraded).To(BeEquivalentTo(0))
	g.Expect(cpu.Status.Ready).To(BeFalse())
}

func TestCPUpgradeReconcileNodeUpgraderCreate(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1], testObjs.cpUpgrade,
		testObjs.nodeUpgrades[0], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeUpgraderInvalidKCPSpec(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}

	for _, test := range []struct {
		name    string
		kcpSpec string
		errMsg  string
	}{
		{
			name:    "invalid base64",
			kcpSpec: "not-a-valid-base-64",
			errMsg:  "decoding cpUpgrade.Spec.ControlPlaneSpec",
		},
		{
			name:    "invalid json",
			kcpSpec: "aW52YWxpZC1qc29uCg==",
			errMsg:  "unmarshaling cpUpgrade.Spec.ControlPlaneSpec",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			testObjs.cpUpgrade.Spec.ControlPlaneSpecData = test.kcpSpec
			objs := []runtime.Object{
				testObjs.cluster, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
				testObjs.cpUpgrade, testObjs.nodeUpgrades[0], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
			}
			client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
			r := controllers.NewControlPlaneUpgradeReconciler(client)
			req := cpUpgradeRequest(testObjs.cpUpgrade)
			_, err := r.Reconcile(ctx, req)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(test.errMsg))
		})
	}
}

func TestCPUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	testObjs.cpUpgrade.Status = anywherev1.ControlPlaneUpgradeStatus{
		Upgraded:       0,
		RequireUpgrade: 2,
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(testObjs.cpUpgrade.Status.Ready).To(BeFalse())
}

func TestCPUpgradeReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	now := metav1.Now()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	testObjs.cpUpgrade.DeletionTimestamp = &now
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.nodeUpgrades[0].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))
}

func TestCPUpgradeObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("controlplaneupgrades.anywhere.eks.amazonaws.com \"cp-upgrade-request\" not found"))
}

func TestCPUpgradeReconcileUpdateCapiMachineVersion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	machine := &clusterv1.Machine{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.nodeUpgrades[0].Spec.Machine.Name, Namespace: constants.EksaSystemNamespace}, machine)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(*machine.Spec.Version).To(BeEquivalentTo(k8s128))
}

func TestCPUpgradeReconcileUpdateKubeadmConfigSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	kcpDecoded, err := base64.StdEncoding.DecodeString(testObjs.cpUpgrade.Spec.ControlPlaneSpecData)
	g.Expect(err).ToNot(HaveOccurred())

	kcpSpec := &controlplanev1.KubeadmControlPlaneSpec{}
	err = json.Unmarshal(kcpDecoded, kcpSpec)
	g.Expect(err).ToNot(HaveOccurred())

	for i := range testObjs.kubeadmConfigs {
		kc := testObjs.kubeadmConfigs[i]
		kcNew := &bootstrapv1.KubeadmConfig{}
		err = client.Get(ctx, types.NamespacedName{Name: kc.Name, Namespace: kc.Namespace}, kcNew)
		g.Expect(err).ToNot(HaveOccurred())

		kcsCopy := kcpSpec.KubeadmConfigSpec.DeepCopy()
		if kcNew.Spec.InitConfiguration == nil {
			kcsCopy.InitConfiguration = nil
		}
		if kcNew.Spec.JoinConfiguration == nil {
			kcsCopy.JoinConfiguration = nil
		}

		g.Expect(kcNew.Spec).To(BeEquivalentTo(*kcsCopy))
	}
}

func TestCPUpgradeReconcileUpdateKubeadmConfigRefNil(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	for i := range testObjs.machines {
		testObjs.machines[i].Spec.Bootstrap.ConfigRef = nil
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("updating kubeadm config: bootstrap config for machine machine01 is nil"))
}

func TestCPUpgradeReconcileUpdateKubeadmConfigNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("updating kubeadm config: retrieving bootstrap config for machine machine01: kubeadmconfigs.bootstrap.cluster.x-k8s.io \"kubeadm-config-\\w{10}\" not found"))
}

func TestCPUpgradeReconcileUpdateInfraMachineAnnotationSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	infraMachine1 := &tinkerbellv1.TinkerbellMachine{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.infraMachines[0].Name, Namespace: constants.EksaSystemNamespace}, infraMachine1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(infraMachine1.Annotations["cluster.x-k8s.io/cloned-from-name"]).To(BeEquivalentTo("new-ref"))
	infraMachine2 := &tinkerbellv1.TinkerbellMachine{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.infraMachines[1].Name, Namespace: constants.EksaSystemNamespace}, infraMachine2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(infraMachine2.Annotations["cluster.x-k8s.io/cloned-from-name"]).To(BeEquivalentTo("new-ref"))
}

func TestCPUpgradeReconcileUpdateInfraMachineAnnotationNilSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	testObjs.infraMachines[0].Annotations = nil
	testObjs.infraMachines[1].Annotations = nil
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1], testObjs.infraMachines[0], testObjs.infraMachines[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	infraMachine1 := &tinkerbellv1.TinkerbellMachine{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.infraMachines[0].Name, Namespace: constants.EksaSystemNamespace}, infraMachine1)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(infraMachine1.Annotations["cluster.x-k8s.io/cloned-from-name"]).To(BeEquivalentTo("new-ref"))
	infraMachine2 := &tinkerbellv1.TinkerbellMachine{}
	err = client.Get(ctx, types.NamespacedName{Name: testObjs.infraMachines[1].Name, Namespace: constants.EksaSystemNamespace}, infraMachine2)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(infraMachine2.Annotations["cluster.x-k8s.io/cloned-from-name"]).To(BeEquivalentTo("new-ref"))
}

func TestCPUpgradeReconcileUpdateInfraMachineAnnotationErrror(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	testObjs := getObjectsForCPUpgradeTest()
	for i := range testObjs.nodeUpgrades {
		testObjs.nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", testObjs.machines[i].Name)
		testObjs.nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{
		testObjs.cluster, testObjs.cpUpgrade, testObjs.machines[0], testObjs.machines[1], testObjs.nodes[0], testObjs.nodes[1],
		testObjs.nodeUpgrades[0], testObjs.nodeUpgrades[1], testObjs.kubeadmConfigs[0], testObjs.kubeadmConfigs[1],
	}
	testObjs.nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(testObjs.cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err).To(MatchError("updating infra machine: retrieving infra machine machine01 for machine machine01: failed to retrieve TinkerbellMachine external object \"eksa-system\"/\"machine01\": tinkerbellmachines.infrastructure.cluster.x-k8s.io \"machine01\" not found"))
}

func getObjectsForCPUpgradeTest() cpUpgradeObjects {
	cluster := generateCluster()
	node1 := generateNode()
	node2 := node1.DeepCopy()
	node2.ObjectMeta.Name = "node02"
	kubeadmConfig1 := generateKubeadmConfig()
	kubeadmConfig2 := generateKubeadmConfig()
	machine1 := generateMachine(cluster, node1, kubeadmConfig1)
	machine2 := generateMachine(cluster, node2, kubeadmConfig2)
	machine2.ObjectMeta.Name = "machine02"
	infraMachine1 := generateAndSetInfraMachine(machine1)
	infraMachine2 := generateAndSetInfraMachine(machine2)
	nodeUpgrade1 := generateNodeUpgrade(machine1)
	nodeUpgrade2 := generateNodeUpgrade(machine2)
	nodeUpgrade2.ObjectMeta.Name = "node-upgrade-request-2"
	machines := []*clusterv1.Machine{machine1, machine2}
	return cpUpgradeObjects{
		cluster:        cluster,
		machines:       []*clusterv1.Machine{machine1, machine2},
		nodes:          []*corev1.Node{node1, node2},
		cpUpgrade:      generateCPUpgrade(machines),
		nodeUpgrades:   []*anywherev1.NodeUpgrade{nodeUpgrade1, nodeUpgrade2},
		kubeadmConfigs: []*bootstrapv1.KubeadmConfig{kubeadmConfig1, kubeadmConfig2},
		infraMachines:  []*tinkerbellv1.TinkerbellMachine{infraMachine1, infraMachine2},
	}
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
	kcpSpec, _ := json.Marshal(generateKcpSpec())
	return &anywherev1.ControlPlaneUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cp-upgrade-request",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.ControlPlaneUpgradeSpec{
			ControlPlane: corev1.ObjectReference{
				Name:      "my-cp",
				Namespace: constants.EksaSystemNamespace,
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
			KubernetesVersion:    k8s129,
			EtcdVersion:          etcd129,
			ControlPlaneSpecData: base64.StdEncoding.EncodeToString(kcpSpec),
		},
	}
}

func generateKcpSpec() *controlplanev1.KubeadmControlPlaneSpec {
	kcs := generateKubeadmConfig().Spec
	kcs.ClusterConfiguration.Etcd.Local.ImageTag = etcd129
	kcs.InitConfiguration = &bootstrapv1.InitConfiguration{}
	kcs.JoinConfiguration = &bootstrapv1.JoinConfiguration{}
	return &controlplanev1.KubeadmControlPlaneSpec{
		KubeadmConfigSpec: kcs,
		Version:           k8s129,
		RolloutStrategy: &controlplanev1.RolloutStrategy{
			Type: "InPlace",
		},
		Replicas: pointer.Int32(3),
		MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
			InfrastructureRef: corev1.ObjectReference{
				Name: "new-ref",
			},
		},
	}
}

func generateKubeadmConfig() *bootstrapv1.KubeadmConfig {
	return &bootstrapv1.KubeadmConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s%s", "kubeadm-config-", rand.String(10)),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: bootstrapv1.KubeadmConfigSpec{
			ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
				Etcd: bootstrapv1.Etcd{
					Local: &bootstrapv1.LocalEtcd{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageTag: etcd128,
						},
					},
				},
			},
			InitConfiguration: &bootstrapv1.InitConfiguration{},
		},
	}
}

func generateAndSetInfraMachine(machine *clusterv1.Machine) *tinkerbellv1.TinkerbellMachine {
	machine.Spec.InfrastructureRef = corev1.ObjectReference{
		Namespace:  machine.Namespace,
		Name:       machine.Name,
		Kind:       "TinkerbellMachine",
		APIVersion: tinkerbellv1.GroupVersion.String(),
	}
	return &tinkerbellv1.TinkerbellMachine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      machine.Name,
			Namespace: machine.Namespace,
			Annotations: map[string]string{
				"cluster.x-k8s.io/cloned-from-name": "old-ref",
			},
		},
	}
}
