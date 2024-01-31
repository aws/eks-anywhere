package flux

import (
	"context"
	"errors"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/validations"
)

// fluxForCluster bundles the Flux struct with a specific clusterSpec, so that all the git and file write
// operations for the clusterSpec can be done in each structure method.
type fluxForCluster struct {
	*Flux
	clusterSpec      *cluster.Spec
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
}

func newFluxForCluster(flux *Flux, clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) *fluxForCluster {
	return &fluxForCluster{
		Flux:             flux,
		clusterSpec:      clusterSpec,
		datacenterConfig: datacenterConfig,
		machineConfigs:   machineConfigs,
	}
}

// commitFluxAndClusterConfigToGit commits the cluster configuration file to the flux-managed git repository.
// If the remote repository does not exist it will initialize a local repository and push it to the configured remote.
// It will generate the kustomization file and marshal the cluster configuration file to the required locations in the repo.
// These will later be used by Flux and our controllers to reconcile the repository contents and the cluster configuration.
func (fc *fluxForCluster) commitFluxAndClusterConfigToGit(ctx context.Context, managementComponents *cluster.ManagementComponents) error {
	logger.Info("Adding cluster configuration files to Git")
	config := fc.clusterSpec.FluxConfig

	if err := fc.validateLocalConfigPathDoesNotExist(); err != nil {
		return err
	}

	g := NewFileGenerator()
	if err := g.Init(fc.writer, fc.eksaSystemDir(), fc.fluxSystemDir()); err != nil {
		return err
	}

	if err := g.WriteEksaFiles(fc.clusterSpec, fc.datacenterConfig, fc.machineConfigs); err != nil {
		return fmt.Errorf("writing eks-a config files: %v", err)
	}

	if fc.clusterSpec.Cluster.IsSelfManaged() {
		if err := g.WriteFluxSystemFiles(managementComponents, fc.clusterSpec); err != nil {
			return fmt.Errorf("writing flux system files: %v", err)
		}
	}

	p := path.Dir(config.Spec.ClusterConfigPath)
	if err := fc.gitClient.Add(p); err != nil {
		return fmt.Errorf("adding %s to git: %v", p, err)
	}

	if err := fc.Flux.pushToRemoteRepo(ctx, p, initialClusterconfigCommitMessage); err != nil {
		return err
	}

	logger.V(3).Info("Finished pushing cluster config and flux custom manifest files to git")
	return nil
}

func (fc *fluxForCluster) syncGitRepo(ctx context.Context) error {
	if !validations.FileExists(path.Join(fc.writer.Dir(), ".git")) {
		if err := fc.clone(ctx); err != nil {
			return fmt.Errorf("cloning git repo: %v", err)
		}
	} else {
		// Make sure the local git repo is on the branch specified in config and up-to-date with the remote
		if err := fc.gitClient.Branch(fc.branch()); err != nil {
			return fmt.Errorf("switching to git branch %s: %v", fc.branch(), err)
		}
	}
	return nil
}

func (fc *fluxForCluster) initializeProviderRepositoryIfNotExists(ctx context.Context) (*git.Repository, error) {
	// If git provider, the repository should be pre-initialized by the user.
	if fc.clusterSpec.FluxConfig.Spec.Git != nil {
		return &git.Repository{}, nil
	}

	r, err := fc.gitClient.GetRepo(ctx)
	if err != nil {
		return nil, fmt.Errorf("describing repo: %v", err)
	}

	if r != nil {
		return r, nil
	}

	if err = fc.createRemoteRepository(ctx); err != nil {
		return nil, err
	}

	if err = fc.initializeLocalRepository(); err != nil {
		return nil, err
	}

	return nil, nil
}

// setupRepository will set up the repository which will house the GitOps configuration for the cluster.
// if the repository exists and is not empty, it will be cloned.
// if the repository exists but is empty, it will be initialized locally, as a bare repository cannot be cloned.
// if the repository does not exist, it will be created and then initialized locally.
func (fc *fluxForCluster) setupRepository(ctx context.Context) (err error) {
	r, err := fc.initializeProviderRepositoryIfNotExists(ctx)
	if err != nil {
		return err
	}

	if r != nil {
		err = fc.clone(ctx)
	}

	var repoEmptyErr *git.RepositoryIsEmptyError
	if errors.As(err, &repoEmptyErr) {
		logger.V(3).Info("remote repository is empty and can't be cloned; will initialize locally")
		if initErr := fc.initializeLocalRepository(); initErr != nil {
			return fmt.Errorf("initializing local repository: %v", initErr)
		}
		return nil
	}

	return err
}

func (fc *fluxForCluster) clone(ctx context.Context) error {
	logger.V(3).Info("Cloning remote repository")
	if err := fc.gitClient.Clone(ctx); err != nil {
		return err
	}

	logger.V(3).Info("Creating a new branch")
	return fc.gitClient.Branch(fc.branch())
}

// createRemoteRepository will create a repository in the remote git provider with the user-provided configuration.
func (fc *fluxForCluster) createRemoteRepository(ctx context.Context) error {
	logger.V(3).Info("Remote Github repo does not exist; will create and initialize", "repo", fc.repository(), "owner", fc.owner())

	opts := git.CreateRepoOpts{
		Name:        fc.repository(),
		Owner:       fc.owner(),
		Description: "EKS-A cluster configuration repository",
		Personal:    fc.personal(),
		Privacy:     true,
	}

	logger.V(4).Info("Creating remote Github repo", "options", opts)
	if err := fc.gitClient.CreateRepo(ctx, opts); err != nil {
		return fmt.Errorf("creating repo: %v", err)
	}

	return nil
}

// initializeLocalRepository will git init the local repository directory, initialize a git repository.
// it will then change branches to the branch specified in the GitOps configuration.
func (fc *fluxForCluster) initializeLocalRepository() error {
	if err := fc.gitClient.Init(); err != nil {
		return fmt.Errorf("initializing repository: %v", err)
	}

	// git requires at least one commit in the repo to branch from
	if err := fc.gitClient.Commit("initializing repository"); err != nil {
		return fmt.Errorf("committing to repository: %v", err)
	}

	if err := fc.gitClient.Branch(fc.branch()); err != nil {
		return fmt.Errorf("creating branch: %v", err)
	}
	return nil
}

// validateLocalConfigPathDoesNotExist returns an exception if the cluster configuration file exists.
// This is done so that we avoid clobbering existing cluster configurations in the user-provided git repository.
func (fc *fluxForCluster) validateLocalConfigPathDoesNotExist() error {
	if fc.clusterSpec.Cluster.IsSelfManaged() {
		p := path.Join(fc.writer.Dir(), fc.path())
		if validations.FileExists(p) {
			return fmt.Errorf("a cluster configuration file already exists at path %s", p)
		}
	}
	return nil
}

func (fc *fluxForCluster) validateRemoteConfigPathDoesNotExist(ctx context.Context) error {
	if !fc.clusterSpec.Cluster.IsSelfManaged() || fc.gitClient == nil {
		return nil
	}

	exists, err := fc.gitClient.PathExists(ctx, fc.owner(), fc.repository(), fc.branch(), fc.path())
	if err != nil {
		return fmt.Errorf("failed validating remote flux config path: %v", err)
	}

	if exists {
		return fmt.Errorf("flux path %s already exists in remote repository", fc.path())
	}

	return nil
}

func (fc *fluxForCluster) namespace() string {
	return fc.clusterSpec.FluxConfig.Spec.SystemNamespace
}

func (fc *fluxForCluster) repository() string {
	if fc.clusterSpec.FluxConfig.Spec.Github != nil {
		return fc.clusterSpec.FluxConfig.Spec.Github.Repository
	}
	if fc.clusterSpec.FluxConfig.Spec.Git != nil {
		r := fc.clusterSpec.FluxConfig.Spec.Git.RepositoryUrl
		return path.Base(strings.TrimSuffix(r, filepath.Ext(r)))
	}
	return ""
}

func (fc *fluxForCluster) owner() string {
	if fc.clusterSpec.FluxConfig.Spec.Github != nil {
		return fc.clusterSpec.FluxConfig.Spec.Github.Owner
	}
	return ""
}

func (fc *fluxForCluster) branch() string {
	return fc.clusterSpec.FluxConfig.Spec.Branch
}

func (fc *fluxForCluster) personal() bool {
	if fc.clusterSpec.FluxConfig.Spec.Github != nil {
		return fc.clusterSpec.FluxConfig.Spec.Github.Personal
	}
	return false
}

func (fc *fluxForCluster) path() string {
	return fc.clusterSpec.FluxConfig.Spec.ClusterConfigPath
}

func (fc *fluxForCluster) eksaSystemDir() string {
	return path.Join(fc.path(), fc.clusterSpec.Cluster.GetName(), eksaSystemDirName)
}

func (fc *fluxForCluster) fluxSystemDir() string {
	return path.Join(fc.path(), fc.namespace())
}
