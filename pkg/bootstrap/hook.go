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

// CustomComponentInstaller installs a set of component manifests to a bootstrap cluster. It
// is intended for use with management cluster workflows where it binds as a post bootstrap
// cluster creation hook.
type CustomComponentInstaller struct {
	fs  fs.FS
	dir string
	k8s KubernetesClient
}

// NewCustomComponentInstaller returns a new CustomComponentInstaller. ManifestPath is a path
// to a Kubernetes manifest that can be applied to a Kubernetes cluster.
func NewCustomComponentInstaller(filesystem fs.FS, dir string) (*CustomComponentInstaller, error) {
	if !fs.ValidPath(dir) {
		return nil, fmt.Errorf("invalid dir: %v", dir)
	}

	return &CustomComponentInstaller{fs: filesystem, dir: dir}, nil
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

// install reads all files found at installers configured dir and.
func (installer *CustomComponentInstaller) install(ctx context.Context, cluster *types.Cluster) error {
	entries, err := fs.ReadDir(installer.fs, ".")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		fh, err := installer.fs.Open(entry.Name())
		if err != nil {
			return err
		}

		if err := installer.k8s.Apply(ctx, cluster, fh); err != nil {
			return err
		}

		if err := fh.Close(); err != nil {
			// Log error
		}
	}

	return nil
}
