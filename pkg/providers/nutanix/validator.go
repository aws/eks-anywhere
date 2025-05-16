package nutanix

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"k8s.io/apimachinery/pkg/api/resource"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const (
	minNutanixCPUSockets   = 1
	minNutanixCPUPerSocket = 1
	minNutanixMemoryMiB    = 2048
	minNutanixDiskGiB      = 20
)

// IPValidator is an interface that defines methods to validate the control plane IP.
type IPValidator interface {
	ValidateControlPlaneIPUniqueness(cluster *anywherev1.Cluster) error
}

// Validator is a client to validate nutanix resources.
type Validator struct {
	httpClient    *http.Client
	certValidator crypto.TlsValidator
	clientCache   *ClientCache
}

// NewValidator returns a new validator client.
func NewValidator(clientCache *ClientCache, certValidator crypto.TlsValidator, httpClient *http.Client) *Validator {
	return &Validator{
		clientCache:   clientCache,
		certValidator: certValidator,
		httpClient:    httpClient,
	}
}

func (v *Validator) validateControlPlaneIP(ip string) error {
	// check if controlPlaneEndpointIp is valid
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s", ip)
	}
	return nil
}

// ValidateClusterSpec validates the cluster spec.
func (v *Validator) ValidateClusterSpec(ctx context.Context, spec *cluster.Spec, creds credentials.BasicAuthCredential) error {
	logger.Info("ValidateClusterSpec for Nutanix datacenter", "NutanixDatacenter", spec.NutanixDatacenter.Name)
	client, err := v.clientCache.GetNutanixClient(spec.NutanixDatacenter, creds)
	if err != nil {
		return err
	}

	if err := v.ValidateDatacenterConfig(ctx, client, spec); err != nil {
		return err
	}

	if err := v.validateControlPlaneIP(spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err != nil {
		return err
	}

	for _, conf := range spec.NutanixMachineConfigs {
		if err := v.ValidateMachineConfig(ctx, client, spec.Cluster, conf); err != nil {
			return fmt.Errorf("failed to validate machine config: %v", err)
		}
	}

	if err := v.validateFreeGPU(ctx, client, spec); err != nil {
		return err
	}

	return v.checkImageNameMatchesKubernetesVersion(ctx, spec, client)
}

func (v *Validator) checkImageNameMatchesKubernetesVersion(ctx context.Context, spec *cluster.Spec, client Client) error {
	controlPlaneMachineConfig := spec.NutanixMachineConfigs[spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find NutanixMachineConfig %v for control plane", spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}
	// validate template field name contains cluster kubernetes version for the control plane machine.
	if err := v.validateTemplateMatchesKubernetesVersion(ctx, controlPlaneMachineConfig.Spec.Image, client, string(spec.Cluster.Spec.KubernetesVersion)); err != nil {
		return fmt.Errorf("machine config %s validation failed: %v", controlPlaneMachineConfig.Name, err)
	}

	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := spec.NutanixMachineConfigs[spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name]
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find NutanixMachineConfig %v for etcd machines", spec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		// validate template field name contains cluster kubernetes version for the external etcd machine.
		if err := v.validateTemplateMatchesKubernetesVersion(ctx, etcdMachineConfig.Spec.Image, client, string(spec.Cluster.Spec.KubernetesVersion)); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", etcdMachineConfig.Name, err)
		}
	}

	for _, workerNodeGroupConfiguration := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		kubernetesVersion := string(spec.Cluster.Spec.KubernetesVersion)
		if workerNodeGroupConfiguration.KubernetesVersion != nil {
			kubernetesVersion = string(*workerNodeGroupConfiguration.KubernetesVersion)
		}
		// validate template field name contains cluster kubernetes version for the control plane machine.
		imageIdentifier := spec.NutanixMachineConfigs[workerNodeGroupConfiguration.MachineGroupRef.Name].Spec.Image
		if err := v.validateTemplateMatchesKubernetesVersion(ctx, imageIdentifier, client, kubernetesVersion); err != nil {
			return fmt.Errorf("machine config %s validation failed: %v", controlPlaneMachineConfig.Name, err)
		}
	}
	return nil
}

// ValidateDatacenterConfig validates the datacenter config.
func (v *Validator) ValidateDatacenterConfig(ctx context.Context, client Client, spec *cluster.Spec) error {
	config := spec.NutanixDatacenter
	if config.Spec.Insecure {
		logger.Info("Warning: Skipping TLS validation for insecure connection to Nutanix Prism Central; this is not recommended for production use")
	}

	if err := v.validateEndpointAndPort(config.Spec); err != nil {
		return err
	}

	if err := v.validateCredentials(ctx, client); err != nil {
		return err
	}

	if err := v.validateTrustBundleConfig(config.Spec); err != nil {
		return err
	}

	if err := v.validateCredentialRef(config); err != nil {
		return err
	}

	if err := v.validateFailureDomains(ctx, client, spec); err != nil {
		return err
	}

	if config.Spec.CcmExcludeNodeIPs != nil {
		if err := v.validateCcmExcludeNodeIPs(config.Spec.CcmExcludeNodeIPs); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateIPRangeForCcmExcludeNodeIPs(ipRange string) error {
	ipRangeStr := strings.TrimSpace(ipRange)
	rangeParts := strings.Split(ipRangeStr, "-")
	if len(rangeParts) != 2 {
		return fmt.Errorf("invalid IP range %s", ipRangeStr)
	}
	startIP := net.ParseIP(strings.TrimSpace(rangeParts[0]))
	if startIP == nil {
		return fmt.Errorf("invalid start IP address %s", rangeParts[0])
	}
	endIP := net.ParseIP(strings.TrimSpace(rangeParts[1]))
	if endIP == nil {
		return fmt.Errorf("invalid end IP address %s", rangeParts[1])
	}
	cmp, err := compareIP(startIP, endIP)
	if err != nil {
		return err
	}
	if cmp > 0 {
		return fmt.Errorf("start IP address %s is greater than end IP address %s", startIP.String(), endIP.String())
	}

	return nil
}

func (v *Validator) validateCcmExcludeNodeIPs(ccmExcludeNodeIPs []string) error {
	for _, ipOrIPRange := range ccmExcludeNodeIPs {
		if strings.Contains(ipOrIPRange, "/") {
			cidrStr := strings.TrimSpace(ipOrIPRange)
			_, _, err := net.ParseCIDR(cidrStr)
			if err != nil {
				return fmt.Errorf("invalid CIDR %s: %v", cidrStr, err)
			}
		} else if strings.Contains(ipOrIPRange, "-") {
			err := v.validateIPRangeForCcmExcludeNodeIPs(ipOrIPRange)
			if err != nil {
				return err
			}
		} else {
			ipStr := strings.TrimSpace(ipOrIPRange)
			ip := net.ParseIP(ipStr)
			if ip == nil {
				return fmt.Errorf("invalid IP address %s", ipStr)
			}
		}
	}

	return nil
}

func (v *Validator) validateFailureDomains(ctx context.Context, client Client, spec *cluster.Spec) error {
	config := spec.NutanixDatacenter

	regexName, err := regexp.Compile("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")
	if err != nil {
		return err
	}

	failureDomainCount := len(config.Spec.FailureDomains)
	for _, fd := range config.Spec.FailureDomains {
		if res := regexName.MatchString(fd.Name); !res {
			errorStr := `failure domain name should contains only small letters, digits, and hyphens.
			it should start with small letter or digit`
			return fmt.Errorf(errorStr)
		}

		if err := v.validateClusterConfig(ctx, client, fd.Cluster); err != nil {
			return err
		}

		for _, subnet := range fd.Subnets {
			if err := v.validateSubnetConfig(ctx, client, fd.Cluster, subnet); err != nil {
				return err
			}
		}

		workerMachineGroups := getWorkerMachineGroups(spec)
		for _, workerMachineGroupName := range fd.WorkerMachineGroups {
			if err := v.validateWorkerMachineGroup(workerMachineGroups, workerMachineGroupName, failureDomainCount); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateWorkerMachineGroup(workerMachineGroups map[string]anywherev1.WorkerNodeGroupConfiguration, workerMachineGroupName string, fdCount int) error {
	if _, ok := workerMachineGroups[workerMachineGroupName]; !ok {
		return fmt.Errorf("worker machine group %s not found in the cluster worker node group definitions", workerMachineGroupName)
	}

	if workerMachineGroups[workerMachineGroupName].Count != nil && *workerMachineGroups[workerMachineGroupName].Count > fdCount {
		return fmt.Errorf("count %d of machines in workerNodeGroupConfiguration %s shouldn't be greater than the failure domain count %d where those machines should be spreaded accross", *workerMachineGroups[workerMachineGroupName].Count, workerMachineGroupName, fdCount)
	}

	return nil
}

func (v *Validator) validateCredentialRef(config *anywherev1.NutanixDatacenterConfig) error {
	if config.Spec.CredentialRef == nil {
		return fmt.Errorf("credentialRef must be provided")
	}

	if config.Spec.CredentialRef.Kind != constants.SecretKind {
		return fmt.Errorf("credentialRef kind must be %s", constants.SecretKind)
	}

	if config.Spec.CredentialRef.Name == "" {
		return fmt.Errorf("credentialRef name must be provided")
	}

	return nil
}

func (v *Validator) validateEndpointAndPort(dcConf anywherev1.NutanixDatacenterConfigSpec) error {
	if !networkutils.IsPortValid(strconv.Itoa(dcConf.Port)) {
		return fmt.Errorf("nutanix prism central port %d out of range", dcConf.Port)
	}

	if dcConf.Endpoint == "" {
		return fmt.Errorf("nutanix prism central endpoint must be provided")
	}
	server := fmt.Sprintf("%s:%d", dcConf.Endpoint, dcConf.Port)
	if !strings.HasPrefix(server, "https://") {
		server = fmt.Sprintf("https://%s", server)
	}

	if _, err := v.httpClient.Get(server); err != nil {
		return fmt.Errorf("failed to reach server %s: %v", server, err)
	}

	return nil
}

func (v *Validator) validateCredentials(ctx context.Context, client Client) error {
	_, err := client.GetCurrentLoggedInUser(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateTrustBundleConfig(dcConf anywherev1.NutanixDatacenterConfigSpec) error {
	if dcConf.AdditionalTrustBundle == "" {
		return nil
	}

	return v.certValidator.ValidateCert(dcConf.Endpoint, fmt.Sprintf("%d", dcConf.Port), dcConf.AdditionalTrustBundle)
}

func (v *Validator) validateMachineSpecs(machineSpec anywherev1.NutanixMachineConfigSpec) error {
	if machineSpec.VCPUSockets < minNutanixCPUSockets {
		return fmt.Errorf("vCPU sockets %d must be greater than or equal to %d", machineSpec.VCPUSockets, minNutanixCPUSockets)
	}

	if machineSpec.VCPUsPerSocket < minNutanixCPUPerSocket {
		return fmt.Errorf("vCPUs per socket %d must be greater than or equal to %d", machineSpec.VCPUsPerSocket, minNutanixCPUPerSocket)
	}

	minNutanixMemory, err := resource.ParseQuantity(fmt.Sprintf("%dMi", minNutanixMemoryMiB))
	if err != nil {
		return err
	}

	if machineSpec.MemorySize.Cmp(minNutanixMemory) < 0 {
		return fmt.Errorf("MemorySize must be greater than or equal to %dMi", minNutanixMemoryMiB)
	}

	minNutanixDisk, err := resource.ParseQuantity(fmt.Sprintf("%dGi", minNutanixDiskGiB))
	if err != nil {
		return err
	}

	if machineSpec.SystemDiskSize.Cmp(minNutanixDisk) < 0 {
		return fmt.Errorf("SystemDiskSize must be greater than or equal to %dGi", minNutanixDiskGiB)
	}

	if machineSpec.BootType != "" && machineSpec.BootType != anywherev1.NutanixBootTypeLegacy && machineSpec.BootType != anywherev1.NutanixBootTypeUEFI {
		return fmt.Errorf("boot type %s is not supported, only legacy and uefi are supported", machineSpec.BootType)
	}

	return nil
}

// ValidateMachineConfig validates the Prism Element cluster, subnet, and image for the machine.
func (v *Validator) ValidateMachineConfig(ctx context.Context, client Client, cluster *anywherev1.Cluster, config *anywherev1.NutanixMachineConfig) error {
	if err := v.validateMachineSpecs(config.Spec); err != nil {
		return err
	}

	if err := v.validateClusterConfig(ctx, client, config.Spec.Cluster); err != nil {
		return err
	}

	if err := v.validateSubnetConfig(ctx, client, config.Spec.Cluster, config.Spec.Subnet); err != nil {
		return err
	}

	if err := v.validateImageConfig(ctx, client, config.Spec.Image); err != nil {
		return err
	}

	if config.Spec.Project != nil {
		if err := v.validateProjectConfig(ctx, client, *config.Spec.Project); err != nil {
			return err
		}
	}

	if config.Spec.AdditionalCategories != nil {
		if err := v.validateAdditionalCategories(ctx, client, config.Spec.AdditionalCategories); err != nil {
			return err
		}
	}

	if err := v.validateGPUInMachineConfig(cluster, config); err != nil {
		return err
	}

	if err := v.validateBootTypeMachineConfig(config); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateGPUInMachineConfig(cluster *anywherev1.Cluster, config *anywherev1.NutanixMachineConfig) error {
	if config.Spec.GPUs != nil {
		if err := checkMachineConfigIsForWorker(config, cluster); err != nil {
			return err
		}

		for _, gpu := range config.Spec.GPUs {
			if err := v.validateGPUConfig(gpu); err != nil {
				return err
			}
		}
	}

	return nil
}

func (v *Validator) validateBootTypeMachineConfig(config *anywherev1.NutanixMachineConfig) error {
	if config.Spec.BootType != "" {
		if err := v.validateBootType(config.Spec.BootType); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateClusterConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing cluster name")
		} else {
			clusterName := *identifier.Name
			if _, err := findClusterUUIDByName(ctx, client, clusterName); err != nil {
				return fmt.Errorf("failed to find cluster with name %q: %v", clusterName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing cluster uuid")
		} else {
			clusterUUID := *identifier.UUID
			if _, err := client.GetCluster(ctx, clusterUUID); err != nil {
				return fmt.Errorf("failed to find cluster with uuid %v: %v", clusterUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid cluster identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateImageConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing image name")
		} else {
			imageName := *identifier.Name
			if _, err := findImageUUIDByName(ctx, client, imageName); err != nil {
				return fmt.Errorf("failed to find image with name %q: %v", imageName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing image uuid")
		} else {
			imageUUID := *identifier.UUID
			if _, err := client.GetImage(ctx, imageUUID); err != nil {
				return fmt.Errorf("failed to find image with uuid %s: %v", imageUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid image identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateTemplateMatchesKubernetesVersion(ctx context.Context, identifier anywherev1.NutanixResourceIdentifier, client Client, kubernetesVersionName string) error {
	var templateName string
	if identifier.Type == anywherev1.NutanixIdentifierUUID {
		imageUUID := *identifier.UUID
		imageDetails, err := client.GetImage(ctx, imageUUID)
		if err != nil {
			return fmt.Errorf("failed to find image with uuid %s: %v", imageUUID, err)
		}
		if imageDetails.Spec == nil || imageDetails.Spec.Name == nil {
			return fmt.Errorf("failed to find image details with uuid %s", imageUUID)
		}
		templateName = *imageDetails.Spec.Name
	} else {
		templateName = *identifier.Name
	}

	// Replace 1.23, 1-23, 1_23 to 123 in the template name string.
	templateReplacer := strings.NewReplacer("-", "", ".", "", "_", "")
	template := templateReplacer.Replace(templateName)
	// Replace 1-23 to 123 in the kubernetesversion string.
	replacer := strings.NewReplacer(".", "")
	kubernetesVersion := replacer.Replace(string(kubernetesVersionName))
	// This will return an error if the template name does not contain specified kubernetes version.
	// For ex if the kubernetes version is 1.23,
	// the template name should include 1.23 or 1-23, 1_23 or 123 i.e. kubernetes-1-23-eks in the string.
	if !strings.Contains(template, kubernetesVersion) {
		return fmt.Errorf("missing kube version from the machine config template name: template=%s, version=%s", templateName, string(kubernetesVersionName))
	}
	return nil
}

func (v *Validator) validateSubnetConfig(ctx context.Context, client Client, cluster, subnet anywherev1.NutanixResourceIdentifier) error {
	clusterUUID, err := getClusterUUID(ctx, client, cluster)
	if err != nil {
		return err
	}

	switch subnet.Type {
	case anywherev1.NutanixIdentifierName:
		if subnet.Name == nil || *subnet.Name == "" {
			return fmt.Errorf("missing subnet name")
		} else {
			subnetName := *subnet.Name
			if _, err = findSubnetUUIDByName(ctx, client, clusterUUID, subnetName); err != nil {
				return fmt.Errorf("failed to find subnet with name %s: %v", subnetName, err)
			}
		}
	case anywherev1.NutanixIdentifierUUID:
		if subnet.UUID == nil || *subnet.UUID == "" {
			return fmt.Errorf("missing subnet uuid")
		} else {
			subnetUUID := *subnet.UUID
			if _, err = client.GetSubnet(ctx, subnetUUID); err != nil {
				return fmt.Errorf("failed to find subnet with uuid %s: %v", subnetUUID, err)
			}
		}
	default:
		return fmt.Errorf("invalid subnet identifier type: %s; valid types are: %q and %q", subnet.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateProjectConfig(ctx context.Context, client Client, identifier anywherev1.NutanixResourceIdentifier) error {
	switch identifier.Type {
	case anywherev1.NutanixIdentifierName:
		if identifier.Name == nil || *identifier.Name == "" {
			return fmt.Errorf("missing project name")
		}
		projectName := *identifier.Name
		if _, err := findProjectUUIDByName(ctx, client, projectName); err != nil {
			return fmt.Errorf("failed to find project with name %q: %v", projectName, err)
		}
	case anywherev1.NutanixIdentifierUUID:
		if identifier.UUID == nil || *identifier.UUID == "" {
			return fmt.Errorf("missing project uuid")
		}
		projectUUID := *identifier.UUID
		if _, err := client.GetProject(ctx, projectUUID); err != nil {
			return fmt.Errorf("failed to find project with uuid %s: %v", projectUUID, err)
		}
	default:
		return fmt.Errorf("invalid project identifier type: %s; valid types are: %q and %q", identifier.Type, anywherev1.NutanixIdentifierName, anywherev1.NutanixIdentifierUUID)
	}

	return nil
}

func (v *Validator) validateAdditionalCategories(ctx context.Context, client Client, categories []anywherev1.NutanixCategoryIdentifier) error {
	for _, category := range categories {
		if category.Key == "" {
			return fmt.Errorf("missing category key")
		}

		if category.Value == "" {
			return fmt.Errorf("missing category value")
		}

		if _, err := client.GetCategoryKey(ctx, category.Key); err != nil {
			return fmt.Errorf("failed to find category with key %q: %v", category.Key, err)
		}

		if _, err := client.GetCategoryValue(ctx, category.Key, category.Value); err != nil {
			return fmt.Errorf("failed to find category value %q for category %q: %v", category.Value, category.Key, err)
		}
	}

	return nil
}

func (v *Validator) validateGPUConfig(gpu anywherev1.NutanixGPUIdentifier) error {
	if gpu.Type == "" {
		return fmt.Errorf("missing GPU type")
	}

	if gpu.Type != anywherev1.NutanixGPUIdentifierDeviceID && gpu.Type != anywherev1.NutanixGPUIdentifierName {
		return fmt.Errorf("invalid GPU identifier type: %s; valid types are: %q and %q", gpu.Type, anywherev1.NutanixGPUIdentifierDeviceID, anywherev1.NutanixGPUIdentifierName)
	}

	if gpu.Type == anywherev1.NutanixGPUIdentifierDeviceID {
		if gpu.DeviceID == nil {
			return fmt.Errorf("missing GPU device ID")
		}
	} else {
		if gpu.Name == "" {
			return fmt.Errorf("missing GPU name")
		}
	}

	return nil
}

func (v *Validator) validateBootType(bootType anywherev1.NutanixBootType) error {
	if bootType != "legacy" && bootType != "uefi" {
		return fmt.Errorf("invalid boot type: %s; valid types are: %q and %q", bootType, anywherev1.NutanixBootTypeLegacy, anywherev1.NutanixBootTypeUEFI)
	}
	return nil
}

func getRequestedGPUsForAllMachines(machineCount int, requestedGpus []anywherev1.NutanixGPUIdentifier) []anywherev1.NutanixGPUIdentifier {
	allMachinesRequestedGPUs := make([]anywherev1.NutanixGPUIdentifier, 0)
	for i := 0; i < machineCount; i++ {
		allMachinesRequestedGPUs = append(allMachinesRequestedGPUs, requestedGpus...)
	}
	return allMachinesRequestedGPUs
}

func (v *Validator) tryAssignGPUsToMachineConfig(machineCount int, requestedGpus []anywherev1.NutanixGPUIdentifier, clusterGpuList []v3.GPU, cluster anywherev1.NutanixResourceIdentifier) ([]v3.GPU, error) {
	allMachinesRequestedGPUs := getRequestedGPUsForAllMachines(machineCount, requestedGpus)

	for _, requestedGpu := range allMachinesRequestedGPUs {
		found := -1
		for index, gpu := range clusterGpuList {
			if isRequestedGPUAssignable(gpu, requestedGpu) {
				found = index
				break
			}
		}

		if found == -1 {
			return nil, errorGPUNotFound(requestedGpu, cluster)
		}

		clusterGpuList = append(clusterGpuList[:found], clusterGpuList[found+1:]...)
	}

	return clusterGpuList, nil
}

func (v *Validator) isGPURequested(configs map[string]*anywherev1.NutanixMachineConfig) bool {
	for _, machineConfig := range configs {
		if machineConfig.Spec.GPUs != nil {
			return true
		}
	}

	return false
}

func (v *Validator) initAvailableGPUsMap(hosts []*v3.HostResponse) map[string][]v3.GPU {
	availableGpu := make(map[string][]v3.GPU)
	for _, host := range hosts {
		if host.Status != nil &&
			host.Status.Resources != nil &&
			host.Status.ClusterReference != nil &&
			host.Status.ClusterReference.UUID != "" {
			clusterUUID := host.Status.ClusterReference.UUID

			availableGpu[clusterUUID] = make([]v3.GPU, 0)
		}
	}
	return availableGpu
}

func (v *Validator) getAvailableGPUs(hosts []*v3.HostResponse) (map[string][]v3.GPU, error) {
	availableGpu := v.initAvailableGPUsMap(hosts)

	for _, host := range hosts {
		if host.Status != nil &&
			host.Status.Resources != nil &&
			host.Status.Resources.GPUList != nil &&
			host.Status.ClusterReference != nil &&
			host.Status.ClusterReference.UUID != "" {
			clusterUUID := host.Status.ClusterReference.UUID

			for _, gpu := range host.Status.Resources.GPUList {
				availableGpu[clusterUUID] = append(availableGpu[clusterUUID], *gpu)
			}
		}
	}

	return availableGpu, nil
}

func (v *Validator) tryAssignGPUsToAllMachineConfigs(ctx context.Context, v3Client Client, cluster *cluster.Spec, availableGpu map[string][]v3.GPU) error {
	configs := cluster.NutanixMachineConfigs
	machineCount := v.getMachineCountForAllMachineConfigs(cluster)

	for _, machineConfig := range configs {
		clusterUUID, err := getClusterUUID(ctx, v3Client, machineConfig.Spec.Cluster)
		if err != nil {
			return err
		}

		if machineConfig.Spec.GPUs != nil {
			if _, ok := machineCount[machineConfig.Name]; ok {
				availableGpu[clusterUUID], err = v.tryAssignGPUsToMachineConfig(machineCount[machineConfig.Name], machineConfig.Spec.GPUs, availableGpu[clusterUUID], machineConfig.Spec.Cluster)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (v *Validator) getMachineCountForAllMachineConfigs(clusterSpec *cluster.Spec) map[string]int {
	machineCountMap := make(map[string]int)
	cluster := clusterSpec.Cluster.Spec
	if cluster.ControlPlaneConfiguration.MachineGroupRef.Kind == constants.NutanixMachineConfigKind {
		machineCountMap[cluster.ControlPlaneConfiguration.MachineGroupRef.Name] = cluster.ControlPlaneConfiguration.Count
	}

	if cluster.ExternalEtcdConfiguration != nil &&
		cluster.ExternalEtcdConfiguration.MachineGroupRef.Kind == constants.NutanixMachineConfigKind {
		machineCountMap[cluster.ExternalEtcdConfiguration.MachineGroupRef.Name] = cluster.ExternalEtcdConfiguration.Count
	}

	for _, workerNodeGroupConfiguration := range cluster.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.MachineGroupRef.Kind == constants.NutanixMachineConfigKind &&
			workerNodeGroupConfiguration.Count != nil {
			machineCountMap[workerNodeGroupConfiguration.MachineGroupRef.Name] = *workerNodeGroupConfiguration.Count
		}
	}
	return machineCountMap
}

func (v *Validator) getGPUModeMapping(hosts []*v3.HostResponse) (map[int64]string, map[string]string, error) {
	gpuDeviceIDToMode := make(map[int64]string)
	gpuNameToMode := make(map[string]string)

	for _, host := range hosts {
		if host.Status != nil &&
			host.Status.Resources != nil &&
			host.Status.Resources.GPUList != nil {
			for _, gpu := range host.Status.Resources.GPUList {
				if gpu.DeviceID != nil {
					gpuDeviceIDToMode[*gpu.DeviceID] = gpu.Mode
				}
				if gpu.Name != "" {
					gpuNameToMode[gpu.Name] = gpu.Mode
				}
			}
		}
	}

	return gpuDeviceIDToMode, gpuNameToMode, nil
}

func (v *Validator) validateGPUModeNotMixed(hosts []*v3.HostResponse, cluster *cluster.Spec) error {
	configs := cluster.NutanixMachineConfigs

	gpuDeviceIDToMode, gpuNameToMode, err := v.getGPUModeMapping(hosts)
	if err != nil {
		return err
	}

	gpuMode := ""
	getGpuModeFunc := createGetGpuModeFunc(gpuDeviceIDToMode, gpuNameToMode)
	for _, machineConfig := range configs {
		if machineConfig.Spec.GPUs != nil {
			for _, gpu := range machineConfig.Spec.GPUs {
				if gpuMode == "" {
					gpuMode = getGpuModeFunc(gpu)
				} else {
					if gpuMode != getGpuModeFunc(gpu) {
						return fmt.Errorf("all GPUs in a machine config must be of the same mode, vGPU or passthrough")
					}
				}
			}
		}
	}

	return nil
}

func createGetGpuModeFunc(gpuDeviceIDToMode map[int64]string, gpuNameToMode map[string]string) func(gpu anywherev1.NutanixGPUIdentifier) string {
	return func(gpu anywherev1.NutanixGPUIdentifier) string {
		if gpu.Type == anywherev1.NutanixGPUIdentifierDeviceID {
			return gpuDeviceIDToMode[*gpu.DeviceID]
		}

		return gpuNameToMode[gpu.Name]
	}
}

func (v *Validator) validateFreeGPU(ctx context.Context, v3Client Client, cluster *cluster.Spec) error {
	if v.isGPURequested(cluster.NutanixMachineConfigs) {
		res, err := v3Client.ListAllHost(ctx)
		if err != nil || len(res.Entities) == 0 {
			return fmt.Errorf("no GPUs found: %v", err)
		}

		err = v.validateGPUModeNotMixed(res.Entities, cluster)
		if err != nil {
			return err
		}

		availableGpu, err := v.getAvailableGPUs(res.Entities)
		if err != nil {
			return err
		}

		if err = v.tryAssignGPUsToAllMachineConfigs(ctx, v3Client, cluster, availableGpu); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) validateUpgradeRolloutStrategy(clusterSpec *cluster.Spec) error {
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.UpgradeRolloutStrategy != nil {
		return fmt.Errorf("upgrade rollout strategy customization is not supported for nutanix provider")
	}
	for _, workerNodeGroupConfiguration := range clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.UpgradeRolloutStrategy != nil {
			return fmt.Errorf("upgrade rollout strategy customization is not supported for nutanix provider")
		}
	}
	return nil
}

func checkMachineConfigIsForWorker(config *anywherev1.NutanixMachineConfig, cluster *anywherev1.Cluster) error {
	if config.Name == cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name {
		return fmt.Errorf("GPUs are not supported for control plane machine")
	}

	if cluster.Spec.ExternalEtcdConfiguration != nil && config.Name == cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name {
		return fmt.Errorf("GPUs are not supported for external etcd machine")
	}

	for _, workerNodeGroupConfiguration := range cluster.Spec.WorkerNodeGroupConfigurations {
		if config.Name == workerNodeGroupConfiguration.MachineGroupRef.Name {
			return nil
		}
	}

	return fmt.Errorf("machine config %s is not associated with any worker node group", config.Name)
}

// findSubnetUUIDByName retrieves the subnet uuid by the given subnet name.
func findSubnetUUIDByName(ctx context.Context, v3Client Client, clusterUUID, subnetName string) (*string, error) {
	res, err := v3Client.ListAllSubnet(ctx, "", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list subnets: %v", err)
	}

	subnets := make([]*v3.SubnetIntentResponse, 0)
	for _, subnet := range res.Entities {
		if subnet.Spec.ClusterReference != nil && *subnet.Spec.ClusterReference.UUID != "" {
			if *subnet.Spec.ClusterReference.UUID == clusterUUID && subnet.Spec.Name != nil && *subnet.Spec.Name == subnetName {
				subnets = append(subnets, subnet)
			}
		}
	}

	if len(subnets) == 0 {
		return nil, fmt.Errorf("failed to find subnet by name %q: %v", subnetName, err)
	}

	if len(subnets) > 1 {
		return nil, fmt.Errorf("found more than one (%v) subnet with name %q and cluster uuid %v", len(subnets), subnetName, clusterUUID)
	}

	return subnets[0].Metadata.UUID, nil
}

// getWorkerMachineGroups retrieves the worker machine group names from the cluster spec.
func getWorkerMachineGroups(spec *cluster.Spec) map[string]anywherev1.WorkerNodeGroupConfiguration {
	result := make(map[string]anywherev1.WorkerNodeGroupConfiguration)

	for _, workerNodeGroupConf := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		result[workerNodeGroupConf.MachineGroupRef.Name] = workerNodeGroupConf
	}

	return result
}

// getClusterUUID retrieves the cluster uuid by the given cluster identifier.
func getClusterUUID(ctx context.Context, v3Client Client, cluster anywherev1.NutanixResourceIdentifier) (string, error) {
	var clusterUUID string
	var err error
	if cluster.Type == anywherev1.NutanixIdentifierUUID {
		if cluster.UUID == nil || *cluster.UUID == "" {
			return "", fmt.Errorf("missing cluster uuid")
		}
		clusterUUID = *cluster.UUID
	}

	if cluster.Type == anywherev1.NutanixIdentifierName {
		clusterName := *cluster.Name
		var uuid *string
		if uuid, err = findClusterUUIDByName(ctx, v3Client, clusterName); err != nil {
			return "", fmt.Errorf("failed to find cluster with name %q: %v", clusterName, err)
		}
		clusterUUID = *uuid
	}
	return clusterUUID, nil
}

// findClusterUUIDByName retrieves the cluster uuid by the given cluster name.
func findClusterUUIDByName(ctx context.Context, v3Client Client, clusterName string) (*string, error) {
	res, err := v3Client.ListAllCluster(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list clusters: %v", err)
	}

	entities := make([]*v3.ClusterIntentResponse, 0)
	for _, entity := range res.Entities {
		if entity.Status != nil && entity.Status.Resources != nil && entity.Status.Resources.Config != nil {
			serviceList := entity.Status.Resources.Config.ServiceList
			isPrismCentral := false
			for _, svc := range serviceList {
				// Prism Central is also internally a cluster, but we filter that out here as we only care about prism element clusters
				if svc != nil && strings.ToUpper(*svc) == "PRISM_CENTRAL" {
					isPrismCentral = true
				}
			}
			if !isPrismCentral && *entity.Spec.Name == clusterName {
				entities = append(entities, entity)
			}
		}
	}

	if len(entities) == 0 {
		return nil, fmt.Errorf("failed to find cluster by name %q: %v", clusterName, err)
	}

	if len(entities) > 1 {
		return nil, fmt.Errorf("found more than one (%v) cluster with name %q", len(entities), clusterName)
	}

	return entities[0].Metadata.UUID, nil
}

// findImageUUIDByName retrieves the image uuid by the given image name.
func findImageUUIDByName(ctx context.Context, v3Client Client, imageName string) (*string, error) {
	res, err := v3Client.ListAllImage(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list images: %v", err)
	}

	images := make([]*v3.ImageIntentResponse, 0)
	for _, image := range res.Entities {
		if image.Spec.Name != nil && *image.Spec.Name == imageName {
			images = append(images, image)
		}
	}

	if len(images) == 0 {
		return nil, fmt.Errorf("failed to find image by name %q: %v", imageName, err)
	}

	if len(images) > 1 {
		return nil, fmt.Errorf("found more than one (%v) image with name %q", len(images), imageName)
	}

	return images[0].Metadata.UUID, nil
}

// findProjectUUIDByName retrieves the project uuid by the given image name.
func findProjectUUIDByName(ctx context.Context, v3Client Client, projectName string) (*string, error) {
	res, err := v3Client.ListAllProject(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %v", err)
	}

	projects := make([]*v3.Project, 0)
	for _, project := range res.Entities {
		if project.Spec.Name == projectName {
			projects = append(projects, project)
		}
	}

	if len(projects) == 0 {
		return nil, fmt.Errorf("failed to find project by name %q: %v", projectName, err)
	}

	if len(projects) > 1 {
		return nil, fmt.Errorf("found more than one (%v) project with name %q", len(res.Entities), projectName)
	}

	return projects[0].Metadata.UUID, nil
}

func isRequestedGPUAssignable(gpu v3.GPU, requestedGpu anywherev1.NutanixGPUIdentifier) bool {
	if requestedGpu.Type == anywherev1.NutanixGPUIdentifierDeviceID {
		return (*gpu.DeviceID == *requestedGpu.DeviceID) && gpu.Assignable
	}

	return (gpu.Name == requestedGpu.Name) && gpu.Assignable
}

func errorGPUNotFound(gpu anywherev1.NutanixGPUIdentifier, cluster anywherev1.NutanixResourceIdentifier) error {
	clusterAddonString := ""
	if cluster.Type == anywherev1.NutanixIdentifierUUID {
		if cluster.UUID != nil {
			clusterAddonString = fmt.Sprintf("on cluster with UUID %s", *cluster.UUID)
		}
	} else {
		if cluster.Name != nil {
			clusterAddonString = fmt.Sprintf("on cluster with name %s", *cluster.Name)
		}
	}

	if gpu.Type == anywherev1.NutanixGPUIdentifierDeviceID {
		return fmt.Errorf("GPU with device ID %d not found %s", *gpu.DeviceID, clusterAddonString)
	}

	return fmt.Errorf("GPU with name %s not found %s", gpu.Name, clusterAddonString)
}
