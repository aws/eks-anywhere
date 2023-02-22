package tinkerbell_test

import (
	"testing"

	. "github.com/onsi/gomega"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestKubeadmControlPlaneGetReplicas(t *testing.T) {
	g := NewWithT(t)

	KCPComparator := &tinkerbell.KubeadmControlPlaneComparator{comparatorKubeadmControlPlane()}

	g.Expect(KCPComparator.GetReplicas(), 1)
}

func TestKubeadmControlPlaneGetVersion(t *testing.T) {
	g := NewWithT(t)

	KCPComparator := &tinkerbell.KubeadmControlPlaneComparator{comparatorKubeadmControlPlane()}

	g.Expect(KCPComparator.GetVersion(), "1.22")
}

func TestMachineDeploymentGetReplicas(t *testing.T) {
	g := NewWithT(t)

	MDComparator := &tinkerbell.MachineDeploymentComparator{comparatorMachineDeployment()}

	g.Expect(MDComparator.GetReplicas(), 1)
}

func TestKubernetesVersionChangeFalse(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparator{version: "1.22"}
	newComparator := &testComparator{version: "1.22"}

	g.Expect(tinkerbell.KubernetesVersionChange(oldComparator, newComparator), false)
}

func TestKubernetesVersionChangeTrue(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparator{version: "1.22"}
	newComparator := &testComparator{version: "1.23"}

	g.Expect(tinkerbell.KubernetesVersionChange(oldComparator, newComparator), true)
}

func TestReplicasChangeUnchanged(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparator{replicas: 1}
	newComparator := &testComparator{replicas: 1}

	g.Expect(tinkerbell.ReplicasChange(oldComparator, newComparator), 0)
}

func TestReplicasChangeChanged(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparator{replicas: 1}
	newComparator := &testComparator{replicas: 3}

	g.Expect(tinkerbell.ReplicasChange(oldComparator, newComparator), 2)
}

func comparatorKubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	k := &controlplanev1.KubeadmControlPlane{
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			Replicas: ptr.Int32(1),
			Version:  "1.22",
		},
	}

	return k
}

func comparatorMachineDeployment() *clusterv1.MachineDeployment {
	m := &clusterv1.MachineDeployment{
		Spec: clusterv1.MachineDeploymentSpec{
			Replicas: ptr.Int32(1),
		},
	}

	return m
}

type testComparator struct {
	version  string
	replicas int
}

func (t *testComparator) GetVersion() string {
	return t.version
}

func (t *testComparator) GetReplicas() int {
	return t.replicas
}
