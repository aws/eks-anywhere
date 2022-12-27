package kubernetes

import (
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NewRuntimeClientFromFileName creates a new controller runtime client given a kubeconfig filename.
func NewRuntimeClientFromFileName(kubeConfigFilename string) (client.Client, error) {
	data, err := os.ReadFile(kubeConfigFilename)
	if err != nil {
		return nil, fmt.Errorf("failed to create new client: %s", err)
	}

	return newRuntimeClient(data, nil, runtime.NewScheme())
}

func initScheme(scheme *runtime.Scheme) error {
	adders := append([]schemeAdder{
		clientgoscheme.AddToScheme,
	}, schemeAdders...)
	if scheme == nil {
		return fmt.Errorf("scheme was not provided")
	}
	return addToScheme(scheme, adders...)
}

func newRuntimeClient(data []byte, rc restConfigurator, scheme *runtime.Scheme) (client.Client, error) {
	if rc == nil {
		rc = restConfigurator(clientcmd.RESTConfigFromKubeConfig)
	}
	restConfig, err := rc.Config(data)
	if err != nil {
		return nil, err
	}

	if err := initScheme(scheme); err != nil {
		return nil, fmt.Errorf("failed to init client scheme %v", err)
	}

	err = clientgoscheme.AddToScheme(scheme)
	if err != nil {
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

// Config generates and returns a rest.Config from a kubeconfig in bytes.
func (c restConfigurator) Config(data []byte) (*rest.Config, error) {
	return c(data)
}
