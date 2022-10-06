package curatedpackages

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
)

var testBundle *packagesv1.PackageBundle = &packagesv1.PackageBundle{
	ObjectMeta: metav1.ObjectMeta{Name: "test-bundle", Namespace: packagesv1.PackageNamespace},
	TypeMeta: metav1.TypeMeta{
		Kind:       "PackageBundle",
		APIVersion: packagesv1.GroupVersion.String(),
	},
	Spec: packagesv1.PackageBundleSpec{
		Packages: []packagesv1.BundlePackage{
			{
				Name:   "test-package",
				Source: packagesv1.BundlePackageSource{},
			},
		},
	},
}

var testPBC *packagesv1.PackageBundleController = &packagesv1.PackageBundleController{
	ObjectMeta: metav1.ObjectMeta{Name: "test-cluster", Namespace: packagesv1.PackageNamespace},
	TypeMeta: metav1.TypeMeta{
		Kind:       "PackageBundleController",
		APIVersion: packagesv1.GroupVersion.String(),
	},
	Spec: packagesv1.PackageBundleControllerSpec{ActiveBundle: "test-bundle"},
}

const testClusterName string = "test-cluster"

func TestClusterActiveOrLatest(t *testing.T) {
	t.Parallel()
	initScheme()

	t.Run("golden path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(testPBC, testBundle).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		bundle, err := c.ActiveOrLatest(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(bundle, testBundle) {
			t.Fatalf("expected testBundle, got %+v", bundle)
		}
	})

	t.Run("handles client error getting PBC", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		bundle, err := c.ActiveOrLatest(ctx)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
		if bundle != nil {
			t.Fatalf("expected nil bundle, got %v", bundle)
		}
	})

	t.Run("handles client error getting Bundle", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(testPBC).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		bundle, err := c.ActiveOrLatest(ctx)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
		if bundle != nil {
			t.Fatalf("expected nil bundle, got %v", bundle)
		}
	})
}

var schemeOnce sync.Once

// initScheme adds thread safety to AddToScheme.
var initScheme = func() {
	schemeOnce.Do(func() {
		utilruntime.Must(packagesv1.AddToScheme(scheme.Scheme))
	})
}

func TestClusterAcitvePackagesBundleController(t *testing.T) {
	t.Parallel()
	initScheme()

	t.Run("golden path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(testPBC).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		pbc, err := c.ActivePackageBundleController(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(pbc, testPBC) {
			t.Fatalf("expected bundle to equal testBundle\n\n%+v\n\n%+v", pbc, testPBC)
		}
		if pbc.Spec.ActiveBundle != testPBC.Spec.ActiveBundle {
			t.Fatalf("epxected active bundle values to match, got %q and %q",
				pbc.Spec.ActiveBundle, testPBC.Spec.ActiveBundle)
		}
	})

	t.Run("handles client error getting PackageBundleController", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		pbc, err := c.ActivePackageBundleController(ctx)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
		if pbc != nil {
			t.Fatalf("expected nil pbc, got %v", pbc)
		}
	})
}

func TestClusterUpgradeBundle(t *testing.T) {
	t.Parallel()
	initScheme()

	t.Run("golden path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(testPBC).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		err := c.UpgradeBundle(ctx, "some-version")
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("handles a non-upgrade (new version == old version)", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			WithObjects(testPBC).
			Build()
		expectedBundleVersion := testPBC.Spec.ActiveBundle

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		err := c.UpgradeBundle(ctx, expectedBundleVersion)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("returns an error if the package bundle is empty", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		err := c.UpgradeBundle(ctx, "")
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("handles client error getting PackageBundleController", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		err := c.UpgradeBundle(ctx, "some-version")
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("handles client error updating PackageBundleController", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()
		fakeKubeClient := fake.NewClientBuilder().
			WithScheme(scheme.Scheme).
			Build()

		c := NewClusterBundleClient(fakeKubeClient, testClusterName)
		err := c.UpgradeBundle(ctx, "some-version")
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})
}

func TestNewBundleClient(t *testing.T) {
	t.Run("when source is Cluster, returns a clusterBundleClient", func(t *testing.T) {
		cfg := test.UseEnvTest(t)
		fakeKubeConfigFile := withFakeFileContents(t, bytes.NewBufferString(""))
		opts := BundleClientOptions{
			ClusterName:      "test-cluster",
			KubeConfig:       fakeKubeConfigFile.Name(),
			restConfigurator: restConfigurator(func(_ []byte) (*rest.Config, error) { return cfg, nil }),
		}

		c, err := NewBundleClient(Cluster, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if reflect.TypeOf(c) != reflect.TypeOf((*clusterBundleClient)(nil)) {
			t.Fatalf("Cluster source should create a clusterBundleClient")
		}
	})

	t.Run("when source is Registry, returns a registryBundleClient", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRegistryClient := mocks.NewMockRegistryClient(ctrl)
		opts := BundleClientOptions{
			KubeVersion:    "1.23",
			RegistryClient: mockRegistryClient,
		}
		c, err := NewBundleClient(Registry, opts)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if reflect.TypeOf(c) != reflect.TypeOf((*registryBundleClient)(nil)) {
			t.Fatalf("Registry source should create a registryBundleClient")
		}
	})

	t.Run("when source is Registry, KubeVersion is required", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		mockRegistryClient := mocks.NewMockRegistryClient(ctrl)
		opts := BundleClientOptions{
			RegistryClient: mockRegistryClient,
		}
		_, err := NewBundleClient(Registry, opts)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("with an invalid source, returns an error", func(t *testing.T) {
		opts := BundleClientOptions{}
		_, err := NewBundleClient(BundleSource("invalid"), opts)
		if errors.Is(err, nil) {
			t.Fatalf("expected an error, got nil")
		}
	})
}

func TestRegistryActiveOrLatest(t *testing.T) {
	t.Parallel()
	initScheme()

	t.Run("golden path", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ctrl := gomock.NewController(t)
		rc := mocks.NewMockRegistryClient(ctrl)
		rc.EXPECT().
			LatestBundle(ctx, "test-registry", "1.23").
			Return(testBundle, nil)
		c := NewRegistryBundleClient(rc, "test-registry", "1.23")
		bundle, err := c.ActiveOrLatest(ctx)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !reflect.DeepEqual(bundle, testBundle) {
			t.Fatalf("expected bundle to equal testBundle")
		}
	})

	t.Run("handles registry errors", func(t *testing.T) {
		t.Parallel()
		ctx := context.Background()

		ctrl := gomock.NewController(t)
		rc := mocks.NewMockRegistryClient(ctrl)
		rc.EXPECT().
			LatestBundle(ctx, "test-registry", "1.23").
			Return(nil, fmt.Errorf("test-error"))
		c := NewRegistryBundleClient(rc, "test-registry", "1.23")
		_, err := c.ActiveOrLatest(ctx)
		if errors.Is(err, nil) {
			t.Fatalf("expected error, got nil")
		}
	})
}

// withFakeFile returns a throwaway file in a test-specific directory.
//
// The file is automatically closed and removed when the test ends.
func withFakeFile(t *testing.T) (f *os.File) {
	f, err := ioutil.TempFile(t.TempDir(), "fake-file")
	if err != nil {
		t.Fatalf("opening temp file: %s", err)
	}
	t.Cleanup(func() { f.Close() })

	return f
}

// withFakeFileContents returns a file containing some data.
//
// The file is automatically closed and removed when the test ends.
func withFakeFileContents(t *testing.T, r io.Reader) (f *os.File) {
	f = withFakeFile(t)
	_, err := io.Copy(f, r)
	if err != nil {
		t.Fatalf("copying contents into fake file %q: %s", f.Name(), err)
	}

	return f
}
