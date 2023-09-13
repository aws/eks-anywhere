package framework

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	gitfactory "github.com/aws/eks-anywhere/pkg/git/factory"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/version"
)

const (
	eksaConfigFileName    = "eksa-cluster.yaml"
	fluxSystemNamespace   = "flux-system"
	GitRepositoryVar      = "T_GIT_REPOSITORY"
	GitRepoSshUrl         = "T_GIT_SSH_REPO_URL"
	GithubUserVar         = "T_GITHUB_USER"
	GithubTokenVar        = "EKSA_GITHUB_TOKEN"
	GitKnownHosts         = "EKSA_GIT_KNOWN_HOSTS"
	GitPrivateKeyFile     = "EKSA_GIT_PRIVATE_KEY"
	DefaultFluxConfigName = "eksa-test"
)

var fluxGithubRequiredEnvVars = []string{
	GitRepositoryVar,
	GithubUserVar,
	GithubTokenVar,
}

var fluxGitRequiredEnvVars = []string{
	GitKnownHosts,
	GitPrivateKeyFile,
	GitRepoSshUrl,
}

var fluxGitCreateGenerateRepoEnvVars = []string{
	GitKnownHosts,
	GitPrivateKeyFile,
	GithubUserVar,
	GithubTokenVar,
}

func getJobIDFromEnv() string {
	return os.Getenv(JobIdVar)
}

func WithFluxGit(opts ...api.FluxConfigOpt) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, fluxGitRequiredEnvVars)
		jobID := strings.Replace(getJobIDFromEnv(), ":", "-", -1)
		e.ClusterConfig.FluxConfig = api.NewFluxConfig(DefaultFluxConfigName,
			api.WithGenericGitProvider(
				api.WithStringFromEnvVarGenericGitProviderConfig(GitRepoSshUrl, api.WithGitRepositoryUrl),
			),
			api.WithSystemNamespace("default"),
			api.WithClusterConfigPath(jobID),
			api.WithBranch(jobID),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithGitOpsRef(DefaultFluxConfigName, v1alpha1.FluxConfigKind),
		)
		// apply the rest of the opts passed into the function
		for _, opt := range opts {
			opt(e.ClusterConfig.FluxConfig)
		}
		e.T.Cleanup(e.CleanUpGitRepo)
	}
}

func WithFluxGithub(opts ...api.FluxConfigOpt) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		fluxConfigName := fluxConfigName()
		checkRequiredEnvVars(e.T, fluxGithubRequiredEnvVars)
		e.ClusterConfig.FluxConfig = api.NewFluxConfig(fluxConfigName,
			api.WithGithubProvider(
				api.WithPersonalGithubRepository(true),
				api.WithStringFromEnvVarGithubProviderConfig(GitRepositoryVar, api.WithGithubRepository),
				api.WithStringFromEnvVarGithubProviderConfig(GithubUserVar, api.WithGithubOwner),
			),
			api.WithSystemNamespace("default"),
			api.WithClusterConfigPath("path2"),
			api.WithBranch("main"),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithGitOpsRef(fluxConfigName, v1alpha1.FluxConfigKind),
		)
		// apply the rest of the opts passed into the function
		for _, opt := range opts {
			opt(e.ClusterConfig.FluxConfig)
		}
		// Adding Job ID suffix to repo name
		// e2e test jobs have Job Id with a ":", replacing with "-"
		jobID := strings.Replace(getJobIDFromEnv(), ":", "-", -1)
		withFluxRepositorySuffix(jobID)(e.ClusterConfig.FluxConfig)

		// Setting GitRepo cleanup since GitOps configured
		e.T.Cleanup(e.CleanUpGithubRepo)
	}
}

// WithFluxGithubConfig returns ClusterConfigFiller that adds FluxConfig using the Github provider to the cluster config.
func WithFluxGithubConfig(opts ...api.FluxConfigOpt) api.ClusterConfigFiller {
	fluxConfigName := fluxConfigName()
	return api.JoinClusterConfigFillers(func(config *cluster.Config) {
		config.FluxConfig = api.NewFluxConfig(fluxConfigName,
			api.WithGithubProvider(
				api.WithPersonalGithubRepository(true),
				api.WithStringFromEnvVarGithubProviderConfig(GitRepositoryVar, api.WithGithubRepository),
				api.WithStringFromEnvVarGithubProviderConfig(GithubUserVar, api.WithGithubOwner),
			),
			api.WithSystemNamespace("default"),
			api.WithBranch("main"),
		)
		// apply the rest of the opts passed into the function
		for _, opt := range opts {
			opt(config.FluxConfig)
		}
		// Adding Job ID suffix to repo name
		// e2e test jobs have Job Id with a ":", replacing with "-"
		jobID := strings.Replace(getJobIDFromEnv(), ":", "-", -1)
		withFluxRepositorySuffix(jobID)(config.FluxConfig)
	}, api.ClusterToConfigFiller(api.WithGitOpsRef(fluxConfigName, v1alpha1.FluxConfigKind)))
}

// WithFluxGithubEnvVarCheck returns a ClusterE2ETestOpt that checks for the required env vars.
func WithFluxGithubEnvVarCheck() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, fluxGithubRequiredEnvVars)
	}
}

// WithFluxGithubCleanup returns a ClusterE2ETestOpt that registers the git repository cleanup operation.
func WithFluxGithubCleanup() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.T.Cleanup(e.CleanUpGithubRepo)
	}
}

func WithClusterUpgradeGit(fillers ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.UpdateClusterConfig(
			api.ClusterToConfigFiller(fillers...),
			func(c *cluster.Config) {
				// TODO: e.ClusterConfig.GitOpsConfig is defined from api.NewGitOpsConfig in WithFluxLegacy()
				// instead of marshalling from the actual file in git repo.
				// By default it does not include the namespace field. But Flux requires namespace always
				// exist for all the objects managed by its kustomization controller.
				// Need to refactor this to read gitopsconfig directly from file in git repo
				// which always has the namespace field.
				if c.GitOpsConfig != nil {
					if c.GitOpsConfig.GetNamespace() == "" {
						c.GitOpsConfig.SetNamespace("default")
					}
					c.FluxConfig = c.GitOpsConfig.ConvertToFluxConfig()
				}

				if c.FluxConfig.GetNamespace() == "" {
					c.FluxConfig.SetNamespace("default")
				}
			},
		)
	}
}

func withFluxRepositorySuffix(suffix string) api.FluxConfigOpt {
	return func(c *v1alpha1.FluxConfig) {
		repository := c.Spec.Github.Repository
		c.Spec.Github.Repository = fmt.Sprintf("%s-%s", repository, suffix)
	}
}

func fluxConfigName() string {
	return fmt.Sprintf("%s-%s", defaultClusterName, test.RandString(5))
}

func (e *ClusterE2ETest) UpgradeWithGitOps(clusterOpts ...ClusterE2ETestOpt) {
	e.upgradeWithGitOps(clusterOpts)
}

func (e *ClusterE2ETest) upgradeWithGitOps(clusterOpts []ClusterE2ETestOpt) {
	ctx := context.Background()
	e.initGit(ctx)

	if err := e.validateInitialFluxState(ctx); err != nil {
		e.T.Errorf("Error validating initial state of cluster gitops system: %v", err)
	}

	err := e.pullRemoteConfig(ctx)
	if err != nil {
		e.T.Errorf("pulling remote configuration: %v", err)
	}

	e.T.Log("Parsing pulled config from repo into test ClusterConfig")
	// Read the cluster config we just pulled into e.ClusterConfig
	e.parseClusterConfigFromLocalGitRepo()

	// Apply the options, these are most of the times fillers, so they will update the
	// cluster config we just read from the repo. This has to happen after we parse the cluster
	// config from the repo or we might be updating a different version of the config.
	for _, opt := range clusterOpts {
		opt(e)
	}

	e.T.Log("Updating local cluster config file in git repo for upgrade")
	// Marshall e.ClusterConfig and write it to the repo path
	e.buildClusterConfigFileForGit()

	if err := e.pushConfigChanges(ctx); err != nil {
		e.T.Errorf("Error pushing local changes to remote git repo: %v", err)
	}
	e.T.Logf("Successfully updated version controlled cluster configuration")

	if err := e.validateWorkerNodeUpdates(ctx); err != nil {
		e.T.Errorf("Error validating worker nodes after updating git repo: %v", err)
	}
}

func (e *ClusterE2ETest) initGit(ctx context.Context) {
	c := e.ClusterConfig.Cluster
	writer, err := filewriter.NewWriter(e.Cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}

	if e.ClusterConfig.GitOpsConfig != nil {
		e.ClusterConfig.FluxConfig = e.ClusterConfig.GitOpsConfig.ConvertToFluxConfig()
	}

	g, err := e.NewGitTools(ctx, c, e.ClusterConfig.FluxConfig, writer, "")
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.GitProvider = g.Provider
	e.GitWriter = g.Writer
	e.GitClient = g.Client
}

func (e *ClusterE2ETest) workloadClusterConfigPath(w *WorkloadCluster) string {
	return e.clusterConfigPathFromName(w.ClusterName)
}

func (e *ClusterE2ETest) workloadClusterConfigGitPath(w *WorkloadCluster) string {
	return filepath.Join(e.GitWriter.Dir(), e.workloadClusterConfigPath(w))
}

func (e *ClusterE2ETest) buildWorkloadClusterConfigFileForGit(w *WorkloadCluster) {
	b := w.generateClusterConfigYaml()
	g := e.GitWriter
	p := filepath.Dir(e.workloadClusterConfigGitPath(w))

	if _, err := os.Stat(p); errors.Is(err, os.ErrNotExist) {
		err := os.MkdirAll(p, os.ModePerm)
		if err != nil {
			w.T.Fatalf("Creating directory [%s]: %v", g.Dir(), err)
		}
	}

	_, err := g.Write(e.workloadClusterConfigPath(w), b, filewriter.PersistentFile)
	if err != nil {
		w.T.Fatalf("Error writing cluster config file to local git folder: %v", err)
	}
}

func (e *ClusterE2ETest) addWorkloadClusterConfigToGit(ctx context.Context, w *WorkloadCluster) error {
	p := e.workloadClusterConfigPath(w)
	g := e.GitClient
	if err := g.Add(p); err != nil {
		return fmt.Errorf("adding cluster path changes at %s: %v", p, err)
	}

	if err := e.pushStagedChanges(ctx, "EKS-A E2E Flux test workload configuration changes added"); err != nil {
		return fmt.Errorf("failed to push workload configuration changes %v", err)
	}
	return nil
}

func (e *ClusterE2ETest) deleteWorkloadClusterConfigFromGit(ctx context.Context, w *WorkloadCluster) error {
	p := filepath.Dir(e.workloadClusterConfigPath(w))
	g := e.GitClient
	if err := g.Remove(p); err != nil {
		return fmt.Errorf("removing cluster config at path %s: %v", p, err)
	}

	if err := e.pushStagedChanges(ctx, "EKS-A E2E Flux test workload configuration deleted"); err != nil {
		return err
	}
	return nil
}

func (e *ClusterE2ETest) pushStagedChanges(ctx context.Context, commitMessage string) error {
	g := e.GitClient
	if err := g.Commit(commitMessage); err != nil {
		return fmt.Errorf("commiting staged changes: %v", err)
	}

	repoUpToDateErr := &git.RepositoryUpToDateError{}
	if err := g.Push(ctx); err != nil {
		if !errors.Is(err, repoUpToDateErr) {
			return fmt.Errorf("pushing staged changes to remote: %v", err)
		}
		e.T.Log(err.Error())
	}
	return nil
}

func (e *ClusterE2ETest) pushWorkloadClusterToGit(w *WorkloadCluster, opts ...api.ClusterConfigFiller) error {
	ctx := context.Background()
	e.initGit(ctx)
	// Pull remote config using managment cluster
	err := e.pullRemoteConfig(ctx)
	if err != nil {
		e.T.Errorf("Pulling remote configuration: %v", err)
	}

	if _, err := os.Stat(e.workloadClusterConfigGitPath(w)); err == nil {
		// Read the cluster config we just pulled into w.ClusterConfig
		e.T.Log("Parsing pulled config from repo into test ClusterConfig")
		w.parseClusterConfigFromDisk(e.workloadClusterConfigGitPath(w))
	}

	// Update the cluster config with the provided api.ClusterConfigFillers
	w.UpdateClusterConfig(opts...)
	e.T.Log("Updating local config file in git repo")
	// Marshall w.ClusterConfig and write it to the repo path
	e.buildWorkloadClusterConfigFileForGit(w)
	if err := e.addWorkloadClusterConfigToGit(ctx, w); err != nil {
		return fmt.Errorf("failed to push local changes to remote git repo: %v", err)
	}

	e.T.Logf("Successfully pushed version controlled cluster configuration")

	return nil
}

func (e *ClusterE2ETest) deleteWorkloadClusterFromGit(w *WorkloadCluster) error {
	ctx := context.Background()
	e.initGit(ctx)

	err := e.pullRemoteConfig(ctx)
	if err != nil {
		e.T.Errorf("Pulling remote configuration: %v", err)
	}

	if err := e.deleteWorkloadClusterConfigFromGit(ctx, w); err != nil {
		return fmt.Errorf("failed to push local changes to remote git repo: %v", err)
	}
	w.T.Logf("Successfully deleted version controlled cluster")

	return nil
}

func (e *ClusterE2ETest) parseClusterConfigFromLocalGitRepo() {
	c, err := cluster.ParseConfigFromFile(e.clusterConfigGitPath())
	if err != nil {
		e.T.Fatalf("Failed parsing cluster config from git repo: %s", err)
	}

	e.ClusterConfig = c
}

func (e *ClusterE2ETest) buildClusterConfigFileForGit() {
	b := e.generateClusterConfigYaml()
	_, err := e.GitWriter.Write(e.clusterConfGitPath(), b, filewriter.PersistentFile)
	if err != nil {
		e.T.Errorf("Error writing cluster config file to local git folder: %v", err)
	}
}

func (e *ClusterE2ETest) ValidateFlux() {
	c := e.ClusterConfig.Cluster

	writer, err := filewriter.NewWriter(e.Cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}
	ctx := context.Background()
	if e.ClusterConfig.GitOpsConfig != nil {
		e.ClusterConfig.FluxConfig = e.ClusterConfig.GitOpsConfig.ConvertToFluxConfig()
	}
	g, err := e.NewGitTools(ctx, c, e.ClusterConfig.FluxConfig, writer, "")
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.GitClient = g.Client
	e.GitProvider = g.Provider
	e.GitWriter = g.Writer

	if err = e.validateInitialFluxState(ctx); err != nil {
		e.T.Errorf("Error validating initial state of cluster gitops system: %v", err)
	}

	if err = e.validateWorkerNodeReplicaUpdates(ctx); err != nil {
		e.T.Errorf("Error validating scaling of Flux managed cluster: %v", err)
	}

	if err = e.validateWorkerNodeMultiConfigUpdates(ctx); err != nil {
		e.T.Errorf("Error upgrading worker nodes: %v", err)
	}
	writer, err = filewriter.NewWriter("")
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}
	repoName := e.gitRepoName()
	gitTools, err := e.NewGitTools(ctx, c, e.ClusterConfig.FluxConfig, writer, e.validateGitopsRepoContentPath(repoName))
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.validateGitopsRepoContent(gitTools)
}

func (e *ClusterE2ETest) CleanUpGitRepo() {
	c := e.ClusterConfig.Cluster
	writer, err := filewriter.NewWriter(e.Cluster().Name)
	if err != nil {
		e.T.Errorf("configuring filewriter for e2e test: %v", err)
	}
	ctx := context.Background()
	repoName := e.gitRepoName()
	gitTools, err := e.NewGitTools(ctx, c, e.ClusterConfig.FluxConfig, writer, fmt.Sprintf("%s/%s", e.ClusterName, repoName))
	if err != nil {
		e.T.Errorf("configuring git client for e2e test: %v", err)
	}
	dirEntries, err := os.ReadDir(gitTools.RepositoryDirectory)
	if errors.Is(err, os.ErrNotExist) {
		e.T.Logf("repository directory %s does not exist; skipping cleanup", gitTools.RepositoryDirectory)
		return
	}

	for _, entry := range dirEntries {
		if entry.Name() == ".git" {
			continue
		}
		if entry.IsDir() {
			err = os.RemoveAll(entry.Name())
			e.T.Logf("cleaning up directory: %v", entry.Name())
			if err != nil {
				e.T.Log("did not remove directory", "dir", entry.Name(), "err", err)
				continue
			}
		}
		if !entry.IsDir() {
			err = os.Remove(entry.Name())
			e.T.Logf("cleaning up file: %v", entry.Name())
			if err != nil {
				e.T.Log("did not remove file", "file", entry.Name(), "err", err)
				continue
			}
		}
	}

	if err = gitTools.Client.Add("*"); err != nil {
		e.T.Logf("did not add files while cleaning up git repo: %v", err)
	}
	if err = gitTools.Client.Push(context.Background()); err != nil {
		e.T.Logf("did not push to repo after cleanup: %v", err)
	}
}

func (e *ClusterE2ETest) CleanUpGithubRepo() {
	c := e.ClusterConfig.Cluster
	writer, err := filewriter.NewWriter(e.Cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}
	ctx := context.Background()

	if e.ClusterConfig.GitOpsConfig != nil {
		e.ClusterConfig.FluxConfig = e.ClusterConfig.GitOpsConfig.ConvertToFluxConfig()
	}
	owner := e.ClusterConfig.FluxConfig.Spec.Github.Owner
	repoName := e.gitRepoName()
	gitTools, err := e.NewGitTools(ctx, c, e.ClusterConfig.FluxConfig, writer, fmt.Sprintf("%s/%s", e.ClusterName, repoName))
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	opts := git.DeleteRepoOpts{Owner: owner, Repository: repoName}
	repo, err := gitTools.Provider.GetRepo(ctx)
	if err != nil {
		e.T.Errorf("error getting Github repo %s: %v", repoName, err)
	}
	if repo == nil {
		e.T.Logf("Skipped repo deletion: remote repo %s does not exist", repoName)
		return
	}
	err = gitTools.Provider.DeleteRepo(ctx, opts)
	if err != nil {
		e.T.Errorf("error while deleting Github repo %s: %v", repoName, err)
	}
}

type providerConfig struct {
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
}

func (e *ClusterE2ETest) validateInitialFluxState(ctx context.Context) error {
	if err := e.validateFluxDeployments(ctx); err != nil {
		return err
	}

	if err := e.validateEksaSystemDeployments(ctx); err != nil {
		return err
	}
	return nil
}

func (e *ClusterE2ETest) validateWorkerNodeMultiConfigUpdates(ctx context.Context) error {
	switch e.ClusterConfig.Cluster.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		clusterConfGitPath := e.clusterConfigGitPath()
		machineTemplateName, err := e.machineTemplateName(ctx)
		if err != nil {
			return err
		}
		vsphereClusterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		config, err := cluster.ParseConfigFromFile(clusterConfGitPath)
		if err != nil {
			return err
		}
		vsphereMachineConfigs := config.VSphereMachineConfigs

		// update workernode specs
		cpName := e.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
		workerName := e.ClusterConfig.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
		etcdName := ""
		if e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		vsphereMachineConfigs[workerName].Spec.DiskGiB = vsphereMachineConfigs[workerName].Spec.DiskGiB + 10
		vsphereMachineConfigs[workerName].Spec.MemoryMiB = 10196
		vsphereMachineConfigs[workerName].Spec.NumCPUs = 1

		// update replica
		clusterSpec, err := e.clusterSpecFromGit()
		if err != nil {
			return err
		}
		count := *clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count + 1
		clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = &count

		providerConfig := providerConfig{
			datacenterConfig: vsphereClusterConfig,
			machineConfigs:   e.convertVSphereMachineConfigs(cpName, workerName, etcdName, vsphereMachineConfigs),
		}
		_, err = e.updateEKSASpecInGit(ctx, clusterSpec, providerConfig)
		if err != nil {
			return err
		}
		err = e.validateWorkerNodeUpdates(ctx)
		if err != nil {
			return err
		}
		newMachineTemplateName, err := e.machineTemplateName(ctx)
		if err != nil {
			return err
		}
		if machineTemplateName == newMachineTemplateName {
			return fmt.Errorf("machine template name should change on machine resource updates, old %s and new %s", machineTemplateName, newMachineTemplateName)
		}
		return nil
	default:
		return nil
	}
}

func (e *ClusterE2ETest) validateGitopsRepoContentPath(repoName string) string {
	return filepath.Join(e.ClusterName, "e2e-validate", repoName)
}

func (e *ClusterE2ETest) validateGitopsRepoContent(gitTools *gitfactory.GitTools) {
	repoName := e.gitRepoName()
	gitFilePath := e.clusterConfigGitPath()
	localFilePath := filepath.Join(e.validateGitopsRepoContentPath(repoName), e.clusterConfGitPath())
	ctx := context.Background()
	gc := gitTools.Client
	err := gc.Clone(ctx)
	if err != nil {
		e.T.Errorf("Error cloning github repo: %v", err)
	}
	branch := e.gitBranch()
	err = gc.Branch(branch)
	if err != nil {
		e.T.Errorf("Error checking out branch: %v", err)
	}
	gitFile, err := os.ReadFile(gitFilePath)
	if err != nil {
		e.T.Errorf("Error opening file from the original repo directory: %v", err)
	}
	localFile, err := os.ReadFile(localFilePath)
	if err != nil {
		e.T.Errorf("Error opening file from the newly created repo directory: %v", err)
	}
	if !bytes.Equal(gitFile, localFile) {
		e.T.Errorf("Error validating the content of git repo: %v", err)
	}
}

func (e *ClusterE2ETest) convertVSphereMachineConfigs(cpName, workerName, etcdName string, vsphereMachineConfigs map[string]*v1alpha1.VSphereMachineConfig) []providers.MachineConfig {
	var configs []providers.MachineConfig
	if vsphereMachineConfigs[cpName] != nil {
		configs = append(configs, vsphereMachineConfigs[cpName])
	}
	if workerName != cpName && vsphereMachineConfigs[workerName] != nil {
		configs = append(configs, vsphereMachineConfigs[workerName])
	}
	if etcdName != "" && etcdName != cpName && etcdName != workerName && vsphereMachineConfigs[etcdName] != nil {
		configs = append(configs, vsphereMachineConfigs[etcdName])
	}
	return configs
}

func (e *ClusterE2ETest) convertCloudstackMachineConfigs(cpName, workerName, etcdName string, cloudstackMachineConfigs map[string]*v1alpha1.CloudStackMachineConfig) []providers.MachineConfig {
	var configs []providers.MachineConfig
	if cloudstackMachineConfigs[cpName] != nil {
		configs = append(configs, cloudstackMachineConfigs[cpName])
	}
	if workerName != cpName && cloudstackMachineConfigs[workerName] != nil {
		configs = append(configs, cloudstackMachineConfigs[workerName])
	}
	if etcdName != "" && etcdName != cpName && etcdName != workerName && cloudstackMachineConfigs[etcdName] != nil {
		configs = append(configs, cloudstackMachineConfigs[etcdName])
	}
	return configs
}

func (e *ClusterE2ETest) validateWorkerNodeReplicaUpdates(ctx context.Context) error {
	machineTemplateName, err := e.machineTemplateName(ctx)
	if err != nil {
		return err
	}
	_, err = e.updateWorkerNodeCountValue(ctx, 3)
	if err != nil {
		return err
	}

	if err := e.validateWorkerNodeUpdates(ctx); err != nil {
		return err
	}

	_, err = e.updateWorkerNodeCountValue(ctx, 1)
	if err != nil {
		return err
	}
	newMachineTemplateName, err := e.machineTemplateName(ctx)
	if err != nil {
		return err
	}
	if machineTemplateName != newMachineTemplateName {
		return fmt.Errorf("machine template name shouldn't change on just replica updates, old %s and new %s", machineTemplateName, newMachineTemplateName)
	}
	return e.validateWorkerNodeUpdates(ctx)
}

func (e *ClusterE2ETest) validateWorkerNodeUpdates(ctx context.Context, opts ...CommandOpt) error {
	clusterConfGitPath := e.clusterConfigGitPath()
	clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
	if err != nil {
		return err
	}
	if err := e.waitForWorkerScaling(clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Name, *clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count); err != nil {
		return err
	}

	e.T.Log("Validating Worker Nodes replicas")
	if err := e.waitForWorkerNodeValidation(); err != nil {
		return err
	}

	e.T.Log("Validating Worker Node Machine Template")
	return e.validateWorkerNodeMachineSpec(ctx, clusterConfGitPath)
}

func (e *ClusterE2ETest) machineTemplateName(ctx context.Context) (string, error) {
	machineTemplateName, err := e.KubectlClient.MachineTemplateName(ctx, e.ClusterConfig.Cluster.Name, e.Cluster().KubeconfigFile, executables.WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return "", err
	}
	return machineTemplateName, nil
}

func (e *ClusterE2ETest) validateFluxDeployments(ctx context.Context) error {
	deploymentReplicas := 1
	expectedDeployments := map[string]int{
		"helm-controller":         deploymentReplicas,
		"kustomize-controller":    deploymentReplicas,
		"notification-controller": deploymentReplicas,
		"source-controller":       deploymentReplicas,
	}
	return e.validateDeploymentsInManagementCluster(ctx, fluxSystemNamespace, expectedDeployments)
}

func (e *ClusterE2ETest) validateEksaSystemDeployments(ctx context.Context) error {
	expectedDeployments := map[string]int{"eksa-controller-manager": 1}
	return e.validateDeploymentsInManagementCluster(ctx, constants.EksaSystemNamespace, expectedDeployments)
}

func (e *ClusterE2ETest) validateDeploymentsInManagementCluster(ctx context.Context, namespace string, expectedeployments map[string]int) error {
	err := retrier.Retry(20, time.Second, func() error {
		e.T.Logf("Getting deployments in %s namespace...", namespace)
		deployments, err := e.KubectlClient.GetDeployments(
			ctx,
			executables.WithKubeconfig(e.managementKubeconfigFilePath()),
			executables.WithNamespace(namespace),
		)
		if err != nil {
			return fmt.Errorf("getting deployments: %v", err)
		}

		for _, deployment := range deployments {
			_, ok := expectedeployments[deployment.Name]
			if !ok {
				e.T.Errorf("Error validating %s deployments; unepxected deployment %s present in namespace", namespace, deployment.Name)
			}
			if expectedeployments[deployment.Name] != int(deployment.Status.ReadyReplicas) {
				e.T.Log("Deployments have not scaled yet")
				return fmt.Errorf("expected %d ready replicas of deployment %s; got %d ready replicas", expectedeployments[deployment.Name], deployment.Name, deployment.Status.ReadyReplicas)
			}
		}
		return nil
	})
	if err != nil {
		e.T.Errorf("Error validating %s deployments: %v", namespace, err)
		return err
	}
	e.T.Logf("Successfully validated %s deployments are present and ready", namespace)
	return nil
}

func (e *ClusterE2ETest) updateWorkerNodeCountValue(ctx context.Context, newValue int) (string, error) {
	clusterConfGitPath := e.clusterConfigGitPath()
	providerConfig, err := e.providerConfig(clusterConfGitPath)
	if err != nil {
		return "", err
	}
	e.T.Logf("Updating workerNodeGroupConfiguration count to new value %d", newValue)

	clusterSpec, err := e.clusterSpecFromGit()
	if err != nil {
		return "", err
	}
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = &newValue

	p, err := e.updateEKSASpecInGit(ctx, clusterSpec, *providerConfig)
	if err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated workerNodeGroupConfiguration count to new value %d", newValue)
	return p, nil
}

func (e *ClusterE2ETest) providerConfig(clusterConfGitPath string) (*providerConfig, error) {
	config, err := cluster.ParseConfigFromFile(clusterConfGitPath)
	if err != nil {
		return nil, err
	}

	var providerConfig providerConfig
	switch e.ClusterConfig.Cluster.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		machineConfigs := config.VSphereMachineConfigs
		providerConfig.datacenterConfig = datacenterConfig
		etcdName := ""
		if e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		providerConfig.machineConfigs = e.convertVSphereMachineConfigs(
			e.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
			e.ClusterConfig.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name,
			etcdName,
			machineConfigs)
	case v1alpha1.DockerDatacenterKind:
		datacenterConfig, err := v1alpha1.GetDockerDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		providerConfig.datacenterConfig = datacenterConfig
	case v1alpha1.CloudStackDatacenterKind:
		datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		providerConfig.datacenterConfig = datacenterConfig
		machineConfigs := config.CloudStackMachineConfigs
		etcdName := ""
		if e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		providerConfig.machineConfigs = e.convertCloudstackMachineConfigs(
			e.ClusterConfig.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
			e.ClusterConfig.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name,
			etcdName,
			machineConfigs)
	default:
		return nil, fmt.Errorf("unexpected DatacenterRef %s", e.ClusterConfig.Cluster.Spec.DatacenterRef.Kind)
	}
	return &providerConfig, nil
}

func (e *ClusterE2ETest) waitForWorkerNodeValidation() error {
	ctx := context.Background()
	return retrier.Retry(120, time.Second*10, func() error {
		e.T.Log("Attempting to validate worker nodes...")
		if err := e.KubectlClient.ValidateWorkerNodes(ctx, e.ClusterConfig.Cluster.Name, e.managementKubeconfigFilePath()); err != nil {
			e.T.Logf("Worker node validation failed: %v", err)
			return fmt.Errorf("validating worker nodes: %v", err)
		}
		return nil
	})
}

func (e *ClusterE2ETest) validateWorkerNodeMachineSpec(ctx context.Context, clusterConfGitPath string) error {
	config, err := cluster.ParseConfigFromFile(clusterConfGitPath)
	if err != nil {
		return err
	}

	switch e.ClusterConfig.Cluster.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		vsphereClusterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		vsphereMachineConfigs := config.VSphereMachineConfigs
		vsphereWorkerConfig := vsphereMachineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
		return retrier.Retry(120, time.Second*10, func() error {
			vsMachineTemplate, err := e.KubectlClient.VsphereWorkerNodesMachineTemplate(ctx, clusterConfig.Name, e.managementKubeconfigFilePath(), constants.EksaSystemNamespace)
			if err != nil {
				return err
			}
			if vsphereWorkerConfig.Spec.NumCPUs != int(vsMachineTemplate.Spec.Template.Spec.NumCPUs) {
				err := fmt.Errorf("MachineSpec %s WorkloadVMsNumCPUs are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.NumCPUs, vsMachineTemplate.Spec.Template.Spec.NumCPUs)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereWorkerConfig.Spec.DiskGiB != int(vsMachineTemplate.Spec.Template.Spec.DiskGiB) {
				err := fmt.Errorf("MachineSpec %s WorkloadDiskGiB are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.DiskGiB, vsMachineTemplate.Spec.Template.Spec.DiskGiB)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereWorkerConfig.Spec.Template != vsMachineTemplate.Spec.Template.Spec.Template {
				err := fmt.Errorf("MachineSpec %s Template are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.Template, vsMachineTemplate.Spec.Template.Spec.Template)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereWorkerConfig.Spec.Folder != vsMachineTemplate.Spec.Template.Spec.Folder {
				err := fmt.Errorf("MachineSpec %s Folder are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.Folder, vsMachineTemplate.Spec.Template.Spec.Folder)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if len(vsMachineTemplate.Spec.Template.Spec.Network.Devices) == 0 {
				err := fmt.Errorf("MachineSpec %s Template has no devices", vsMachineTemplate.Name)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereClusterConfig.Spec.Network != vsMachineTemplate.Spec.Template.Spec.Network.Devices[0].NetworkName {
				err := fmt.Errorf("MachineSpec %s Template are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereClusterConfig.Spec.Network, vsMachineTemplate.Spec.Template.Spec.Network.Devices[0].NetworkName)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereWorkerConfig.Spec.Datastore != vsMachineTemplate.Spec.Template.Spec.Datastore {
				err := fmt.Errorf("MachineSpec %s Datastore are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.Datastore, vsMachineTemplate.Spec.Template.Spec.Datastore)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereClusterConfig.Spec.Datacenter != vsMachineTemplate.Spec.Template.Spec.Datacenter {
				err := fmt.Errorf("MachineSpec %s Datacenter are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereClusterConfig.Spec.Datacenter, vsMachineTemplate.Spec.Template.Spec.Datacenter)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereWorkerConfig.Spec.ResourcePool != vsMachineTemplate.Spec.Template.Spec.ResourcePool {
				err := fmt.Errorf("MachineSpec %s ResourcePool are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereWorkerConfig.Spec.ResourcePool, vsMachineTemplate.Spec.Template.Spec.ResourcePool)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereClusterConfig.Spec.Server != vsMachineTemplate.Spec.Template.Spec.Server {
				err := fmt.Errorf("MachineSpec %s Server are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereClusterConfig.Spec.Server, vsMachineTemplate.Spec.Template.Spec.Server)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if vsphereClusterConfig.Spec.Thumbprint != vsMachineTemplate.Spec.Template.Spec.Thumbprint {
				err := fmt.Errorf("MachineSpec %s Template are not at desired value; target: %v, actual: %v", vsMachineTemplate.Name, vsphereClusterConfig.Spec.Thumbprint, vsMachineTemplate.Spec.Template.Spec.Thumbprint)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			e.T.Logf("Worker MachineTemplate values have matched expected values")
			return nil
		})
	case v1alpha1.CloudStackDatacenterKind:
		clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		cloudstackMachineConfigs := config.CloudStackMachineConfigs
		cloudstackWorkerConfig := cloudstackMachineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
		return retrier.Retry(120, time.Second*10, func() error {
			csMachineTemplate, err := e.KubectlClient.CloudstackWorkerNodesMachineTemplate(ctx, clusterConfig.Name, e.managementKubeconfigFilePath(), constants.EksaSystemNamespace)
			if err != nil {
				return err
			}
			if cloudstackWorkerConfig.Spec.Template.Name != csMachineTemplate.Spec.Template.Spec.Template.Name {
				err := fmt.Errorf("MachineSpec %s Template are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.Template, csMachineTemplate.Spec.Template.Spec.Template)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if cloudstackWorkerConfig.Spec.ComputeOffering.Name != csMachineTemplate.Spec.Template.Spec.Offering.Name {
				err := fmt.Errorf("MachineSpec %s Offering are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.ComputeOffering, csMachineTemplate.Spec.Template.Spec.Offering)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if !reflect.DeepEqual(cloudstackWorkerConfig.Spec.UserCustomDetails, csMachineTemplate.Spec.Template.Spec.Details) {
				err := fmt.Errorf("MachineSpec %s Details are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.UserCustomDetails, csMachineTemplate.Spec.Template.Spec.Details)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			var symlinks []string
			for key, value := range cloudstackWorkerConfig.Spec.Symlinks {
				symlinks = append(symlinks, key+":"+value)
			}
			if strings.Join(symlinks, ",") != csMachineTemplate.Annotations["symlinks."+constants.CloudstackAnnotationSuffix] {
				err := fmt.Errorf("MachineSpec %s Symlinks are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.Symlinks, csMachineTemplate.Annotations["symlinks."+constants.CloudstackAnnotationSuffix])
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if !reflect.DeepEqual(cloudstackWorkerConfig.Spec.AffinityGroupIds, csMachineTemplate.Spec.Template.Spec.AffinityGroupIDs) {
				err := fmt.Errorf("MachineSpec %s AffinityGroupIds are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.AffinityGroupIds, csMachineTemplate.Spec.Template.Spec.AffinityGroupIDs)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			e.T.Logf("Worker MachineTemplate values have matched expected values")
			return nil
		})
	default:
		return nil
	}
}

func (e *ClusterE2ETest) waitForWorkerScaling(name string, targetvalue int) error {
	e.T.Logf("Waiting for worker node group %v MachineDeployment to scale to target value %d", name, targetvalue)
	ctx := context.Background()
	return retrier.Retry(120, time.Second*10, func() error {
		md, err := e.KubectlClient.GetMachineDeployment(ctx, fmt.Sprintf("%v-%v", e.ClusterName, name),
			executables.WithKubeconfig(e.managementKubeconfigFilePath()),
			executables.WithNamespace(constants.EksaSystemNamespace),
		)
		if err != nil {
			e.T.Logf("Unable to get machine deployment: %v", err)
			return err
		}

		r := int(md.Status.Replicas)
		if r != targetvalue {
			e.T.Logf("Waiting for worker node MachineDeployment %s replicas to scale; target: %d, actual: %d", md.Name, targetvalue, r)
			return fmt.Errorf(" MachineDeployment %s replicas are not at desired scale; target: %d, actual: %d", md.Name, targetvalue, r)
		}
		e.T.Logf("Worker node MachineDeployment %s Ready replicas have reached target scale %d", md.Name, r)
		e.T.Logf("All worker node MachineDeployments have reached target scale %d", targetvalue)
		return nil
	})
}

func (e *ClusterE2ETest) updateEKSASpecInGit(ctx context.Context, s *cluster.Spec, providersConfig providerConfig) (string, error) {
	err := e.pullRemoteConfig(ctx)
	if err != nil {
		return "", err
	}

	p, err := e.writeEKSASpec(s, providersConfig.datacenterConfig, providersConfig.machineConfigs)
	if err != nil {
		return "", err
	}
	if err := e.pushConfigChanges(ctx); err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated version controlled cluster configuration")
	return p, nil
}

func (e *ClusterE2ETest) pushConfigChanges(ctx context.Context) error {
	p := e.clusterConfGitPath()
	g := e.GitClient
	if err := g.Add(p); err != nil {
		return fmt.Errorf("adding cluster config changes at path %s: %v", p, err)
	}
	if err := e.pushStagedChanges(ctx, "EKS-A E2E Flux test configuration update"); err != nil {
		return fmt.Errorf("failed to push config changes %v", err)
	}
	return nil
}

func (e *ClusterE2ETest) pullRemoteConfig(ctx context.Context) error {
	g := e.GitClient
	repoUpToDateErr := &git.RepositoryUpToDateError{}
	if err := g.Pull(ctx, e.gitBranch()); err != nil {
		if !errors.Is(err, repoUpToDateErr) {
			return fmt.Errorf("pulling from remote before pushing config changes: %v", err)
		}
		e.T.Log(err.Error())
	}
	return nil
}

// todo: reuse logic in clustermanager to template resources
func (e *ClusterE2ETest) writeEKSASpec(s *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) (path string, err error) {
	resourcesSpec, err := clustermarshaller.MarshalClusterSpec(s, datacenterConfig, machineConfigs)
	if err != nil {
		return "", err
	}
	p := e.clusterConfGitPath()

	e.T.Logf("writing cluster config to path %s", p)
	clusterConfGitPath, err := e.GitWriter.Write(p, resourcesSpec, filewriter.PersistentFile)
	if err != nil {
		return "", err
	}
	return clusterConfGitPath, nil
}

func (e *ClusterE2ETest) gitRepoName() string {
	if e.ClusterConfig.FluxConfig.Spec.Github != nil {
		return e.ClusterConfig.FluxConfig.Spec.Github.Repository
	}
	if e.ClusterConfig.FluxConfig.Spec.Git != nil {
		r := e.ClusterConfig.FluxConfig.Spec.Git.RepositoryUrl
		return strings.TrimSuffix(path.Base(r), filepath.Ext(r))
	}
	return ""
}

func (e *ClusterE2ETest) gitBranch() string {
	return e.ClusterConfig.FluxConfig.Spec.Branch
}

func (e *ClusterE2ETest) clusterConfigPathFromName(clusterName string) string {
	p := e.ClusterConfig.FluxConfig.Spec.ClusterConfigPath
	if len(p) == 0 {
		p = path.Join("clusters", e.ClusterName)
	}
	return path.Join(p, clusterName, constants.EksaSystemNamespace, eksaConfigFileName)
}

func (e *ClusterE2ETest) clusterConfGitPath() string {
	return e.clusterConfigPathFromName(e.ClusterName)
}

func (e *ClusterE2ETest) clusterConfigGitPath() string {
	return filepath.Join(e.GitWriter.Dir(), e.clusterConfGitPath())
}

func (e *ClusterE2ETest) clusterSpecFromGit() (*cluster.Spec, error) {
	var opts []cluster.FileSpecBuilderOpt
	if getBundlesOverride() == "true" {
		// This makes sure that the cluster.Spec uses the same Bundles we pass to the CLI
		// It avoids the budlesRef getting overwritten with whatever default Bundles the
		// e2e test build is configured to use
		opts = append(opts, cluster.WithOverrideBundlesManifest(defaultBundleReleaseManifestFile))
	}
	b := cluster.NewFileSpecBuilder(files.NewReader(), version.Get(), opts...)
	s, err := b.Build(e.clusterConfigGitPath())
	if err != nil {
		return nil, fmt.Errorf("unable to build spec from git: %v", err)
	}
	return s, nil
}

func RequiredFluxGithubEnvVars() []string {
	return fluxGithubRequiredEnvVars
}

func RequiredFluxGitCreateRepoEnvVars() []string {
	return fluxGitCreateGenerateRepoEnvVars
}
