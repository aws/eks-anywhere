package hardware_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	rufioalphav1 "github.com/tinkerbell/rufio/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

func TestNewCatalogueFromETCDSuccess(t *testing.T) {
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
	err := etcdReader.NewCatalogueFromETCD(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(etcdReader.GetCatalogue().AllHardware())).To(Equal(1))
}

func TestNewCatalogueFromETCDNoHardware(t *testing.T) {
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
	err := etcdReader.NewCatalogueFromETCD(ctx)
	g.Expect(err).To(HaveOccurred())
	g.Expect(err.Error()).To(ContainSubstring("no available hardware"))
}

func TestNewMachineCatalogueFromETCDSuccess(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	hw := &tinkv1alpha1.Hardware{
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
			BMCRef: &corev1.TypedLocalObjectReference{
				Name: "bmc-hw1",
				Kind: "Machine",
			},
		},
	}
	machine := &rufioalphav1.Machine{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc-hw1",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: rufioalphav1.MachineSpec{
			Connection: rufioalphav1.Connection{
				AuthSecretRef: corev1.SecretReference{
					Name:      "bmc-hw1-auth-secret",
					Namespace: constants.EksaSystemNamespace,
				},
			},
		},
	}
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "bmc-hw1-auth-secret",
			Namespace: constants.EksaSystemNamespace,
		},
	}

	scheme := runtime.NewScheme()
	_ = tinkv1alpha1.AddToScheme(scheme)
	_ = rufioalphav1.AddToScheme(scheme)
	_ = corev1.AddToScheme(scheme)

	objs := []runtime.Object{hw, machine, secret}
	cb := fake.NewClientBuilder()
	cl := cb.WithScheme(scheme).WithRuntimeObjects(objs...).Build()

	etcdReader := hardware.NewETCDReader(cl)
	err := etcdReader.NewMachineCatalogueFromETCD(ctx)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(len(etcdReader.GetCatalogue().AllHardware())).To(Equal(1))
	g.Expect(len(etcdReader.GetCatalogue().AllBMCs())).To(Equal(1))
	g.Expect(len(etcdReader.GetCatalogue().AllSecrets())).To(Equal(1))
}
