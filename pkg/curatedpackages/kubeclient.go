package curatedpackages

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

// NewKubeClientFromFilename creates a controller-runtime k8s client for use
// by CLI commands.
func NewKubeClientFromFilename(kubeConfigFilename string) (client.Client, error) {
	return newKubeClient(kubeConfigFilename, nil)
}

// newKubeClient creates a controller-runtime k8s client for use by CLI commands.
//
// If the RESTConfigurator is nil, a default configurator will be used.
//
// This isn't exported because it shouldn't be necessary in normal usage. It
// exists because we wanted to be able to inject a RESTConfigurator for tests,
// to bypass the network calls that the k8s client makes at
// initialization. There's no reason why the RESTConfigurator would need to be
// modified otherwise. If that condition changes, then so too should the
// constructor story here.
func newKubeClient(kubeConfigFilename string, rc restConfigurator) (client.Client, error) {
	data, err := os.ReadFile(kubeConfigFilename)
	if err != nil {
		return nil, err
	}
	if rc == nil {
		rc = restConfigurator(clientcmd.RESTConfigFromKubeConfig)
	}
	restConfig, err := rc.Config(data)
	if err != nil {
		return nil, err
	}
	scheme := runtime.NewScheme()
	if err := packagesv1.AddToScheme(scheme); err != nil {
		return nil, err
	}
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}
	return client.New(restConfig, client.Options{Scheme: scheme})
}

// restConfigurator abstracts the creation of a controller-runtime *rest.Config.
//
// This abstraction improves testing, as all known methods of instantiating a
// *rest.Config try to make network calls, and that's something we'd like to
// keep out of our unit tests as much as possible. In addition, where we do
// use them in unit tests, we need to be prepared with a controller-runtime
// EnvTest environment.
//
// For normal, non-test use, this can safely be ignored.
type restConfigurator func([]byte) (*rest.Config, error)

func (c restConfigurator) Config(data []byte) (*rest.Config, error) {
	return c(data)
}
