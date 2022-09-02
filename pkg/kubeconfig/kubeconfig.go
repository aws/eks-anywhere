package kubeconfig

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/aws/eks-anywhere/pkg/validations"
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

// ValidateFile loads a file to validate it's basic contents.
//
// The values of the fields within aren't validated, but the file's existence
// and basic structure are checked.
func ValidateFile(filename string) error {
	wrapError := func(err error) error {
		return fmt.Errorf("validating kubeconfig %q: %w", filename, err)
	}

	if !validations.FileExists(filename) {
		return wrapError(fs.ErrNotExist)
	}

	if !validations.FileExistsAndIsNotEmpty(filename) {
		return wrapError(fmt.Errorf("is empty"))
	}

	if _, err := clientcmd.LoadFromFile(filename); err != nil {
		return wrapError(err)
	}

	return nil
}
