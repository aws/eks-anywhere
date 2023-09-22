package hardware_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	rufiov1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestLoadHardwareSuccess(t *testing.T) {
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

	kubeReader := hardware.NewKubeReader(cl)
	err := kubeReader.LoadHardware(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(kubeReader.GetCatalogue().AllHardware())).To(Equal(1))
}

func TestLoadHardwareNoHardware(t *testing.T) {
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

	kubeReader := hardware.NewKubeReader(cl)
	err := kubeReader.LoadHardware(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(kubeReader.GetCatalogue().AllHardware())).To(Equal(0))
}

func TestLoadRufioMachinesSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	rm := rufiov1alpha1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bm1",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	scheme := runtime.NewScheme()
	_ = rufiov1alpha1.AddToScheme(scheme)
	objs := []runtime.Object{&rm}
	cb := fake.NewClientBuilder()
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	kubeReader := hardware.NewKubeReader(cl)
	err := kubeReader.LoadRufioMachines(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(kubeReader.GetCatalogue().AllBMCs())).To(Equal(1))
}

func TestLoadRufioMachinesListFail(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects().Build()

	kubeReader := hardware.NewKubeReader(cl)
	err := kubeReader.LoadRufioMachines(ctx)
	g.Expect(err).To(HaveOccurred())
}
