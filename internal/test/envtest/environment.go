package envtest

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	admissionv1beta1 "k8s.io/api/admission/v1beta1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/api/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	clusterctlv1 "sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	rufiov1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	capiPackage         = "sigs.k8s.io/cluster-api"
	capdPackage         = "sigs.k8s.io/cluster-api/test"
	capvPackage         = "sigs.k8s.io/cluster-api-provider-vsphere"
	captPackage         = "github.com/tinkerbell/cluster-api-provider-tinkerbell"
	tinkerbellPackage   = "github.com/tinkerbell/tink"
	etcdProviderPackage = "github.com/aws/etcdadm-controller"
	capcPackage         = "sigs.k8s.io/cluster-api-provider-cloudstack"

	kubebuilderAssetsEnvVar = "KUBEBUILDER_ASSETS"
)

func init() {
	// Register CRDs in Scheme in init so fake clients benefit from it
	utilruntime.Must(corev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(releasev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(clusterv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(clusterctlv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(controlplanev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(bootstrapv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(vspherev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(dockerv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(cloudstackv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(etcdv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(admissionv1beta1.AddToScheme(scheme.Scheme))
	utilruntime.Must(anywherev1.AddToScheme(scheme.Scheme))
	utilruntime.Must(eksdv1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(snowv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(addonsv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(tinkerbellv1.AddToScheme(scheme.Scheme))
	utilruntime.Must(tinkv1alpha1.AddToScheme(scheme.Scheme))
	utilruntime.Must(rufiov1alpha1.AddToScheme(scheme.Scheme))
}

var packages = []moduleWithCRD{
	mustBuildModuleWithCRDs(capiPackage,
		withAdditionalCustomCRDPath("bootstrap/kubeadm/config/crd/bases"),
		withAdditionalCustomCRDPath("controlplane/kubeadm/config/crd/bases"),
	),
	mustBuildModuleWithCRDs(captPackage,
		withMainCustomCRDPath("config/crd/bases/infrastructure.cluster.x-k8s.io_tinkerbellclusters.yaml"),
		withAdditionalCustomCRDPath("config/crd/bases/infrastructure.cluster.x-k8s.io_tinkerbellmachinetemplates.yaml")),
	mustBuildModuleWithCRDs(tinkerbellPackage),
	mustBuildModuleWithCRDs(capvPackage),
	mustBuildModuleWithCRDs(capdPackage,
		withMainCustomCRDPath("infrastructure/docker/config/crd/bases"),
	),
	mustBuildModuleWithCRDs(etcdProviderPackage),
	mustBuildModuleWithCRDs(capcPackage),
}

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
// This ensures the envtest setup is always run from a TestMain.
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
	currentDir := currentDir()

	if err := ensureEnvtest(ctx, root); err != nil {
		return nil, err
	}

	crdDirectoryPaths := make([]string, 0, len(packages)+2)
	crdDirectoryPaths = append(crdDirectoryPaths,
		filepath.Join(root, "config", "crd", "bases"),
		filepath.Join(currentDir, "config", "eks-d-crds.yaml"),
		filepath.Join(currentDir, "config", "snow-crds.yaml"),
		filepath.Join(currentDir, "config", "rufio-crds.yaml"),
	)
	extraCRDPaths, err := getPathsToPackagesCRDs(root, packages...)
	if err != nil {
		return nil, err
	}
	crdDirectoryPaths = append(crdDirectoryPaths, extraCRDPaths...)

	testEnv := &envtest.Environment{
		CRDDirectoryPaths:     crdDirectoryPaths,
		ErrorIfCRDPathMissing: true,
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

func ensureEnvtest(ctx context.Context, rootDir string) error {
	// Only if the envtest config envvar is not set, try to setup assets
	if _, ok := os.LookupEnv(kubebuilderAssetsEnvVar); ok {
		return nil
	}

	cmd := exec.CommandContext(ctx, "make", "-s", "envtest-setup")
	cmd.Dir = rootDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed setting up env-test:\n%s", string(out))
	}

	// Last line of output is expected to be KUBEBUILDER_ASSETS=[path to envtest setup]
	lines := bytes.Split(out, []byte("\n"))
	lastLine := lines[len(lines)-2]
	split := bytes.Split(lastLine, []byte("="))
	if len(split) != 2 || string(split[0]) != kubebuilderAssetsEnvVar {
		return fmt.Errorf("invalid last line of env-test setup: %s", string(lastLine))
	}

	fmt.Printf("Envtest auto-setup using installation path %s\n", string(split[1]))
	os.Setenv(kubebuilderAssetsEnvVar, string(split[1]))

	return nil
}

func (e *Environment) stop() error {
	fmt.Println("Stopping the test environment")
	e.cancelF() // Cancels context that will stop the manager
	return e.env.Stop()
}

func (e *Environment) Client() client.Client {
	return e.client
}

// APIReader returns a non cached reader client.
func (e *Environment) APIReader() client.Reader {
	return e.apiReader
}

// Manager returns a Manager for the test environment.
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
	return path.Join(currentDir(), "..", "..", "..")
}

func currentDir() string {
	_, currentFilePath, _, _ := goruntime.Caller(0)
	return path.Dir(currentFilePath)
}
