package envtest

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/mrajashree/etcdadm-controller/api/v1beta1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta2"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	capiPackage = "sigs.k8s.io/cluster-api"
	capvPackage = "sigs.k8s.io/cluster-api-provider-vsphere"
)

func init() {
	// Register CRDs in Scheme in init so fake clients benefit from it
	utilruntime.Must(releasev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(clusterctlv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(controlplanev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(vspherev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(cloudstackv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(etcdv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme.Scheme))
	utilruntime.Must(anywherev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(eksdv1alpha1.AddToScheme(scheme.Scheme))
}

var packages = mustBuildModulesWithCRDs(capiPackage, capvPackage)

type Environment struct {
	scheme  *runtime.Scheme
	client  client.Client
	env     *envtest.Environment
	manager manager.Manager
	// apiReader is a non cached client (only for reads), helpful when testing the actual state of objects
	apiReader client.Reader
	cancelF   context.CancelFunc
}

type EnvironmentOpt func(ctx context.Context, e *Environment)

func WithAssignment(envRef **Environment) EnvironmentOpt {
	return func(ctx context.Context, e *Environment) {
		*envRef = e
	}
}

// RunWithEnvironment runs a suite of tests with an envtest that is shared across all tests
// We use testing.M as the input to avoid having this called directly from a test
// This ensures the envtest setup is always run from a TestMain
func RunWithEnvironment(m *testing.M, opts ...EnvironmentOpt) int {
	ctx := ctrl.SetupSignalHandler()
	env, err := newEnvironment(ctx)
	if err != nil {
		fmt.Printf("Failed setting up envtest: %s\n", err)
		return 1
	}

	for _, o := range opts {
		o(ctx, env)
	}

	returnCode := m.Run()

	if err = env.stop(); err != nil {
		fmt.Printf("Failed stopping envtest: %s", err)
		return 1
	}

	return returnCode
}

func newEnvironment(ctx context.Context) (*Environment, error) {
	root := getRootPath()
	crdDirectoryPaths := make([]string, 0, len(packages)+1)
	crdDirectoryPaths = append(crdDirectoryPaths, filepath.Join(root, "config", "crd", "bases"))
	extraCRDPaths, err := getPathsToPackagesCRDs(root, packages...)
	if err != nil {
		return nil, err
	}
	crdDirectoryPaths = append(crdDirectoryPaths, extraCRDPaths...)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     crdDirectoryPaths,
		ErrorIfCRDPathMissing: true,
		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join(root, "config", "webhook")},
		},
	}

	scheme := scheme.Scheme
	ctx, cancel := context.WithCancel(ctx)
	env := &Environment{
		env:     testEnv,
		scheme:  scheme,
		cancelF: cancel,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		return nil, err
	}

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:             scheme,
		Host:               webhookInstallOptions.LocalServingHost,
		Port:               webhookInstallOptions.LocalServingPort,
		CertDir:            webhookInstallOptions.LocalServingCertDir,
		LeaderElection:     false,
		MetricsBindAddress: "0",
	})
	if err != nil {
		return nil, err
	}
	env.manager = mgr

	go func() {
		err = mgr.Start(ctx)
	}()
	<-mgr.Elected()
	if err != nil {
		return nil, err
	}

	env.client = mgr.GetClient()
	env.apiReader = mgr.GetAPIReader()

	return env, nil
}

func (e *Environment) stop() error {
	fmt.Println("Stopping the test environment")
	e.cancelF() // Cancels context that will stop the manager
	return e.env.Stop()
}

func (e *Environment) Client() client.Client {
	return e.client
}

// APIReader returns a non cached reader client
func (e *Environment) APIReader() client.Reader {
	return e.apiReader
}

// Manager returns a Manager for the test environment
func (e *Environment) Manager() manager.Manager {
	return e.manager
}

func (e *Environment) CreateNamespaceForTest(ctx context.Context, t *testing.T) string {
	t.Helper()
	name := strings.ReplaceAll(t.Name(), "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ToLower(name)
	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := e.client.Create(ctx, namespace); err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		if err := e.client.Delete(ctx, namespace); err != nil && !apierrors.IsNotFound(err) {
			t.Fatal(err)
		}
	})

	return namespace.Name
}

func getRootPath() string {
	_, currentFilePath, _, _ := goruntime.Caller(0)
	return path.Join(path.Dir(currentFilePath), "..", "..", "..")
}
