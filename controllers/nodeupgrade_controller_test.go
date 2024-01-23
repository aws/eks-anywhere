package controllers_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	upgrader "github.com/aws/eks-anywhere/pkg/nodeupgrader"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestNodeUpgradeReconcilerReconcileFirstControlPlane(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	nodeUpgrade.Spec.FirstNodeToBeUpgraded = true
	nodeUpgrade.Spec.EtcdVersion = ptr.String("v3.5.9-eks-1-28-9")
	node.Labels = map[string]string{
		"node-role.kubernetes.io/control-plane": "true",
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestNodeUpgradeReconcilerReconcileNextControlPlane(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	node.Labels = map[string]string{
		"node-role.kubernetes.io/control-plane": "true",
	}
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestNodeUpgradeReconcilerReconcileWorker(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestNodeUpgradeReconcilerReconcileCreateUpgraderPodState(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil).Times(2)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	g.Expect(client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)).To(Succeed())

	statuses := []corev1.ContainerStatus{
		{
			Name: upgrader.CopierContainerName,
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					ExitCode: 0,
				},
			},
		},
		{
			Name: upgrader.ContainerdUpgraderContainerName,
			State: corev1.ContainerState{
				Running: &corev1.ContainerStateRunning{},
			},
		},
		{
			Name: upgrader.CNIPluginsUpgraderContainerName,
			State: corev1.ContainerState{
				Waiting: &corev1.ContainerStateWaiting{},
			},
		},
		{
			Name: upgrader.KubeadmUpgraderContainerName,
			State: corev1.ContainerState{
				Terminated: &corev1.ContainerStateTerminated{
					ExitCode: 1,
				},
			},
		},
		{
			Name:  upgrader.KubeletUpgradeContainerName,
			State: corev1.ContainerState{},
		},
	}

	pod.Status.InitContainerStatuses = append(pod.Status.InitContainerStatuses, statuses...)
	g.Expect(client.Update(ctx, pod)).To(Succeed())

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())
}

func TestNodeUpgradeReconcilerReconcileDelete(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil).Times(2)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).ToNot(HaveOccurred())

	err = client.Delete(ctx, nodeUpgrade)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod = &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).To(MatchError("pods \"node01-node-upgrader\" not found"))
}

func TestNodeUpgradeReconcilerReconcileDeleteUpgraderPodAlreadyDeleted(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	clientRegistry := mocks.NewMockRemoteClientRegistry(ctrl)

	cluster, machine, node, nodeUpgrade, configMap := getObjectsForNodeUpgradeTest()
	client := fake.NewClientBuilder().WithRuntimeObjects(cluster, machine, node, nodeUpgrade, configMap).Build()

	clientRegistry.EXPECT().GetClient(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}).Return(client, nil).Times(2)

	r := controllers.NewNodeUpgradeReconciler(client, clientRegistry)
	req := nodeUpgradeRequest(nodeUpgrade)
	_, err := r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod := &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).ToNot(HaveOccurred())

	err = client.Delete(ctx, nodeUpgrade)
	g.Expect(err).ToNot(HaveOccurred())

	err = client.Delete(ctx, pod)
	g.Expect(err).ToNot(HaveOccurred())

	_, err = r.Reconcile(ctx, req)
	g.Expect(err).ToNot(HaveOccurred())

	pod = &corev1.Pod{}
	err = client.Get(ctx, types.NamespacedName{Name: upgrader.PodName(node.Name), Namespace: "eksa-system"}, pod)
	g.Expect(err).To(MatchError("pods \"node01-node-upgrader\" not found"))
}

func getObjectsForNodeUpgradeTest() (*clusterv1.Cluster, *clusterv1.Machine, *corev1.Node, *anywherev1.NodeUpgrade, *corev1.ConfigMap) {
	cluster := generateCluster()
	node := generateNode()
	machine := generateMachine(cluster, node)
	nodeUpgrade := generateNodeUpgrade(machine)
	configMap := generateConfigMap()
	return cluster, machine, node, nodeUpgrade, configMap
}

func nodeUpgradeRequest(nodeUpgrade *anywherev1.NodeUpgrade) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      nodeUpgrade.Name,
			Namespace: nodeUpgrade.Namespace,
		},
	}
}

func generateNodeUpgrade(machine *clusterv1.Machine) *anywherev1.NodeUpgrade {
	return &anywherev1.NodeUpgrade{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-upgrade-request",
			Namespace: "eksa-system",
		},
		Spec: anywherev1.NodeUpgradeSpec{
			Machine: corev1.ObjectReference{
				Name:      machine.Name,
				Namespace: machine.Namespace,
			},
			KubernetesVersion: "v1.28.3-eks-1-28-9",
		},
	}
}

func generateMachine(cluster *clusterv1.Cluster, node *corev1.Node) *clusterv1.Machine {
	return &clusterv1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "machine01",
			Namespace: "eksa-system",
		},
		Spec: clusterv1.MachineSpec{
			Version:     ptr.String("v1.27.8-eks-1-27-18"),
			ClusterName: cluster.Name,
		},
		Status: clusterv1.MachineStatus{
			NodeRef: &corev1.ObjectReference{
				Name: node.Name,
			},
		},
	}
}

func generateNode() *corev1.Node {
	return &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "node01",
		},
	}
}

func generateCluster() *clusterv1.Cluster {
	return &clusterv1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "eksa-system",
		},
	}
}

func generateConfigMap() *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "in-place-upgrade",
			Namespace: "eksa-system",
		},
		Data: map[string]string{"v1.28.3-eks-1-28-9": "test"},
	}
}
