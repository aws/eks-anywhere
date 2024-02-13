package executables

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
	anywherecluster "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	clusterCtlPath                = "clusterctl"
	clusterctlConfigFile          = "clusterctl_tmp.yaml"
	capiPrefix                    = "/generated/overrides"
	etcdadmBootstrapProviderName  = "etcdadm-bootstrap"
	etcdadmControllerProviderName = "etcdadm-controller"
	kubeadmBootstrapProviderName  = "kubeadm"
)

//go:embed config/clusterctl.yaml
var clusterctlConfigTemplate string

type Clusterctl struct {
	Executable
	writer filewriter.FileWriter
	reader manifests.FileReader
}

type clusterctlConfiguration struct {
	coreVersion              string
	bootstrapVersion         string
	controlPlaneVersion      string
	configFile               string
	etcdadmBootstrapVersion  string
	etcdadmControllerVersion string
}

// NewClusterctl builds a new [Clusterctl].
func NewClusterctl(executable Executable, writer filewriter.FileWriter, reader manifests.FileReader) *Clusterctl {
	return &Clusterctl{
		Executable: executable,
		writer:     writer,
		reader:     reader,
	}
}

func imageRepository(image v1alpha1.Image) string {
	return path.Dir(image.Image())
}

// This method will write the configuration files
// used by cluster api to install components.
// See: https://cluster-api.sigs.k8s.io/clusterctl/configuration.html
func (c *Clusterctl) buildOverridesLayer(managementComponents *cluster.ManagementComponents, clusterName string, provider providers.Provider) error {
	// Adding cluster name to path temporarily following suggestion.
	//
	// This adds an implicit dependency between this method
	// and the writer passed to NewClusterctl
	// Ideally the writer implementation should be modified to
	// accept a path and file name and it should create the path in case it
	// does not exists.
	prefix := filepath.Join(clusterName, generatedDir, overridesDir)

	infraBundles := []types.InfrastructureBundle{
		{
			FolderName: filepath.Join("cert-manager", managementComponents.CertManager.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.CertManager.Manifest,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-kubeadm", managementComponents.Bootstrap.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.Bootstrap.Components,
				managementComponents.Bootstrap.Metadata,
			},
		},
		{
			FolderName: filepath.Join("cluster-api", managementComponents.ClusterAPI.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.ClusterAPI.Components,
				managementComponents.ClusterAPI.Metadata,
			},
		},
		{
			FolderName: filepath.Join("control-plane-kubeadm", managementComponents.ControlPlane.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.ControlPlane.Components,
				managementComponents.ControlPlane.Metadata,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-etcdadm-bootstrap", managementComponents.ExternalEtcdBootstrap.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.ExternalEtcdBootstrap.Components,
				managementComponents.ExternalEtcdBootstrap.Metadata,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-etcdadm-controller", managementComponents.ExternalEtcdController.Version),
			Manifests: []v1alpha1.Manifest{
				managementComponents.ExternalEtcdController.Components,
				managementComponents.ExternalEtcdController.Metadata,
			},
		},
	}

	infraBundles = append(infraBundles, *provider.GetInfrastructureBundle(managementComponents))
	for _, infraBundle := range infraBundles {
		if err := c.writeInfrastructureBundle(prefix, &infraBundle); err != nil {
			return err
		}
	}

	return nil
}

func (c *Clusterctl) writeInfrastructureBundle(rootFolder string, bundle *types.InfrastructureBundle) error {
	if bundle == nil {
		return nil
	}

	infraFolder := filepath.Join(rootFolder, bundle.FolderName)
	if err := os.MkdirAll(infraFolder, os.ModePerm); err != nil {
		return err
	}
	for _, manifest := range bundle.Manifests {
		m, err := bundles.ReadManifest(c.reader, manifest)
		if err != nil {
			return fmt.Errorf("can't load infrastructure bundle for manifest %s: %v", manifest.URI, err)
		}

		if err := os.WriteFile(filepath.Join(infraFolder, m.Filename), m.Content, 0o644); err != nil {
			return fmt.Errorf("generating file for infrastructure bundle %s: %v", m.Filename, err)
		}
	}

	return nil
}

// BackupManagement saves the CAPI resources of a cluster to the provided path. This will overwrite any existing contents
// in the path if the backup succeeds. If `clusterName` is provided, it filters and backs up only the provided cluster.
func (c *Clusterctl) BackupManagement(ctx context.Context, cluster *types.Cluster, managementStatePath, clusterName string) error {
	filePath := filepath.Join(".", cluster.Name, managementStatePath)

	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create backup file for CAPI objects: %v", err)
	}

	_, err = c.Execute(
		ctx, "move",
		"--to-directory", filePath,
		"--kubeconfig", cluster.KubeconfigFile,
		"--namespace", constants.EksaSystemNamespace,
		"--filter-cluster", clusterName,
	)
	if err != nil {
		return fmt.Errorf("failed taking backup of CAPI objects: %v", err)
	}
	return nil
}

// MoveManagement moves management components `from` cluster `to` cluster
// If `clusterName` is provided, it filters and moves only the provided cluster.
func (c *Clusterctl) MoveManagement(ctx context.Context, from, to *types.Cluster, clusterName string) error {
	params := []string{
		"move", "--to-kubeconfig", to.KubeconfigFile, "--namespace", constants.EksaSystemNamespace,
		"--filter-cluster", clusterName,
	}
	if from.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", from.KubeconfigFile)
	}

	_, err := c.Execute(
		ctx, params...,
	)
	if err != nil {
		return fmt.Errorf("failed moving management cluster: %v", err)
	}
	return err
}

func (c *Clusterctl) GetWorkloadKubeconfig(ctx context.Context, clusterName string, cluster *types.Cluster) ([]byte, error) {
	stdOut, err := c.Execute(
		ctx, "get", "kubeconfig", clusterName,
		"--kubeconfig", cluster.KubeconfigFile,
		"--namespace", constants.EksaSystemNamespace,
	)
	if err != nil {
		return nil, fmt.Errorf("executing get kubeconfig: %v", err)
	}
	return stdOut.Bytes(), nil
}

// InitInfrastructure initializes the infrastructure for the cluster using clusterctl.
func (c *Clusterctl) InitInfrastructure(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	if cluster == nil {
		return fmt.Errorf("invalid cluster (nil)")
	}
	if cluster.Name == "" {
		return fmt.Errorf("invalid cluster name '%s'", cluster.Name)
	}
	clusterctlConfig, err := c.buildConfig(managementComponents, cluster.Name, provider)
	if err != nil {
		return err
	}

	params := []string{
		"init",
		"--core", clusterctlConfig.coreVersion,
		"--bootstrap", clusterctlConfig.bootstrapVersion,
		"--control-plane", clusterctlConfig.controlPlaneVersion,
		"--infrastructure", fmt.Sprintf("%s:%s", provider.Name(), provider.Version(managementComponents)),
		"--config", clusterctlConfig.configFile,
		"--bootstrap", clusterctlConfig.etcdadmBootstrapVersion,
		"--bootstrap", clusterctlConfig.etcdadmControllerVersion,
	}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	envMap, err := provider.EnvMap(managementComponents, clusterSpec)
	if err != nil {
		return err
	}

	_, err = c.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return fmt.Errorf("executing init: %v", err)
	}

	return nil
}

func (c *Clusterctl) buildConfig(managementComponents *anywherecluster.ManagementComponents, clusterName string, provider providers.Provider) (*clusterctlConfiguration, error) {
	t := templater.New(c.writer)

	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"CertManagerInjectorRepository":                   imageRepository(managementComponents.CertManager.Cainjector),
		"CertManagerInjectorTag":                          managementComponents.CertManager.Cainjector.Tag(),
		"CertManagerControllerRepository":                 imageRepository(managementComponents.CertManager.Controller),
		"CertManagerControllerTag":                        managementComponents.CertManager.Controller.Tag(),
		"CertManagerWebhookRepository":                    imageRepository(managementComponents.CertManager.Webhook),
		"CertManagerWebhookTag":                           managementComponents.CertManager.Webhook.Tag(),
		"CertManagerVersion":                              managementComponents.CertManager.Version,
		"ClusterApiControllerRepository":                  imageRepository(managementComponents.ClusterAPI.Controller),
		"ClusterApiControllerTag":                         managementComponents.ClusterAPI.Controller.Tag(),
		"ClusterApiKubeRbacProxyRepository":               imageRepository(managementComponents.ClusterAPI.KubeProxy),
		"ClusterApiKubeRbacProxyTag":                      managementComponents.ClusterAPI.KubeProxy.Tag(),
		"KubeadmBootstrapControllerRepository":            imageRepository(managementComponents.Bootstrap.Controller),
		"KubeadmBootstrapControllerTag":                   managementComponents.Bootstrap.Controller.Tag(),
		"KubeadmBootstrapKubeRbacProxyRepository":         imageRepository(managementComponents.Bootstrap.KubeProxy),
		"KubeadmBootstrapKubeRbacProxyTag":                managementComponents.Bootstrap.KubeProxy.Tag(),
		"KubeadmControlPlaneControllerRepository":         imageRepository(managementComponents.ControlPlane.Controller),
		"KubeadmControlPlaneControllerTag":                managementComponents.ControlPlane.Controller.Tag(),
		"KubeadmControlPlaneKubeRbacProxyRepository":      imageRepository(managementComponents.ControlPlane.KubeProxy),
		"KubeadmControlPlaneKubeRbacProxyTag":             managementComponents.ControlPlane.KubeProxy.Tag(),
		"ClusterApiVSphereControllerRepository":           imageRepository(managementComponents.VSphere.ClusterAPIController),
		"ClusterApiVSphereControllerTag":                  managementComponents.VSphere.ClusterAPIController.Tag(),
		"ClusterApiNutanixControllerRepository":           imageRepository(managementComponents.Nutanix.ClusterAPIController),
		"ClusterApiNutanixControllerTag":                  managementComponents.Nutanix.ClusterAPIController.Tag(),
		"ClusterApiCloudStackManagerRepository":           imageRepository(managementComponents.CloudStack.ClusterAPIController),
		"ClusterApiCloudStackManagerTag":                  managementComponents.CloudStack.ClusterAPIController.Tag(),
		"ClusterApiCloudStackKubeRbacProxyRepository":     imageRepository(managementComponents.CloudStack.KubeRbacProxy),
		"ClusterApiCloudStackKubeRbacProxyTag":            managementComponents.CloudStack.KubeRbacProxy.Tag(),
		"ClusterApiVSphereKubeRbacProxyRepository":        imageRepository(managementComponents.VSphere.KubeProxy),
		"ClusterApiVSphereKubeRbacProxyTag":               managementComponents.VSphere.KubeProxy.Tag(),
		"DockerKubeRbacProxyRepository":                   imageRepository(managementComponents.Docker.KubeProxy),
		"DockerKubeRbacProxyTag":                          managementComponents.Docker.KubeProxy.Tag(),
		"DockerManagerRepository":                         imageRepository(managementComponents.Docker.Manager),
		"DockerManagerTag":                                managementComponents.Docker.Manager.Tag(),
		"EtcdadmBootstrapProviderRepository":              imageRepository(managementComponents.ExternalEtcdBootstrap.Controller),
		"EtcdadmBootstrapProviderTag":                     managementComponents.ExternalEtcdBootstrap.Controller.Tag(),
		"EtcdadmBootstrapProviderKubeRbacProxyRepository": imageRepository(managementComponents.ExternalEtcdBootstrap.KubeProxy),
		"EtcdadmBootstrapProviderKubeRbacProxyTag":        managementComponents.ExternalEtcdBootstrap.KubeProxy.Tag(),
		"EtcdadmControllerRepository":                     imageRepository(managementComponents.ExternalEtcdController.Controller),
		"EtcdadmControllerTag":                            managementComponents.ExternalEtcdController.Controller.Tag(),
		"EtcdadmControllerKubeRbacProxyRepository":        imageRepository(managementComponents.ExternalEtcdController.KubeProxy),
		"EtcdadmControllerKubeRbacProxyTag":               managementComponents.ExternalEtcdController.KubeProxy.Tag(),
		"DockerProviderVersion":                           managementComponents.Docker.Version,
		"VSphereProviderVersion":                          managementComponents.VSphere.Version,
		"CloudStackProviderVersion":                       managementComponents.CloudStack.Version,
		"SnowProviderVersion":                             managementComponents.Snow.Version,
		"TinkerbellProviderVersion":                       managementComponents.Tinkerbell.Version,
		"NutanixProviderVersion":                          managementComponents.Nutanix.Version,
		"ClusterApiProviderVersion":                       managementComponents.ClusterAPI.Version,
		"KubeadmControlPlaneProviderVersion":              managementComponents.ControlPlane.Version,
		"KubeadmBootstrapProviderVersion":                 managementComponents.Bootstrap.Version,
		"EtcdadmBootstrapProviderVersion":                 managementComponents.ExternalEtcdBootstrap.Version,
		"EtcdadmControllerProviderVersion":                managementComponents.ExternalEtcdController.Version,
		"dir":                                             path + "/" + clusterName + capiPrefix,
	}

	filePath, err := t.WriteToFile(clusterctlConfigTemplate, data, clusterctlConfigFile)
	if err != nil {
		return nil, fmt.Errorf("generating configuration file for clusterctl: %v", err)
	}
	if err := c.buildOverridesLayer(managementComponents, clusterName, provider); err != nil {
		return nil, err
	}

	return &clusterctlConfiguration{
		configFile:               filePath,
		bootstrapVersion:         fmt.Sprintf("%s:%s", kubeadmBootstrapProviderName, managementComponents.Bootstrap.Version),
		controlPlaneVersion:      fmt.Sprintf("kubeadm:%s", managementComponents.ControlPlane.Version),
		coreVersion:              fmt.Sprintf("cluster-api:%s", managementComponents.ClusterAPI.Version),
		etcdadmBootstrapVersion:  fmt.Sprintf("%s:%s", etcdadmBootstrapProviderName, managementComponents.ExternalEtcdBootstrap.Version),
		etcdadmControllerVersion: fmt.Sprintf("%s:%s", etcdadmControllerProviderName, managementComponents.ExternalEtcdController.Version),
	}, nil
}

var providerNamespaces = map[string]string{
	constants.VSphereProviderName:    constants.CapvSystemNamespace,
	constants.DockerProviderName:     constants.CapdSystemNamespace,
	constants.CloudStackProviderName: constants.CapcSystemNamespace,
	constants.AWSProviderName:        constants.CapaSystemNamespace,
	constants.SnowProviderName:       constants.CapasSystemNamespace,
	constants.NutanixProviderName:    constants.CapxSystemNamespace,
	constants.TinkerbellProviderName: constants.CaptSystemNamespace,
	etcdadmBootstrapProviderName:     constants.EtcdAdmBootstrapProviderSystemNamespace,
	etcdadmControllerProviderName:    constants.EtcdAdmControllerSystemNamespace,
	kubeadmBootstrapProviderName:     constants.CapiKubeadmBootstrapSystemNamespace,
}

// Upgrade executes an upgrade of the cluster to the new management components and the spec.
func (c *Clusterctl) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, managementComponents *cluster.ManagementComponents, newSpec *cluster.Spec, changeDiff *clusterapi.CAPIChangeDiff) error {
	clusterctlConfig, err := c.buildConfig(managementComponents, managementCluster.Name, provider)
	if err != nil {
		return err
	}

	upgradeCommand := []string{
		"upgrade", "apply",
		"--config", clusterctlConfig.configFile,
		"--kubeconfig", managementCluster.KubeconfigFile,
	}

	if changeDiff.ControlPlane != nil {
		upgradeCommand = append(upgradeCommand, "--control-plane", fmt.Sprintf("%s/kubeadm:%s", constants.CapiKubeadmControlPlaneSystemNamespace, changeDiff.ControlPlane.NewVersion))
	}

	if changeDiff.Core != nil {
		upgradeCommand = append(upgradeCommand, "--core", fmt.Sprintf("%s/cluster-api:%s", constants.CapiSystemNamespace, changeDiff.Core.NewVersion))
	}

	if changeDiff.InfrastructureProvider != nil {
		newInfraProvider := fmt.Sprintf("%s/%s:%s", providerNamespaces[changeDiff.InfrastructureProvider.ComponentName], changeDiff.InfrastructureProvider.ComponentName, changeDiff.InfrastructureProvider.NewVersion)
		upgradeCommand = append(upgradeCommand, "--infrastructure", newInfraProvider)
	}

	for _, bootstrapProvider := range changeDiff.BootstrapProviders {
		newBootstrapProvider := fmt.Sprintf("%s/%s:%s", providerNamespaces[bootstrapProvider.ComponentName], bootstrapProvider.ComponentName, bootstrapProvider.NewVersion)
		upgradeCommand = append(upgradeCommand, "--bootstrap", newBootstrapProvider)
	}

	providerEnvMap, err := provider.EnvMap(managementComponents, newSpec)
	if err != nil {
		return fmt.Errorf("failed generating provider env map for clusterctl upgrade: %v", err)
	}

	if _, err = c.ExecuteWithEnv(ctx, providerEnvMap, upgradeCommand...); err != nil {
		return fmt.Errorf("failed running upgrade apply with clusterctl: %v", err)
	}

	return nil
}

// InstallEtcdadmProviders installs the etcdadm providers for the cluster using clusterctl.
func (c *Clusterctl) InstallEtcdadmProviders(ctx context.Context, managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec, cluster *types.Cluster, infraProvider providers.Provider, installProviders []string) error {
	if cluster == nil {
		return fmt.Errorf("invalid cluster (nil)")
	}
	if cluster.Name == "" {
		return fmt.Errorf("invalid cluster name '%s'", cluster.Name)
	}

	clusterctlConfig, err := c.buildConfig(managementComponents, cluster.Name, infraProvider)
	if err != nil {
		return err
	}

	params := []string{
		"init",
		"--config", clusterctlConfig.configFile,
	}

	for _, provider := range installProviders {
		switch provider {
		case constants.EtcdAdmBootstrapProviderName:
			params = append(params, "--bootstrap", clusterctlConfig.etcdadmBootstrapVersion)
		case constants.EtcdadmControllerProviderName:
			params = append(params, "--bootstrap", clusterctlConfig.etcdadmControllerVersion)
		default:
			return fmt.Errorf("unrecognized capi provider %s", provider)
		}
	}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	envMap, err := infraProvider.EnvMap(managementComponents, clusterSpec)
	if err != nil {
		return err
	}

	_, err = c.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return fmt.Errorf("executing init: %v", err)
	}

	return nil
}
