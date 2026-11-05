package framework

import (
	"context"
	"strings"
	"time"

	"github.com/aws/eks-anywhere/pkg/types"
)

// The packages controller replaces its own pod when the active package bundle pins a different controller
// than the one EKS-A installed, and again when a test changes the active bundle. The Package webhook is
// unreachable until the new pod serves, and the eks-anywhere-packages Package already reports installed
// at that point, so waiting on it isn't enough. Package changes are retried for that window instead.
var (
	packagesWebhookRetryTimeout  = 3 * time.Minute
	packagesWebhookRetryInterval = 10 * time.Second
)

// isPackagesWebhookUnavailable reports whether output shows that the API server couldn't reach the
// packages controller webhook (connection refused, no endpoints, timeout). A request the webhook rejects
// ("admission webhook ... denied the request") doesn't match.
func isPackagesWebhookUnavailable(output string) bool {
	return strings.Contains(output, "failed calling webhook") && strings.Contains(output, "eks-anywhere-packages")
}

// retryWhilePackagesWebhookUnavailable runs op until it succeeds, fails for another reason, or
// packagesWebhookRetryTimeout passes. op returns the command output to check along with its error.
func (e *ClusterE2ETest) retryWhilePackagesWebhookUnavailable(op func() (string, error)) error {
	deadline := time.Now().Add(packagesWebhookRetryTimeout)
	for {
		output, err := op()
		if err == nil || !isPackagesWebhookUnavailable(output) || time.Now().After(deadline) {
			return err
		}
		e.T.Logf("Packages controller webhook is unavailable, retrying in %s: %v", packagesWebhookRetryInterval, err)
		time.Sleep(packagesWebhookRetryInterval)
	}
}

// applyPackageSpec applies a Package manifest to namespace, retrying while the packages controller
// webhook is unavailable.
func (e *ClusterE2ETest) applyPackageSpec(ctx context.Context, cluster *types.Cluster, spec []byte, namespace string) error {
	return e.retryWhilePackagesWebhookUnavailable(func() (string, error) {
		err := e.KubectlClient.ApplyKubeSpecFromBytesWithNamespace(ctx, cluster, spec, namespace)
		if err != nil {
			return err.Error(), err
		}
		return "", nil
	})
}

// runEKSAPackageCommand runs an eksctl anywhere command that creates or changes packages, retrying while
// the packages controller webhook is unavailable. Other failures fail the test like RunEKSA.
func (e *ClusterE2ETest) runEKSAPackageCommand(args []string) {
	var output string
	err := e.retryWhilePackagesWebhookUnavailable(func() (string, error) {
		var err error
		output, err = e.runCommand(e.eksaBinaryLocation, args...)
		return output, err
	})
	if err != nil {
		e.failCommand(e.eksaBinaryLocation, args, output, err)
	}
}
