package hardware_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestHardwareFromETCDSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	hw := tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hw1",
			Namespace: constants.EksaSystemNamespace,
			Labels: map[string]string{
				"type": "cp",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Instance: &tinkv1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	}

	scheme := runtime.NewScheme()
	_ = tinkv1alpha1.AddToScheme(scheme)
	objs := []runtime.Object{&hw}
	cb := fake.NewClientBuilder()
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	etcdReader := hardware.NewETCDReader(cl)
	err := etcdReader.HardwareFromETCD(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(etcdReader.GetCatalogue().AllHardware())).To(Equal(1))
}

func TestHardwareFromETCDNoHardware(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	hw := tinkv1alpha1.Hardware{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hw1",
			Labels: map[string]string{
				hardware.OwnerNameLabel: "cluster",
				"type":                  "cp",
			},
		},
		Spec: tinkv1alpha1.HardwareSpec{
			Metadata: &tinkv1alpha1.HardwareMetadata{
				Instance: &tinkv1alpha1.MetadataInstance{
					ID: "foo",
				},
			},
		},
	}
	scheme := runtime.NewScheme()
	_ = tinkv1alpha1.AddToScheme(scheme)
	objs := []runtime.Object{&hw}
	cb := fake.NewClientBuilder()
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	etcdReader := hardware.NewETCDReader(cl)
	err := etcdReader.HardwareFromETCD(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(etcdReader.GetCatalogue().AllHardware())).To(Equal(0))
}
