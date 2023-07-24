package executables

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
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
func (c *Clusterctl) buildOverridesLayer(clusterSpec *cluster.Spec, clusterName string, provider providers.Provider) error {
	versionsBundle := clusterSpec.ControlPlaneVersionsBundle()

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
			FolderName: filepath.Join("cert-manager", versionsBundle.CertManager.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.CertManager.Manifest,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-kubeadm", versionsBundle.Bootstrap.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.Bootstrap.Components,
				versionsBundle.Bootstrap.Metadata,
			},
		},
		{
			FolderName: filepath.Join("cluster-api", versionsBundle.ClusterAPI.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.ClusterAPI.Components,
				versionsBundle.ClusterAPI.Metadata,
			},
		},
		{
			FolderName: filepath.Join("control-plane-kubeadm", versionsBundle.ControlPlane.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.ControlPlane.Components,
				versionsBundle.ControlPlane.Metadata,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-etcdadm-bootstrap", versionsBundle.ExternalEtcdBootstrap.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.ExternalEtcdBootstrap.Components,
				versionsBundle.ExternalEtcdBootstrap.Metadata,
			},
		},
		{
			FolderName: filepath.Join("bootstrap-etcdadm-controller", versionsBundle.ExternalEtcdController.Version),
			Manifests: []v1alpha1.Manifest{
				versionsBundle.ExternalEtcdController.Components,
				versionsBundle.ExternalEtcdController.Metadata,
			},
		},
	}

	infraBundles = append(infraBundles, *provider.GetInfrastructureBundle(clusterSpec))
	for _, infraBundle := range infraBundles {
		if err := c.writeInfrastructureBundle(clusterSpec, prefix, &infraBundle); err != nil {
			return err
		}
	}

	return nil
}

func (c *Clusterctl) writeInfrastructureBundle(clusterSpec *cluster.Spec, rootFolder string, bundle *types.InfrastructureBundle) error {
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
// in the path if the backup succeeds.
func (c *Clusterctl) BackupManagement(ctx context.Context, cluster *types.Cluster, managementStatePath string) error {
	filePath := filepath.Join(c.writer.Dir(), managementStatePath)

	// check for existing backup to prevent partial overwrites
	_, err := os.Stat(filePath)
	if err == nil {
		tempPath := filepath.Join(c.writer.TempDir(), managementStatePath)
		defer func() {
			os.RemoveAll(tempPath)
		}()
		err = c.backupManagement(ctx, cluster, tempPath)
		if err != nil {
			return err
		}

		return ReplacePath(tempPath, filePath)
	}
	return c.backupManagement(ctx, cluster, filePath)
}

func (c *Clusterctl) backupManagement(ctx context.Context, cluster *types.Cluster, filePath string) error {
	err := os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("could not create backup file for CAPI objects: %v", err)
	}
	_, err = c.Execute(
		ctx, "move",
		"--to-directory", filePath,
		"--kubeconfig", cluster.KubeconfigFile,
		"--namespace", constants.EksaSystemNamespace,
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

func (c *Clusterctl) InitInfrastructure(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, provider providers.Provider) error {
	if cluster == nil {
		return fmt.Errorf("invalid cluster (nil)")
	}
	if cluster.Name == "" {
		return fmt.Errorf("invalid cluster name '%s'", cluster.Name)
	}
	clusterctlConfig, err := c.buildConfig(clusterSpec, cluster.Name, provider)
	if err != nil {
		return err
	}

	params := []string{
		"init",
		"--core", clusterctlConfig.coreVersion,
		"--bootstrap", clusterctlConfig.bootstrapVersion,
		"--control-plane", clusterctlConfig.controlPlaneVersion,
		"--infrastructure", fmt.Sprintf("%s:%s", provider.Name(), provider.Version(clusterSpec)),
		"--config", clusterctlConfig.configFile,
		"--bootstrap", clusterctlConfig.etcdadmBootstrapVersion,
		"--bootstrap", clusterctlConfig.etcdadmControllerVersion,
	}

	if cluster.KubeconfigFile != "" {
		params = append(params, "--kubeconfig", cluster.KubeconfigFile)
	}

	envMap, err := provider.EnvMap(clusterSpec)
	if err != nil {
		return err
	}

	_, err = c.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return fmt.Errorf("executing init: %v", err)
	}

	return nil
}

func (c *Clusterctl) buildConfig(clusterSpec *cluster.Spec, clusterName string, provider providers.Provider) (*clusterctlConfiguration, error) {
	t := templater.New(c.writer)
	versionsBundle := clusterSpec.ControlPlaneVersionsBundle()

	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	data := map[string]string{
		"CertManagerInjectorRepository":                   imageRepository(versionsBundle.CertManager.Cainjector),
		"CertManagerInjectorTag":                          versionsBundle.CertManager.Cainjector.Tag(),
		"CertManagerControllerRepository":                 imageRepository(versionsBundle.CertManager.Controller),
		"CertManagerControllerTag":                        versionsBundle.CertManager.Controller.Tag(),
		"CertManagerWebhookRepository":                    imageRepository(versionsBundle.CertManager.Webhook),
		"CertManagerWebhookTag":                           versionsBundle.CertManager.Webhook.Tag(),
		"CertManagerVersion":                              versionsBundle.CertManager.Version,
		"ClusterApiControllerRepository":                  imageRepository(versionsBundle.ClusterAPI.Controller),
		"ClusterApiControllerTag":                         versionsBundle.ClusterAPI.Controller.Tag(),
		"ClusterApiKubeRbacProxyRepository":               imageRepository(versionsBundle.ClusterAPI.KubeProxy),
		"ClusterApiKubeRbacProxyTag":                      versionsBundle.ClusterAPI.KubeProxy.Tag(),
		"KubeadmBootstrapControllerRepository":            imageRepository(versionsBundle.Bootstrap.Controller),
		"KubeadmBootstrapControllerTag":                   versionsBundle.Bootstrap.Controller.Tag(),
		"KubeadmBootstrapKubeRbacProxyRepository":         imageRepository(versionsBundle.Bootstrap.KubeProxy),
		"KubeadmBootstrapKubeRbacProxyTag":                versionsBundle.Bootstrap.KubeProxy.Tag(),
		"KubeadmControlPlaneControllerRepository":         imageRepository(versionsBundle.ControlPlane.Controller),
		"KubeadmControlPlaneControllerTag":                versionsBundle.ControlPlane.Controller.Tag(),
		"KubeadmControlPlaneKubeRbacProxyRepository":      imageRepository(versionsBundle.ControlPlane.KubeProxy),
		"KubeadmControlPlaneKubeRbacProxyTag":             versionsBundle.ControlPlane.KubeProxy.Tag(),
		"ClusterApiVSphereControllerRepository":           imageRepository(versionsBundle.VSphere.ClusterAPIController),
		"ClusterApiVSphereControllerTag":                  versionsBundle.VSphere.ClusterAPIController.Tag(),
		"ClusterApiNutanixControllerRepository":           imageRepository(versionsBundle.Nutanix.ClusterAPIController),
		"ClusterApiNutanixControllerTag":                  versionsBundle.Nutanix.ClusterAPIController.Tag(),
		"ClusterApiCloudStackManagerRepository":           imageRepository(versionsBundle.CloudStack.ClusterAPIController),
		"ClusterApiCloudStackManagerTag":                  versionsBundle.CloudStack.ClusterAPIController.Tag(),
		"ClusterApiCloudStackKubeRbacProxyRepository":     imageRepository(versionsBundle.CloudStack.KubeRbacProxy),
		"ClusterApiCloudStackKubeRbacProxyTag":            versionsBundle.CloudStack.KubeRbacProxy.Tag(),
		"ClusterApiVSphereKubeRbacProxyRepository":        imageRepository(versionsBundle.VSphere.KubeProxy),
		"ClusterApiVSphereKubeRbacProxyTag":               versionsBundle.VSphere.KubeProxy.Tag(),
		"DockerKubeRbacProxyRepository":                   imageRepository(versionsBundle.Docker.KubeProxy),
		"DockerKubeRbacProxyTag":                          versionsBundle.Docker.KubeProxy.Tag(),
		"DockerManagerRepository":                         imageRepository(versionsBundle.Docker.Manager),
		"DockerManagerTag":                                versionsBundle.Docker.Manager.Tag(),
		"EtcdadmBootstrapProviderRepository":              imageRepository(versionsBundle.ExternalEtcdBootstrap.Controller),
		"EtcdadmBootstrapProviderTag":                     versionsBundle.ExternalEtcdBootstrap.Controller.Tag(),
		"EtcdadmBootstrapProviderKubeRbacProxyRepository": imageRepository(versionsBundle.ExternalEtcdBootstrap.KubeProxy),
		"EtcdadmBootstrapProviderKubeRbacProxyTag":        versionsBundle.ExternalEtcdBootstrap.KubeProxy.Tag(),
		"EtcdadmControllerRepository":                     imageRepository(versionsBundle.ExternalEtcdController.Controller),
		"EtcdadmControllerTag":                            versionsBundle.ExternalEtcdController.Controller.Tag(),
		"EtcdadmControllerKubeRbacProxyRepository":        imageRepository(versionsBundle.ExternalEtcdController.KubeProxy),
		"EtcdadmControllerKubeRbacProxyTag":               versionsBundle.ExternalEtcdController.KubeProxy.Tag(),
		"DockerProviderVersion":                           versionsBundle.Docker.Version,
		"VSphereProviderVersion":                          versionsBundle.VSphere.Version,
		"CloudStackProviderVersion":                       versionsBundle.CloudStack.Version,
		"SnowProviderVersion":                             versionsBundle.Snow.Version,
		"TinkerbellProviderVersion":                       versionsBundle.Tinkerbell.Version,
		"NutanixProviderVersion":                          versionsBundle.Nutanix.Version,
		"ClusterApiProviderVersion":                       versionsBundle.ClusterAPI.Version,
		"KubeadmControlPlaneProviderVersion":              versionsBundle.ControlPlane.Version,
		"KubeadmBootstrapProviderVersion":                 versionsBundle.Bootstrap.Version,
		"EtcdadmBootstrapProviderVersion":                 versionsBundle.ExternalEtcdBootstrap.Version,
		"EtcdadmControllerProviderVersion":                versionsBundle.ExternalEtcdController.Version,
		"dir":                                             path + "/" + clusterName + capiPrefix,
	}

	filePath, err := t.WriteToFile(clusterctlConfigTemplate, data, clusterctlConfigFile)
	if err != nil {
		return nil, fmt.Errorf("generating configuration file for clusterctl: %v", err)
	}
	if err := c.buildOverridesLayer(clusterSpec, clusterName, provider); err != nil {
		return nil, err
	}

	return &clusterctlConfiguration{
		configFile:               filePath,
		bootstrapVersion:         fmt.Sprintf("%s:%s", kubeadmBootstrapProviderName, versionsBundle.Bootstrap.Version),
		controlPlaneVersion:      fmt.Sprintf("kubeadm:%s", versionsBundle.ControlPlane.Version),
		coreVersion:              fmt.Sprintf("cluster-api:%s", versionsBundle.ClusterAPI.Version),
		etcdadmBootstrapVersion:  fmt.Sprintf("%s:%s", etcdadmBootstrapProviderName, versionsBundle.ExternalEtcdBootstrap.Version),
		etcdadmControllerVersion: fmt.Sprintf("%s:%s", etcdadmControllerProviderName, versionsBundle.ExternalEtcdController.Version),
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

func (c *Clusterctl) Upgrade(ctx context.Context, managementCluster *types.Cluster, provider providers.Provider, newSpec *cluster.Spec, changeDiff *clusterapi.CAPIChangeDiff) error {
	clusterctlConfig, err := c.buildConfig(newSpec, managementCluster.Name, provider)
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

	providerEnvMap, err := provider.EnvMap(newSpec)
	if err != nil {
		return fmt.Errorf("failed generating provider env map for clusterctl upgrade: %v", err)
	}

	if _, err = c.ExecuteWithEnv(ctx, providerEnvMap, upgradeCommand...); err != nil {
		return fmt.Errorf("failed running upgrade apply with clusterctl: %v", err)
	}

	return nil
}

func (c *Clusterctl) InstallEtcdadmProviders(ctx context.Context, clusterSpec *cluster.Spec, cluster *types.Cluster, infraProvider providers.Provider, installProviders []string) error {
	if cluster == nil {
		return fmt.Errorf("invalid cluster (nil)")
	}
	if cluster.Name == "" {
		return fmt.Errorf("invalid cluster name '%s'", cluster.Name)
	}
	clusterctlConfig, err := c.buildConfig(clusterSpec, cluster.Name, infraProvider)
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

	envMap, err := infraProvider.EnvMap(clusterSpec)
	if err != nil {
		return err
	}

	_, err = c.ExecuteWithEnv(ctx, envMap, params...)
	if err != nil {
		return fmt.Errorf("executing init: %v", err)
	}

	return nil
}

// ReplacePath replaces the contents of newpath with oldpath. The contents of newpath are first saved off to a temp directory
// before being replaced to prevent overwriting existing files on os.Rename failure.
func ReplacePath(oldpath string, newpath string) error {
	tempPath := fmt.Sprintf("%s_temp", newpath)
	err := os.Rename(newpath, tempPath)
	if err != nil {
		return fmt.Errorf("renaming new path to temp: %v", err)
	}
	err = os.Rename(oldpath, newpath)
	if err != nil {
		return fmt.Errorf("renaming old path: %v", err)
	}
	err = os.RemoveAll(tempPath)
	if err != nil {
		return fmt.Errorf("removing temp path: %v", err)
	}
	return nil
}
