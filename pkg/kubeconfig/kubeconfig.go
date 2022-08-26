package kubeconfig

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
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

type missingFileError struct {
	Path string
}

// NewMissingFileError creates a missing kubeconfig file error.
func NewMissingFileError(path string) error {
	return missingFileError{Path: path}
}

func (m missingFileError) Error() string {
	return fmt.Sprintf("kubeconfig file missing: path=%v", m.Path)
}

// Validate checks that kubeconfig data is readable.
//
// This does not validate the values contained within; just that the data has
// the basic structure of a kubeconfig.
func Validate(r io.Reader) error {
	kubeConfigData, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// I wish I knew why the k8s authors avoid Reader/Writer interfaces so
	// much. :(
	if err := validator(kubeConfigData); err != nil {
		return err
	}

	return nil
}

// validator can be overridden for easier testing.
var validator = func(b []byte) error { _, err := clientcmd.Load(b); return err }

// ValidateFile is a helper for Validate.
func ValidateFile(filename string) error {
	f, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("%q: %w", filename, err)
	}
	defer f.Close()

	if err := Validate(f); err != nil {
		return fmt.Errorf("%q: %w", filename, err)
	}

	return nil
}
