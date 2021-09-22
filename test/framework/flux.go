package framework

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
	"time"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/git"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/templater"
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

func WithFlux(opts ...api.GitOpsConfigOpt) E2ETestOpt {
	return func(e *E2ETest) {
		checkRequiredEnvVars(e.T, fluxRequiredEnvVars)
		e.GitOpsConfig = api.NewGitOpsConfig(defaultClusterName,
			api.WithPersonalFluxRepository(true),
			api.WithStringFromEnvVarGitOpsConfig(gitRepositoryVar, api.WithFluxRepository),
			api.WithStringFromEnvVarGitOpsConfig(githubUserVar, api.WithFluxOwner),
			api.WithStringFromEnvVarGitOpsConfig("main", api.WithFluxBranch),
			api.WithFluxNamespace("main"),
			api.WithFluxConfigurationPath("path2"),
			api.WithFluxBranch("default"),
		)
		e.clusterFillers = append(e.clusterFillers,
			api.WithGitOpsRef(defaultClusterName),
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

func withFluxRepositorySuffix(suffix string) api.GitOpsConfigOpt {
	return func(c *v1alpha1.GitOpsConfig) {
		repository := c.Spec.Flux.Github.Repository
		c.Spec.Flux.Github.Repository = fmt.Sprintf("%s-%s", repository, suffix)
	}
}

func (e *E2ETest) ValidateFlux() {
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

func (e *E2ETest) CleanUpGithubRepo() {
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
	err = gitOptions.Git.DeleteRepo(ctx, opts)
	if err != nil {
		e.T.Errorf("error while deleting Github repo %s: %v", repoName, err)
	}
}

type providerConfig struct {
	datacenterConfig providers.DatacenterConfig
	machineConfigs   []providers.MachineConfig
}

func (e *E2ETest) validateInitialFluxState(ctx context.Context) error {
	if err := e.validateFluxDeployments(ctx); err != nil {
		return err
	}

	if err := e.validateEksaSystemDeployments(ctx); err != nil {
		return err
	}
	return nil
}

func (e *E2ETest) validateWorkerNodeMultiConfigUpdates(ctx context.Context) error {
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
		vsphereMachineConfigs[workerName].Spec.DiskGiB = vsphereMachineConfigs[workerName].Spec.DiskGiB + 10
		vsphereMachineConfigs[workerName].Spec.MemoryMiB = 10196
		vsphereMachineConfigs[workerName].Spec.NumCPUs = 1

		// update replica
		clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
		if err != nil {
			return err
		}
		clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count += 1

		providerConfig := providerConfig{
			datacenterConfig: vsphereClusterConfig,
			machineConfigs:   e.convertVSphereMachineConfigs(cpName, workerName, vsphereMachineConfigs),
		}
		_, err = e.updateEKSASpecInGit(clusterConfig, providerConfig)
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

func (e *E2ETest) validateGitopsRepoContent(gitOptions *GitOptions) {
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

func (e *E2ETest) convertVSphereMachineConfigs(cpName, workerName string, vsphereMachineConfigs map[string]*v1alpha1.VSphereMachineConfig) []providers.MachineConfig {
	var configs []providers.MachineConfig
	configs = append(configs, vsphereMachineConfigs[cpName])
	if workerName != cpName {
		configs = append(configs, vsphereMachineConfigs[workerName])
	}
	return configs
}

func (e *E2ETest) validateWorkerNodeReplicaUpdates(ctx context.Context) error {
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

func (e *E2ETest) validateWorkerNodeUpdates(ctx context.Context) error {
	clusterConfGitPath := e.clusterConfigGitPath()
	clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
	if err != nil {
		return err
	}
	if err := e.waitForWorkerScaling(clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count); err != nil {
		return err
	}

	e.T.Log("Validating Worker Nodes replicas")
	if err := e.waitForWorkerNodeValidation(); err != nil {
		return err
	}

	e.T.Log("Validating Worker Node Machine Template")
	return e.validateWorkerNodeMachineSpec(ctx, clusterConfGitPath)
}

func (e *E2ETest) machineTemplateName(ctx context.Context) (string, error) {
	machineTemplateName, err := e.KubectlClient.MachineTemplateName(ctx, e.ClusterConfig.Name, e.cluster().KubeconfigFile, executables.WithNamespace(constants.EksaSystemNamespace))
	if err != nil {
		return "", err
	}
	return machineTemplateName, nil
}

func (e *E2ETest) validateFluxDeployments(ctx context.Context) error {
	deploymentReplicas := 1
	expectedDeployments := map[string]int{
		"helm-controller":         deploymentReplicas,
		"kustomize-controller":    deploymentReplicas,
		"notification-controller": deploymentReplicas,
		"source-controller":       deploymentReplicas,
	}
	return e.validateDeployments(ctx, fluxSystemNamespace, expectedDeployments)
}

func (e *E2ETest) validateEksaSystemDeployments(ctx context.Context) error {
	expectedDeployments := map[string]int{"eksa-controller-manager": 1}
	return e.validateDeployments(ctx, constants.EksaSystemNamespace, expectedDeployments)
}

func (e *E2ETest) validateDeployments(ctx context.Context, namespace string, expectedeployments map[string]int) error {
	err := retrier.Retry(20, time.Second, func() error {
		cluster := e.cluster()

		e.T.Logf("Getting deployments in %s namespace...", namespace)
		deployments, err := e.KubectlClient.GetDeployments(
			ctx,
			executables.WithCluster(cluster),
			executables.WithNamespace(namespace),
		)
		if err != nil {
			return fmt.Errorf("error getting deployments: %v", err)
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

func (e *E2ETest) updateWorkerNodeCountValue(newValue int) (string, error) {
	clusterConfGitPath := e.clusterConfigGitPath()
	providerConfig, err := e.providerConfig(clusterConfGitPath)
	if err != nil {
		return "", err
	}
	e.T.Logf("Updating workerNodeGroupConfiguration count to new value %d", newValue)

	clusterConfig, err := v1alpha1.GetClusterConfig(clusterConfGitPath)
	if err != nil {
		return "", err
	}
	clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count = newValue

	p, err := e.updateEKSASpecInGit(clusterConfig, *providerConfig)
	if err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated workerNodeGroupConfiguration count to new value %d", newValue)
	return p, nil
}

func (e *E2ETest) providerConfig(clusterConfGitPath string) (*providerConfig, error) {
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
		providerConfig.machineConfigs = e.convertVSphereMachineConfigs(
			e.ClusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name,
			e.ClusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name,
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

func (e *E2ETest) waitForWorkerNodeValidation() error {
	ctx := context.Background()
	return retrier.Retry(10, time.Second*10, func() error {
		e.T.Log("Attempting to validate worker nodes...")
		if err := e.KubectlClient.ValidateWorkerNodes(ctx, e.cluster()); err != nil {
			e.T.Logf("Worker node validation failed: %v", err)
			return fmt.Errorf("error while validating worker nodes: %v", err)
		}
		return nil
	})
}

func (e *E2ETest) validateWorkerNodeMachineSpec(ctx context.Context, clusterConfGitPath string) error {
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
			vsMachineTemplate, err := e.KubectlClient.VsphereWorkerNodesMachineTemplate(ctx, clusterConfig.Name, e.cluster().KubeconfigFile, constants.EksaSystemNamespace)
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
	default:
		return nil
	}
}

func (e *E2ETest) waitForWorkerScaling(targetvalue int) error {
	e.T.Logf("Waiting for worker node MachineDeployment to scale to target value %d", targetvalue)
	ctx := context.Background()
	cluster := e.cluster()
	return retrier.Retry(120, time.Second*10, func() error {
		md, err := e.KubectlClient.GetMachineDeployments(ctx,
			executables.WithCluster(cluster),
		)
		if err != nil {
			return err
		}

		for _, d := range md {
			r := int(d.Status.Replicas)
			if r != targetvalue {
				e.T.Logf("Waiting for worker node MachineDeployment %s replicas to scale; target: %d, actual: %d", d.Name, targetvalue, r)
				return fmt.Errorf(" MachineDeployment %s replicas are not at desired scale; target: %d, actual: %d", d.Name, targetvalue, r)
			}
			e.T.Logf("Worker node MachineDeployment %s Ready replicas have reached target scale %d", d.Name, r)
		}
		e.T.Logf("All worker node MachineDeployments have reached target scale %d", targetvalue)
		return nil
	})
}

func (e *E2ETest) updateEKSASpecInGit(c *v1alpha1.Cluster, providersConfig providerConfig) (string, error) {
	p, err := e.writeEKSASpec(c, providersConfig.datacenterConfig, providersConfig.machineConfigs)
	if err != nil {
		return "", err
	}
	if err := e.pushConfigChanges(); err != nil {
		return "", err
	}
	e.T.Logf("Successfully updated version controlled cluster configuration")
	return p, nil
}

func (e *E2ETest) pushConfigChanges() error {
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
func (e *E2ETest) writeEKSASpec(c *v1alpha1.Cluster, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) (path string, err error) {
	clusterObj, err := yaml.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("error outputting cluster yaml: %v", err)
	}

	datacenterObj, err := yaml.Marshal(datacenterConfig)
	if err != nil {
		return "", fmt.Errorf("error outputting datacenter yaml: %v", err)
	}
	resources := [][]byte{clusterObj, datacenterObj}
	for _, m := range machineConfigs {
		mObj, err := yaml.Marshal(m)
		if err != nil {
			return "", fmt.Errorf("error outputting machine yaml: %v", err)
		}
		resources = append(resources, mObj)
	}
	resourcesSpec := templater.AppendYamlResources(resources...)

	p := e.clusterConfGitPath()

	e.T.Logf("writing cluster config to path %s", p)
	clusterConfGitPath, err := e.GitWriter.Write(p, resourcesSpec, filewriter.PersistentFile)
	if err != nil {
		return "", err
	}
	return clusterConfGitPath, nil
}

func (e *E2ETest) gitRepoName() string {
	return e.GitOpsConfig.Spec.Flux.Github.Repository
}

func (e *E2ETest) gitBranch() string {
	return e.GitOpsConfig.Spec.Flux.Github.Branch
}

func (e *E2ETest) clusterConfGitPath() string {
	p := e.GitOpsConfig.Spec.Flux.Github.ClusterConfigPath
	if len(p) == 0 {
		p = path.Join("clusters", e.ClusterName)
	}
	return path.Join(p, constants.EksaSystemNamespace, eksaConfigFileName)
}

func (e *E2ETest) clusterConfigGitPath() string {
	return filepath.Join(e.GitWriter.Dir(), e.clusterConfGitPath())
}

func RequiredFluxEnvVars() []string {
	return fluxRequiredEnvVars
}
