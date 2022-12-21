package clientutil_test

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
)

func TestObjectsToClientObjectsKubernetesObjects(t *testing.T) {
	g := NewWithT(t)
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-1",
		},
	}
	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-2",
		},
	}
	objs := []kubernetes.Object{cm1, cm2}

	g.Expect(clientutil.ObjectsToClientObjects(objs)).To(ConsistOf(cm1, cm2))
}

func TestObjectsToClientObjectsEnvtestObjects(t *testing.T) {
	g := NewWithT(t)
	cm1 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-1",
		},
	}
	cm2 := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cm-2",
		},
	}
	objs := []client.Object{cm1, cm2}

	g.Expect(clientutil.ObjectsToClientObjects(objs)).To(ConsistOf(cm1, cm2))
}
