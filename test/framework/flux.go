package framework

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/version"
)

const (
	eksaConfigFileName  = "eksa-cluster.yaml"
	fluxSystemNamespace = "flux-system"
	gitRepositoryVar    = "T_GIT_REPOSITORY"
	githubUserVar       = "T_GITHUB_USER"
	githubTokenVar      = "EKSA_GITHUB_TOKEN"
)

var fluxRequiredEnvVars = []string{
	gitRepositoryVar,
	githubUserVar,
	githubTokenVar,
}

func WithFlux(opts ...api.GitOpsConfigOpt) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, fluxRequiredEnvVars)
		gitOpsConfigName := gitOpsConfigName()
		e.GitOpsConfig = api.NewGitOpsConfig(gitOpsConfigName,
			api.WithPersonalFluxRepository(true),
			api.WithStringFromEnvVarGitOpsConfig(gitRepositoryVar, api.WithFluxRepository),
			api.WithStringFromEnvVarGitOpsConfig(githubUserVar, api.WithFluxOwner),
			api.WithFluxNamespace("default"),
			api.WithFluxConfigurationPath("path2"),
			api.WithFluxBranch("main"),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithGitOpsRef(gitOpsConfigName),
		)
		// apply the rest of the opts passed into the function
		for _, opt := range opts {
			opt(e.GitOpsConfig)
		}
		// Adding Job ID suffix to repo name
		// e2e test jobs have Job Id with a ":", replacing with "-"
		jobId := strings.Replace(e.getJobIdFromEnv(), ":", "-", -1)
		withFluxRepositorySuffix(jobId)(e.GitOpsConfig)
		// Setting GitRepo cleanup since GitOps configured
		e.T.Cleanup(e.CleanUpGithubRepo)
	}
}

func WithClusterUpgradeGit(fillers ...api.ClusterFiller) ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		e.ClusterConfigB = e.customizeClusterConfig(e.clusterConfigGitPath(), fillers...)

		// TODO: e.GitopsConfig is defined from api.NewGitOpsConfig in WithFlux()
		// instead of marshalling from the actual file in git repo.
		// By default it does not include the namespace field. But Flux requires namespace always
		// exist for all the objects managed by its kustomization controller.
		// Need to refactor this to read gitopsconfig directly from file in git repo
		// which always has the namespace field.
		if e.GitOpsConfig.GetNamespace() == "" {
			e.GitOpsConfig.SetNamespace("default")
		}
	}
}

func withFluxRepositorySuffix(suffix string) api.GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		repository := c.Spec.Flux.Github.Repository
		c.Spec.Flux.Github.Repository = fmt.Sprintf("%s-%s", repository, suffix)
	}
}

func gitOpsConfigName() string {
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

	for _, opt := range clusterOpts {
		opt(e)
	}

	e.buildClusterConfigFileForGit()

	if err := e.pushConfigChanges(); err != nil {
		e.T.Errorf("Error pushing local changes to remote git repo: %v", err)
	}
	e.T.Logf("Successfully updated version controlled cluster configuration")

	if err := e.validateWorkerNodeUpdates(ctx); err != nil {
		e.T.Errorf("Error validating worker nodes after updating git repo: %v", err)
	}
}

func (e *ClusterE2ETest) initGit(ctx context.Context) {
	c := e.clusterConfig()
	writer, err := filewriter.NewWriter(e.cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}

	g, err := e.NewGitOptions(ctx, c, e.GitOpsConfig, writer, "")
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.GitProvider = g.Git
	e.GitWriter = g.Writer
}

func (e *ClusterE2ETest) buildClusterConfigFileForGit() {
	b := e.generateClusterConfig()
	_, err := e.GitWriter.Write(e.clusterConfGitPath(), b, filewriter.PersistentFile)
	if err != nil {
		e.T.Errorf("Error writing cluster config file to local git folder: %v", err)
	}
}

func (e *ClusterE2ETest) ValidateFlux() {
	c := e.clusterConfig()

	writer, err := filewriter.NewWriter(e.cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}
	ctx := context.Background()
	g, err := e.NewGitOptions(ctx, c, e.GitOpsConfig, writer, "")
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.GitProvider = g.Git
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
	gitOptions, err := e.NewGitOptions(ctx, c, e.GitOpsConfig, writer, fmt.Sprintf("%s/%s", e.ClusterName, repoName))
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	e.validateGitopsRepoContent(gitOptions)
}

func (e *ClusterE2ETest) CleanUpGithubRepo() {
	c := e.clusterConfig()
	writer, err := filewriter.NewWriter(e.cluster().Name)
	if err != nil {
		e.T.Errorf("Error configuring filewriter for e2e test: %v", err)
	}
	ctx := context.Background()
	owner := e.GitOpsConfig.Spec.Flux.Github.Owner
	repoName := e.gitRepoName()
	gitOptions, err := e.NewGitOptions(ctx, c, e.GitOpsConfig, writer, fmt.Sprintf("%s/%s", e.ClusterName, repoName))
	if err != nil {
		e.T.Errorf("Error configuring git client for e2e test: %v", err)
	}
	opts := git.DeleteRepoOpts{Owner: owner, Repository: repoName}
	repo, err := gitOptions.Git.GetRepo(ctx)
	if err != nil {
		e.T.Errorf("error getting Github repo %s: %v", repoName, err)
	}
	if repo == nil {
		e.T.Logf("Skipped repo deletion: remote repo %s does not exist", repoName)
		return
	}
	err = gitOptions.Git.DeleteRepo(ctx, opts)
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
	switch e.ClusterConfig.Spec.DatacenterRef.Kind {
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
		// update workernode specs
		vsphereMachineConfigs, err := v1alpha1.GetVSphereMachineConfigs(clusterConfGitPath)
		if err != nil {
			return err
		}
		cpName := e.ClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
		workerName := e.ClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
		etcdName := ""
		if e.ClusterConfig.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		vsphereMachineConfigs[workerName].Spec.DiskGiB = vsphereMachineConfigs[workerName].Spec.DiskGiB + 10
		vsphereMachineConfigs[workerName].Spec.MemoryMiB = 10196
		vsphereMachineConfigs[workerName].Spec.NumCPUs = 1

		// update replica
		clusterSpec, err := e.clusterSpecFromGit()
		if err != nil {
			return err
		}
		clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count += 1

		providerConfig := providerConfig{
			datacenterConfig: vsphereClusterConfig,
			machineConfigs:   e.convertVSphereMachineConfigs(cpName, workerName, etcdName, vsphereMachineConfigs),
		}
		_, err = e.updateEKSASpecInGit(clusterSpec, providerConfig)
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
	case v1alpha1.CloudStackDatacenterKind:
		clusterConfGitPath := e.clusterConfigGitPath()
		machineTemplateName, err := e.machineTemplateName(ctx)
		if err != nil {
			return err
		}
		cloudstackClusterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		// update workernode specs
		cloudstackMachineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(clusterConfGitPath)
		if err != nil {
			return err
		}
		cpName := e.ClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name
		workerName := e.ClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name
		etcdName := ""
		if e.ClusterConfig.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}

		// update replica
		clusterSpec, err := e.clusterSpecFromGit()
		if err != nil {
			return err
		}
		cloudstackMachineConfigs[workerName].Spec.UserCustomDetails = map[string]string{
			"foo": "bar",
		}

		providerConfig := providerConfig{
			datacenterConfig: cloudstackClusterConfig,
			machineConfigs:   e.convertCloudStackMachineConfigs(cpName, workerName, etcdName, cloudstackMachineConfigs),
		}
		_, err = e.updateEKSASpecInGit(clusterSpec, providerConfig)
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

func (e *ClusterE2ETest) validateGitopsRepoContent(gitOptions *GitOptions) {
	repoName := e.gitRepoName()
	gitFilePath := e.clusterConfigGitPath()
	localFilePath := filepath.Join(e.ClusterName, repoName, e.clusterConfGitPath())
	ctx := context.Background()
	g := gitOptions.Git
	err := g.Clone(ctx)
	if err != nil {
		e.T.Errorf("Error cloning github repo: %v", err)
	}
	branch := e.gitBranch()
	err = g.Branch(branch)
	if err != nil {
		e.T.Errorf("Error checking out branch: %v", err)
	}
	gitFile, err := ioutil.ReadFile(gitFilePath)
	if err != nil {
		e.T.Errorf("Error opening file from the original repo directory: %v", err)
	}
	localFile, err := ioutil.ReadFile(localFilePath)
	if err != nil {
		e.T.Errorf("Error opening file from the newly created repo directory: %v", err)
	}
	if !bytes.Equal(gitFile, localFile) {
		e.T.Errorf("Error validating the content of github repo: %v", err)
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

func (e *ClusterE2ETest) convertCloudStackMachineConfigs(cpName, workerName, etcdName string, cloudstackMachineConfigs map[string]*v1alpha1.CloudStackMachineConfig) []providers.MachineConfig {
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
	_, err = e.updateWorkerNodeCountValue(3)
	if err != nil {
		return err
	}

	if err := e.validateWorkerNodeUpdates(ctx); err != nil {
		return err
	}

	_, err = e.updateWorkerNodeCountValue(1)
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
	if err := e.waitForWorkerScaling(clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Name, clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count); err != nil {
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
	machineTemplateName, err := e.KubectlClient.MachineTemplateName(ctx, e.ClusterConfig.Name, e.cluster().KubeconfigFile, executables.WithNamespace(constants.EksaSystemNamespace))
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

func (e *ClusterE2ETest) updateWorkerNodeCountValue(newValue int) (string, error) {
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
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].Count = newValue

	p, err := e.updateEKSASpecInGit(clusterSpec, *providerConfig)
	if err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated workerNodeGroupConfiguration count to new value %d", newValue)
	return p, nil
}

func (e *ClusterE2ETest) providerConfig(clusterConfGitPath string) (*providerConfig, error) {
	var providerConfig providerConfig
	switch e.ClusterConfig.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		machineConfigs, err := v1alpha1.GetVSphereMachineConfigs(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		providerConfig.datacenterConfig = datacenterConfig
		etcdName := ""
		if e.ClusterConfig.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		providerConfig.machineConfigs = e.convertVSphereMachineConfigs(
			e.ClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
			e.ClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name,
			etcdName,
			machineConfigs)
	case v1alpha1.CloudStackDatacenterKind:
		datacenterConfig, err := v1alpha1.GetCloudStackDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		machineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		providerConfig.datacenterConfig = datacenterConfig
		etcdName := ""
		if e.ClusterConfig.Spec.ExternalEtcdConfiguration != nil {
			etcdName = e.ClusterConfig.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name
		}
		providerConfig.machineConfigs = e.convertCloudStackMachineConfigs(
			e.ClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
			e.ClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name,
			etcdName,
			machineConfigs)
	case v1alpha1.DockerDatacenterKind:
		datacenterConfig, err := v1alpha1.GetDockerDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return nil, err
		}
		providerConfig.datacenterConfig = datacenterConfig
	default:
		return nil, fmt.Errorf("unexpected DatacenterRef %s", e.ClusterConfig.Spec.DatacenterRef.Kind)
	}
	return &providerConfig, nil
}

func (e *ClusterE2ETest) waitForWorkerNodeValidation() error {
	ctx := context.Background()
	return retrier.Retry(120, time.Second*10, func() error {
		e.T.Log("Attempting to validate worker nodes...")
		if err := e.KubectlClient.ValidateWorkerNodes(ctx, e.ClusterConfig.Name, e.managementKubeconfigFilePath()); err != nil {
			e.T.Logf("Worker node validation failed: %v", err)
			return fmt.Errorf("validating worker nodes: %v", err)
		}
		return nil
	})
}

func (e *ClusterE2ETest) validateWorkerNodeMachineSpec(ctx context.Context, clusterConfGitPath string) error {
	switch e.ClusterConfig.Spec.DatacenterRef.Kind {
	case v1alpha1.VSphereDatacenterKind:
		clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		vsphereClusterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		vsphereMachineConfigs, err := v1alpha1.GetVSphereMachineConfigs(clusterConfGitPath)
		if err != nil {
			return err
		}
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
		cloudstackMachineConfigs, err := v1alpha1.GetCloudStackMachineConfigs(clusterConfGitPath)
		if err != nil {
			return err
		}
		cloudstackWorkerConfig := cloudstackMachineConfigs[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
		return retrier.Retry(120, time.Second*10, func() error {
			csMachineTemplate, err := e.KubectlClient.CloudstackWorkerNodesMachineTemplate(ctx, clusterConfig.Name, e.cluster().KubeconfigFile, constants.EksaSystemNamespace)
			if err != nil {
				return err
			}
			if cloudstackWorkerConfig.Spec.Template.Name != csMachineTemplate.Spec.Spec.Spec.Template.Name {
				err := fmt.Errorf("MachineSpec %s Template are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.Template, csMachineTemplate.Spec.Spec.Spec.Template)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if cloudstackWorkerConfig.Spec.ComputeOffering.Name != csMachineTemplate.Spec.Spec.Spec.Offering.Name {
				err := fmt.Errorf("MachineSpec %s Offering are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.ComputeOffering, csMachineTemplate.Spec.Spec.Spec.Offering)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if !reflect.DeepEqual(cloudstackWorkerConfig.Spec.UserCustomDetails, csMachineTemplate.Spec.Spec.Spec.Details) {
				err := fmt.Errorf("MachineSpec %s Details are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.UserCustomDetails, csMachineTemplate.Spec.Spec.Spec.Details)
				e.T.Logf("Waiting for WorkerNode Specs to match - %s", err.Error())
				return err
			}
			if !reflect.DeepEqual(cloudstackWorkerConfig.Spec.AffinityGroupIds, csMachineTemplate.Spec.Spec.Spec.AffinityGroupIDs) {
				err := fmt.Errorf("MachineSpec %s AffinityGroupIds are not at desired value; target: %v, actual: %v", csMachineTemplate.Name, cloudstackWorkerConfig.Spec.AffinityGroupIds, csMachineTemplate.Spec.Spec.Spec.AffinityGroupIDs)
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

func (e *ClusterE2ETest) updateEKSASpecInGit(s *cluster.Spec, providersConfig providerConfig) (string, error) {
	p, err := e.writeEKSASpec(s, providersConfig.datacenterConfig, providersConfig.machineConfigs)
	if err != nil {
		return "", err
	}
	if err := e.pushConfigChanges(); err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated version controlled cluster configuration")
	return p, nil
}

func (e *ClusterE2ETest) pushConfigChanges() error {
	p := e.clusterConfGitPath()
	g := e.GitProvider
	if err := g.Add(p); err != nil {
		return err
	}
	if err := g.Commit("EKS-A E2E Flux test configuration update"); err != nil {
		return err
	}
	return g.Push(context.Background())
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
	return e.GitOpsConfig.Spec.Flux.Github.Repository
}

func (e *ClusterE2ETest) gitBranch() string {
	return e.GitOpsConfig.Spec.Flux.Github.Branch
}

func (e *ClusterE2ETest) clusterConfGitPath() string {
	p := e.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath
	if len(p) == 0 {
		p = path.Join("clusters", e.ClusterName)
	}
	return path.Join(p, e.ClusterName, constants.EksaSystemNamespace, eksaConfigFileName)
}

func (e *ClusterE2ETest) clusterConfigGitPath() string {
	return filepath.Join(e.GitWriter.Dir(), e.clusterConfGitPath())
}

func (e *ClusterE2ETest) clusterSpecFromGit() (*cluster.Spec, error) {
	s, err := cluster.NewSpecFromClusterConfig(
		e.clusterConfigGitPath(),
		version.Get(),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to build spec from git: %v", err)
	}
	return s, nil
}

func RequiredFluxEnvVars() []string {
	return fluxRequiredEnvVars
}
