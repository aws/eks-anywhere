package curatedpackages

import (
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

// NewKubeClientFromFilename creates a controller-runtime k8s client for use
// by CLI commands.
//
// If the RESTConfigurator is nil, a default configurator will be used.
func NewKubeClientFromFilename(kubeConfigFilename string, rc RESTConfigurator) (client.Client, error) {
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
	utilruntime.Must(packagesv1.AddToScheme(scheme))
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	return client.New(restConfig, client.Options{Scheme: scheme})
}

// RESTConfigurator abstracts the creation of a controller-runtime *rest.Config.
//
// This abstraction improves testing, as all known methods of instantiating a
// *rest.Config try to make network calls, and that's something we'd like to
// keep out of our unit tests as much as possible. In addition, where we do
// use them in unit tests, we need to be prepared with a controller-runtime
// EnvTest environment.
type RESTConfigurator interface {
	Config([]byte) (*rest.Config, error)
}

// restConfigurator implements a RESTConfigurator from a bare function.
//
// This can turn a regular function, for example
// clientcmd.RESTConfigFromKubeConfig, into a RESTConfigurator.
type restConfigurator func([]byte) (*rest.Config, error)

func (c restConfigurator) Config(data []byte) (*rest.Config, error) {
	return c(data)
}
