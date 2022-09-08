package kubeconfig

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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

// FromEnvironment returns the first kubeconfig file specified in the
// KUBECONFIG environment variable.
//
// The environment variable can contain a list of files, much like how the
// PATH environment variable contains a list of directories.
func FromEnvironment() string {
	trimmed := strings.TrimSpace(os.Getenv(EnvName))
	for _, filename := range filepath.SplitList(trimmed) {
		return filename
	}
	return ""
}

// ValidateFile loads a file to validate it's basic contents.
//
// The values of the fields within aren't validated, but the file's existence
// and basic structure are checked.
func ValidateFile(filename string) error {
	wrapError := func(err error) error {
		return fmt.Errorf("validating kubeconfig %q: %w", filename, err)
	}

	// Trim whitespace from the beginning and end of the filename. While these
	// could technically be valid filenames, it's far more likely a typo or
	// shell-parsing bug.
	trimmed := strings.TrimSpace(filename)

	if !validations.FileExists(trimmed) {
		return wrapError(fs.ErrNotExist)
	}

	if !validations.FileExistsAndIsNotEmpty(trimmed) {
		return wrapError(fmt.Errorf("is empty"))
	}

	if _, err := clientcmd.LoadFromFile(trimmed); err != nil {
		return wrapError(err)
	}

	return nil
}

// ValidateFileOrEnv validates the given filename if it's not empty. If it's
// empty, it attempts to validate the filename returned from the
// FromEnvironment function. If that too is empty, it returns the empty
// string.
func ValidateFileOrEnv(filename string) (string, error) {
	var err error

	// Trim whitespace from the beginning and end of the filename. While these
	// could technically be valid filenames, it's far more likely a typo or
	// shell-parsing bug.
	if trimmed := strings.TrimSpace(filename); trimmed != "" {
		err = ValidateFile(trimmed)
		if err != nil {
			return "", err
		}
		return trimmed, nil
	}

	if envFilename := FromEnvironment(); envFilename != "" {
		err = ValidateFile(envFilename)
		if err != nil {
			return "", err
		}
		return envFilename, nil
	}

	return "", nil
}
