package kubeconfig

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"k8s.io/client-go/tools/clientcmd"

	"github.com/aws/eks-anywhere/pkg/validations"
)

// Writer reads the kubeconfig secret on a cluster and copies the contents to a writer.
type Writer interface {
	WriteKubeconfig(ctx context.Context, clusterName, kubeconfig string, w io.Writer) error
}

// FromClusterFormat defines the format of the kubeconfig of the.
const FromClusterFormat = "%s-eks-a-cluster.kubeconfig"

// EnvName is the standard KubeConfig environment variable name.
// https://kubernetes.io/docs/tasks/access-application-cluster/configure-access-multiple-clusters/#set-the-kubeconfig-environment-variable
const EnvName = "KUBECONFIG"

// FromClusterName formats an expected Kubeconfig path for EKS-A clusters. This includes a subdirecftory
// named after the cluster name. For example, if the clusterName is 'sandbox' the generated path would be
// sandbox/sandbox-eks-a-cluster.kubeconfig.
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

// ResolveFilename returns a path to a kubeconfig file by priority.
//
// The priority is:
//
//  1. CLI flag (flagValue)
//  2. A file created at cluster creation, found by a combining the cluster
//     name with present working directory.
//  3. The first filename found in the KUBECONFIG environment variable.
//
// NO VALIDATION IS PERFORMED. See ValidateFile for validation.
//
// There are other places one may wish to consult or load a kubeconfig file
// from, but this function tries to walk the narrow line between what the
// kubernetes client code does (#1, #3, and some other things that we more or
// less don't support), with some of the existing EKSA CLI tools that look for
// kubeconfig files relative to the working directory that were created at
// cluster creation time. These different functionalities don't always mesh,
// and aren't always compatible, but this function tries to combine them as
// much as possible, without breaking either.
func ResolveFilename(flagValue, clusterName string) string {
	if flagValue != "" {
		return flagValue
	}

	if clusterName != "" {
		return FromClusterName(clusterName)
	}

	return FromEnvironment()
}

// ResolveAndValidate composes ResolveFilename and ValidateFile.
//
// Literally, that's all it does. They're frequently called together, so
// hopefully this is a helper.
func ResolveAndValidateFilename(flagValue, clusterName string) (string, error) {
	filename := ResolveFilename(flagValue, clusterName)
	if err := ValidateFilename(filename); err != nil {
		return "", err
	}
	return filename, nil
}

// ValidateFilename loads a file to validate it's basic contents.
//
// The values of the fields within aren't validated, but the file's existence
// and basic structure are checked.
func ValidateFilename(filename string) error {
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

func ValidateKubeconfigPath(clusterName string, parentFolders ...string) error {
	kubeconfigPath := FromClusterName(clusterName)
	for _, folder := range parentFolders {
		kubeconfigPath = filepath.Join(folder, kubeconfigPath)
	}
	info, err := os.Stat(kubeconfigPath)
	if err == nil && info.Size() > 0 {
		return fmt.Errorf(
			"old cluster config file exists under %s, please use a different clusterName to proceed",
			clusterName,
		)
	}
	return nil
}
