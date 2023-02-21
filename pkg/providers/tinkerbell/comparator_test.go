package tinkerbell_test

import (
	"testing"

	. "github.com/onsi/gomega"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

func TestKubeadmControlPlaneReplicas(t *testing.T) {
	g := NewWithT(t)

	KCPComparator := &tinkerbell.KubeadmControlPlane{comparatorKubeadmControlPlane()}

	g.Expect(KCPComparator.Replicas()).To(Equal(1))
}

func TestKubeadmControlPlaneVersion(t *testing.T) {
	g := NewWithT(t)

	KCPComparator := &tinkerbell.KubeadmControlPlane{comparatorKubeadmControlPlane()}

	g.Expect(KCPComparator.Version()).To(Equal("1.22"))
}

func TestMachineDeploymentReplicas(t *testing.T) {
	g := NewWithT(t)

	MDComparator := &tinkerbell.MachineDeployment{comparatorMachineDeployment()}

	g.Expect(MDComparator.Replicas()).To(Equal(1))
}

func TestKubernetesVersionChangeFalse(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparable{version: "1.22"}
	newComparator := &testComparable{version: "1.22"}

	g.Expect(tinkerbell.HasKubernetesVersionChange(oldComparator, newComparator)).To(Equal(false))
}

func TestKubernetesVersionChangeTrue(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparable{version: "1.22"}
	newComparator := &testComparable{version: "1.23"}

	g.Expect(tinkerbell.HasKubernetesVersionChange(oldComparator, newComparator)).To(Equal(true))
}

func TestReplicasChangeUnchanged(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparable{replicas: 1}
	newComparator := &testComparable{replicas: 1}

	g.Expect(tinkerbell.ReplicasDiff(oldComparator, newComparator)).To(Equal(0))
}

func TestReplicasChangeChanged(t *testing.T) {
	g := NewWithT(t)

	oldComparator := &testComparable{replicas: 1}
	newComparator := &testComparable{replicas: 3}

	g.Expect(tinkerbell.ReplicasDiff(oldComparator, newComparator)).To(Equal(2))
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

type testComparable struct {
	version  string
	replicas int
}

func (t *testComparable) Version() string {
	return t.version
}

func (t *testComparable) Replicas() int {
	return t.replicas
}
