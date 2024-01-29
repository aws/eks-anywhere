package controllers_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
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
	k8s128  = "v1.28.3-eks-1-28-9"
	k8s129  = "v1.29.0-eks-1-29-0"
)

func TestCPUpgradeReconcile(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}

	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileEarly(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	cpUpgrade.Status.Ready = true
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeNotUpgraded(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: false,
		}
	}
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeUpgradeError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, _, kubeadmConfigs := getObjectsForCPUpgradeTest()

	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, kubeadmConfigs[0], kubeadmConfigs[1]}
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
	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	cpu := &anywherev1.ControlPlaneUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: cpUpgrade.Name, Namespace: constants.EksaSystemNamespace}, cpu)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestCPUpgradeReconcileNodeUpgraderInvalidKCPSpec(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
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
			cpUpgrade.Spec.ControlPlaneSpecData = test.kcpSpec
			objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], cpUpgrade, nodeUpgrades[0], kubeadmConfigs[0], kubeadmConfigs[1]}
			client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
			r := controllers.NewControlPlaneUpgradeReconciler(client)
			req := cpUpgradeRequest(cpUpgrade)
			_, err := r.Reconcile(ctx, req)
			g.Expect(err).To(HaveOccurred())
			g.Expect(err.Error()).To(ContainSubstring(test.errMsg))
		})
	}
}

func TestCPUpgradeReconcileNodesNotReadyYet(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
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
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
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

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	cpUpgrade.DeletionTimestamp = &now
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)
	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	n := &anywherev1.NodeUpgrade{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Name, Namespace: constants.EksaSystemNamespace}, n)
	g.Expect(err).To(MatchError("nodeupgrades.anywhere.eks.amazonaws.com \"machine01-node-upgrader\" not found"))
}

func TestCPUpgradeObjectDoesNotExist(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(MatchError("controlplaneupgrades.anywhere.eks.amazonaws.com \"cp-upgrade-request\" not found"))
}

func TestCPUpgradeReconcileUpdateCapiMachineVersion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
	machine := &clusterv1.Machine{}
	err = client.Get(ctx, types.NamespacedName{Name: nodeUpgrades[0].Spec.Machine.Name, Namespace: constants.EksaSystemNamespace}, machine)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(*machine.Spec.Version).To(BeEquivalentTo(k8s128))
}

func TestCPUpgradeReconcileUpdateKubeadmConfigSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	kcpDecoded, err := base64.StdEncoding.DecodeString(cpUpgrade.Spec.ControlPlaneSpecData)
	g.Expect(err).ToNot(HaveOccurred())

	kcpSpec := &controlplanev1.KubeadmControlPlaneSpec{}
	err = json.Unmarshal(kcpDecoded, kcpSpec)
	g.Expect(err).ToNot(HaveOccurred())

	for i := range kubeadmConfigs {
		kc := kubeadmConfigs[i]
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

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, kubeadmConfigs := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	for i := range machines {
		machines[i].Spec.Bootstrap.ConfigRef = nil
	}
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1], kubeadmConfigs[0], kubeadmConfigs[1]}
	nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("updating kubeadm config: bootstrap config for machine machine01 is nil"))
}

func TestCPUpgradeReconcileUpdateKubeadmConfigNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster, machines, nodes, cpUpgrade, nodeUpgrades, _ := getObjectsForCPUpgradeTest()
	for i := range nodeUpgrades {
		nodeUpgrades[i].Name = fmt.Sprintf("%s-node-upgrader", machines[i].Name)
		nodeUpgrades[i].Status = anywherev1.NodeUpgradeStatus{
			Completed: true,
		}
	}
	objs := []runtime.Object{cluster, cpUpgrade, machines[0], machines[1], nodes[0], nodes[1], nodeUpgrades[0], nodeUpgrades[1]}
	nodeUpgrades[0].Status.Completed = true
	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).Build()
	r := controllers.NewControlPlaneUpgradeReconciler(client)

	req := cpUpgradeRequest(cpUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(MatchRegexp("updating kubeadm config: retrieving bootstrap config for machine machine01: kubeadmconfigs.bootstrap.cluster.x-k8s.io \"kubeadm-config-\\w{10}\" not found"))
}

func getObjectsForCPUpgradeTest() (*clusterv1.Cluster, []*clusterv1.Machine, []*corev1.Node, *anywherev1.ControlPlaneUpgrade, []*anywherev1.NodeUpgrade, []*bootstrapv1.KubeadmConfig) {
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
	nodeUpgrade2 := generateNodeUpgrade(machine2)
	nodeUpgrade2.ObjectMeta.Name = "node-upgrade-request-2"
	machines := []*clusterv1.Machine{machine1, machine2}
	nodes := []*corev1.Node{node1, node2}
	nodeUpgrades := []*anywherev1.NodeUpgrade{nodeUpgrade1, nodeUpgrade2}
	cpUpgrade := generateCPUpgrade(machines)
	return cluster, machines, nodes, cpUpgrade, nodeUpgrades, []*bootstrapv1.KubeadmConfig{kubeadmConfig1, kubeadmConfig2}
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
