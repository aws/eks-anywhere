package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"

	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/workflow"
	"github.com/aws/eks-anywhere/pkg/workflow/management"
	"github.com/aws/eks-anywhere/pkg/workflow/workflowcontext"
)

// KubernetesClient is a client for interacting with a Kubernetes cluster.
type KubernetesClient interface {
	Apply(ctx context.Context, cluster *types.Cluster, r io.Reader) error
}

// CustomComponentInstaller installs a component manifest at a configurable path to a bootstrap
// cluster. It's entrypoint is via a workflow hook.
type CustomComponentInstaller struct {
	fs           fs.FS
	manifestPath string
	k8s          KubernetesClient
}

// NewCustomComponentInstaller returns a new CustomComponentInstaller. ManifestPath is a path
// to a Kubernetes manifest that can be applied to a Kubernetes cluster.
func NewCustomComponentInstaller(filesystem fs.FS, manifestPath string) (*CustomComponentInstaller, error) {
	if !fs.ValidPath(manifestPath) {
		return nil, fmt.Errorf("invalid dir: %v", manifestPath)
	}

	return &CustomComponentInstaller{fs: filesystem, manifestPath: manifestPath}, nil
}

// RegisterCreateManagementClusterHooks satisfies management.CreateClusterHookRegistrar.
func (installer *CustomComponentInstaller) RegisterCreateManagementClusterHooks(binder workflow.HookBinder) {
	binder.BindPostTaskHook(
		management.CreateBootstrapCluster,
		workflow.TaskFunc(func(ctx context.Context) (context.Context, error) {
			bootstrap := workflowcontext.BootstrapCluster(ctx)
			if bootstrap == nil {
				return ctx, errors.New("bootstrap cluster not found in context")
			}

			return ctx, installer.install(ctx, bootstrap)
		}),
	)
}

// install opens installer's configured manifest path and applies it to the bootstrap cluster.
func (installer *CustomComponentInstaller) install(ctx context.Context, cluster *types.Cluster) error {
	fh, err := installer.fs.Open(installer.manifestPath)
	if err != nil {
		return err
	}

	return installer.k8s.Apply(ctx, cluster, fh)
}
