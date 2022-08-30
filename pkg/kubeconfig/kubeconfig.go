package kubeconfig

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// FromClusterFormat defines the format of the kubeconfig of the
const FromClusterFormat = "%s-eks-a-cluster.kubeconfig"

// EnvName is the standard KubeConfig environment variable name.
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#set-the-kubeconfig-environment-variable
const EnvName = "KUBECONFIG"

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters. This includes a subdirecftory
// named after the cluster name. For example, if the clusterName is 'sandbox' the generated path would be
// sandbox/sandbox-eks-a-cluster.kubeconfig
func FromClusterName(clusterName string) string {
	return filepath.Join(clusterName, fmt.Sprintf(FromClusterFormat, clusterName))
}

// FromEnvironment reads the KubeConfig path from the standard KUBECONFIG environment variable.
func FromEnvironment() string {
	return os.Getenv(EnvName)
}

// Loader abstracts the unmarshaling of a kubeconfig file into a Config.
//
// This is essentially building an interface around the clientcmd.Load
// function so that it can be swapped out for tests.
type Loader interface {
	// Load a Config from a byte slice.
	//
	// (Why do the k8s authors avoid io.Reader, forcing me to waste heap space
	// with extra byte slices?)
	Load([]byte) (*clientcmdapi.Config, error)
}

// clientcmdLoader loads a Config via the clientcmd.Load function.
type clientcmdLoader struct{}

var _ Loader = (*clientcmdLoader)(nil)

// Load implements Loader.
func (l *clientcmdLoader) Load(b []byte) (*clientcmdapi.Config, error) {
	return clientcmd.Load(b)
}

// Validator abstracts the validation of kubeconfig data.
type Validator interface {
	// Validate config data from an io.Reader.
	Validate(r io.Reader) error

	// ValidateFile validates config data from a filename.
	ValidateFile(filename string) error
}

type validator struct {
	loader Loader
}

var _ Validator = (*validator)(nil)

// NewValidatorWithLoader creates a Validator that loads config data with
// Loader.
//
// Most users will want to use NewValidator, which uses a default Loader.
func NewValidatorWithLoader(loader Loader) Validator {
	// Use the default if nil is passed, to avoid a nil pointer dereference.
	if loader == nil {
		loader = &clientcmdLoader{}
	}
	return &validator{
		loader: loader,
	}
}

// NewValidator creates a default Validator for kubeconfig data.
func NewValidator() Validator {
	return NewValidatorWithLoader(&clientcmdLoader{})
}

// Validate implements the Validator interface.
func (v *validator) Validate(r io.Reader) error {
	kubeConfigData, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// This is the only validation done at the moment. However, it
	// accomplishes a few things:
	//   1. Checks that the data exist (size > 0).
	//   2. Checks that the data fits the basic schema of a Config.
	//   3. If called via ValidateFile, checks that the file exists.
	if _, err := v.loader.Load(kubeConfigData); err != nil {
		return fmt.Errorf("validating kubeconfig: %w", err)
	}
	return nil
}

// ValidateFile implements the Validator interface.
func (v *validator) ValidateFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("%q: %w", filename, err)
	}
	defer f.Close()

	if err := v.Validate(f); err != nil {
		return fmt.Errorf("%q: %w", filename, err)
	}
	return nil
}
