package curatedpackages_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest/fake"
	"sigs.k8s.io/yaml"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/curatedpackages/mocks"
	"github.com/aws/eks-anywhere/pkg/version"
)

type bundleTest struct {
	*WithT
	ctx            context.Context
	kubeConfig     string
	kubeVersion    string
	kubeRESTClient *fake.RESTClient
	kubectl        *mocks.MockKubectlRunner
	bundleManager  *mocks.MockManager
	Command        *curatedpackages.BundleReader
	activeBundle   string
	bundleCtrl     *packagesv1.PackageBundleController
	packageBundle  *packagesv1.PackageBundle
	registry       *mocks.MockBundleRegistry
	cliVersion     version.Info
	activeCluster  string
}

func newBundleTest(t *testing.T) *bundleTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlRunner(ctrl)
	fakeRestClient := fakeRESTClient()
	bm := mocks.NewMockManager(ctrl)
	kubeConfig := "test.kubeconfig"
	kubeVersion := "1.21"
	registry := mocks.NewMockBundleRegistry(ctrl)
	activeBundle := "v1.21-1000"
	activeCluster := "test-cluster"
	cliVersion := version.Info{GitVersion: "v1.0.0"}
	bundleCtrl := packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			ActiveBundle: activeBundle,
		},
	}
	packageBundle := packagesv1.PackageBundle{
		Spec: packagesv1.PackageBundleSpec{
			Packages: []packagesv1.BundlePackage{
				{
					Name: "harbor",
				},
			},
		},
	}

	return &bundleTest{
		WithT:          NewWithT(t),
		ctx:            context.Background(),
		kubeConfig:     kubeConfig,
		kubeVersion:    kubeVersion,
		kubeRESTClient: fakeRestClient,
		kubectl:        k,
		bundleManager:  bm,
		bundleCtrl:     &bundleCtrl,
		packageBundle:  &packageBundle,
		activeBundle:   activeBundle,
		registry:       registry,
		cliVersion:     cliVersion,
		activeCluster:  activeCluster,
	}
}

func TestGetLatestBundleFromClusterSucceeds(t *testing.T) {
	tt := newBundleTest(t)

	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Cluster,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)

	// Three requests are expected:
	//     1st- a request to find the name of the packages bundle controller
	//     2nd- a request for the packages bundle controller
	//     3rd- the active bundle references by the packages bundle controller
	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusOK, mockBodyActiveCluster(tt.activeCluster)),
		fakeResponse(http.StatusOK, mockBodyPBC("v1-22-33")),
		fakeResponse(http.StatusOK, mockBodyPackageBundle("harbor")),
	)
	result, err := tt.Command.GetLatestBundle(tt.ctx, "v1-22-33")
	tt.Expect(err).To(BeNil())
	tt.Expect(result.Spec.Packages[0].Name).To(BeEquivalentTo(tt.packageBundle.Spec.Packages[0].Name))
}

func TestGetLatestBundleFromRegistrySucceeds(t *testing.T) {
	tt := newBundleTest(t)
	baseRef := "test_host/test_env/test_controller"
	tt.registry.EXPECT().GetRegistryBaseRef(tt.ctx).Return(baseRef, nil)
	tt.bundleManager.EXPECT().LatestBundle(tt.ctx, baseRef, tt.kubeVersion).Return(tt.packageBundle, nil)
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Registry,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)
	result, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(BeNil())
	tt.Expect(result.Spec.Packages[0].Name).To(BeEquivalentTo(tt.packageBundle.Spec.Packages[0].Name))
}

func TestGetLatestBundleFromUnknownSourceFails(t *testing.T) {
	tt := newBundleTest(t)
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		"Unknown",
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)
	_, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("unknown source")))
}

func TestLatestBundleFromClusterUnknownBundle(t *testing.T) {
	tt := newBundleTest(t)
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Cluster,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)

	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusOK, mockBodyKubeadmControlPlane(tt.activeCluster)),
		fakeResponse(http.StatusOK, mockBodyPBC("v1-22-33")),
		fakeResponse(http.StatusInternalServerError, mockBodyf("error reading bundle")),
	)
	_, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("error reading bundle")))

	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusOK, mockBodyKubeadmControlPlane(tt.activeCluster)),
		fakeResponse(http.StatusOK, mockBodyPBC("v1-22-33")),
		fakeResponse(http.StatusNotFound, nil),
	)
	_, err = tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("the server could not find the requested resource")))
}

func TestGetLatestBundleFromRegistryWhenError(t *testing.T) {
	tt := newBundleTest(t)
	tt.registry.EXPECT().GetRegistryBaseRef(tt.ctx).Return("", errors.New("registry doesn't exist"))
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Registry,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)
	_, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("registry doesn't exist")))
}

func TestLatestBundleFromClusterUnknownCtrl(t *testing.T) {
	tt := newBundleTest(t)
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Cluster,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)
	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusOK, mockBodyKubeadmControlPlane(tt.activeCluster)),
		fakeResponse(http.StatusInternalServerError, nil),
	)
	_, err := tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("getting package bundle controller")))
	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusNotFound, nil),
	)
	_, err = tt.Command.GetLatestBundle(tt.ctx, tt.kubeVersion)
	tt.Expect(err).To(MatchError(ContainSubstring("could not find the requested resource")))
}

func TestUpgradeBundleSucceeds(t *testing.T) {
	tt := newBundleTest(t)
	newBundle := "new-bundle"

	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Cluster,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)

	copy := tt.bundleCtrl.DeepCopy()
	copy.Spec.ActiveBundle = newBundle
	resp, err := yaml.Marshal(copy)
	tt.Expect(err).To(BeNil())
	tt.kubeRESTClient.GroupVersion = packagesv1.GroupVersion
	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusOK,
			io.NopCloser(bytes.NewBuffer(resp))),
	)
	err = tt.Command.UpgradeBundle(tt.ctx, tt.bundleCtrl, newBundle)
	tt.Expect(err).To(BeNil())
}

func TestUpgradeBundleFails(t *testing.T) {
	tt := newBundleTest(t)
	newBundle := "new-bundle"
	tt.Command = curatedpackages.NewBundleReader(
		tt.kubeConfig,
		curatedpackages.Cluster,
		tt.kubeRESTClient,
		tt.bundleManager,
		tt.registry,
	)
	tt.kubeRESTClient.Client = fakedResponses(
		fakeResponse(http.StatusInternalServerError, nil),
	)

	err := tt.Command.UpgradeBundle(tt.ctx, tt.bundleCtrl, newBundle)
	tt.Expect(err).NotTo(BeNil())
}

func convertJsonToBytes(obj interface{}) bytes.Buffer {
	b, _ := json.Marshal(obj)
	return *bytes.NewBuffer(b)
}

func fakeRESTClient() *fake.RESTClient {
	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(packagesv1.AddToScheme(scheme))

	return &fake.RESTClient{
		GroupVersion:         packagesv1.GroupVersion,
		NegotiatedSerializer: serializer.NewCodecFactory(scheme),
	}
}

// fakedResponses is a simple http.RoundTripper.
//
// It simply returns each response in sequence. This is an attempt to remain
// resilient in the face of changing paths or names.  It will panic if you ask
// for more responses than it has. This is considered a feature.
func fakedResponses(responses ...*http.Response) *http.Client {
	idx := 0
	return fake.CreateHTTPClient(func(r *http.Request) (*http.Response, error) {
		defer func() { idx++ }()
		return responses[idx], nil
	})
}

// fakeResponse helps build *http.Response objects.
func fakeResponse(code int, body io.ReadCloser) *http.Response {
	return &http.Response{
		StatusCode: code,
		Body:       body,
	}
}

// mockBodyPBC builds an http response body containing a minimal packages
// bundle controller with the provided active bundle.
func mockBodyPBC(activeBundleName string) io.ReadCloser {
	jsonSafe := "unknown"
	data, err := json.Marshal(activeBundleName)
	if err != nil {
		log.Printf("failed to marshal active bundle name, using default: %s", err)
	} else {
		jsonSafe = string(data)
	}

	return mockBodyf(`
{
  "apiVersion": "packages.eks.amazonaws.com/v1alpha1",
  "kind": "PackageBundleController",
  "spec": {
    "activeBundle": %s
  }
}
`, jsonSafe)
}

// mockBodyf wraps a string body into a io.ReadCloser for use as an
// http.Response.Body.
func mockBodyf(template string, args ...interface{}) io.ReadCloser {
	return io.NopCloser(bytes.NewBufferString(fmt.Sprintf(template, args...)))
}

// mockBodyPackageBundle builds an http response body containing a minimal
// package bundle with the provided package names.
func mockBodyPackageBundle(packageNames ...string) io.ReadCloser {
	list := []struct {
		Name string `json:"name"`
	}{}
	for _, name := range packageNames {
		list = append(list, struct {
			Name string `json:"name"`
		}{name})
	}
	jsonSafe := "[]"
	data, err := json.Marshal(list)
	if err != nil {
		log.Printf("failed to marshal packages bundle, using fallback: %s", err)
	} else {
		jsonSafe = string(data)
	}

	return mockBodyf(`
{
    "apiVersion": "packages.eks.amazonaws.com/v1alpha1",
    "kind": "PackageBundle",
    "spec": {"packages": %s}
}
`, jsonSafe)
}

// mockBodyActiveCluster builds an http response body containing a minimal
// packages bundle controller with the provided name.
func mockBodyActiveCluster(clusterName string) io.ReadCloser {
	jsonSafe := "unknown"
	data, err := json.Marshal(clusterName)
	if err != nil {
		log.Printf("failed to marshal cluster name, using default: %s", err)
	} else {
		jsonSafe = string(data)
	}

	return mockBodyf(`
{
  "apiVersion": "packages.eks.amazonaws.com/v1alpha1",
  "kind": "PackageBundleController",
  "metadata":{
    "name": %s
  },
  "spec": {}
}
`, jsonSafe)
}

// mockBodyKubeadmControlPlane builds an http response body containing a
// minimal kubeadmcontrolplate response with the provided name.
func mockBodyKubeadmControlPlane(clusterName string) io.ReadCloser {
	jsonSafe := "unknown"
	data, err := json.Marshal(clusterName)
	if err != nil {
		log.Printf("failed to marshal cluster name, using default: %s", err)
	} else {
		jsonSafe = string(data)
	}

	return mockBodyf(`
{
    "apiVersion": "controlplane.cluster.x-k8s.io/v1beta1",
    "kind": "KubeadmControlPlane",
    "metadata": {
        "name": %s,
        "namespace": "eksa-system"
    },
    "spec": {},
    "status": {}
}
`, jsonSafe)
}
