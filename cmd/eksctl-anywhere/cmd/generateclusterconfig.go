package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/validations"
)

var removeFromDefaultConfig = []string{"spec.clusterNetwork.dns"}

var generateClusterConfigCmd = &cobra.Command{
	Use:    "clusterconfig <cluster-name> (max 80 chars)",
	Short:  "Generate cluster config",
	Long:   "This command is used to generate a cluster config yaml for the create cluster command",
	PreRun: preRunGenerateClusterConfig,
	RunE: func(cmd *cobra.Command, args []string) error {
		clusterName, err := validations.ValidateClusterNameArg(args)
		if err != nil {
			return err
		}
		err = generateClusterConfig(clusterName)
		if err != nil {
			return fmt.Errorf("generating eks-a cluster config: %v", err) // need to have better error handling here in own func
		}
		return nil
	},
}

func preRunGenerateClusterConfig(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("initializing flags: %v", err)
		}
	})
}

func init() {
	generateCmd.AddCommand(generateClusterConfigCmd)
	generateClusterConfigCmd.Flags().StringP("provider", "p", "", fmt.Sprintf("Provider to use (%s)", strings.Join(constants.SupportedProviders, " or ")))
	generateClusterConfigCmd.Flags().StringP("paramsFile", "m", "", "parameters file (vsphere or tinkerbell)")
	err := generateClusterConfigCmd.MarkFlagRequired("provider")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}
}

func generateClusterConfig(clusterName string) error {
	var resources [][]byte
	var datacenterYaml []byte
	var machineGroupYaml [][]byte
	var clusterConfigOpts []v1alpha1.ClusterGenerateOpt
	var kubernetesVersion string
	var tinkerbellTemplateConfigTemplate string
	var podsCidrBlocks []string
	var servicesCidrBlocks []string
	var paramsData []byte
	var err error

	// use cluster name as the default management cluster name.
	managementClusterName := clusterName

	if viper.IsSet("paramsFile") {
		paramsFile := viper.GetString("paramsFile")
		paramsData, err = os.ReadFile(paramsFile)

		switch strings.ToLower(viper.GetString("provider")) {
		case constants.VSphereProviderName:
			if err != nil {
				return fmt.Errorf("reading paramsFile: %v\nSample paramsFile has content:\n%s", err, GetDefaultVSphereParamsTemplate())
			}
		case constants.TinkerbellProviderName:
			if err != nil {
				return fmt.Errorf("reading paramsFile: %v\nSample paramsFile has content:\n%s", err, GetDefaultTinkerbellParamsTemplate())
			}
		default:
			return fmt.Errorf("parameter file is only supported for vsphere and tinkerbell")
		}
	}

	switch strings.ToLower(viper.GetString("provider")) {
	case constants.DockerProviderName:
		datacenterConfig := v1alpha1.NewDockerDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(1),
			v1alpha1.ExternalETCDConfigCount(1),
			v1alpha1.WorkerNodeConfigCount(1),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml
	case constants.VSphereProviderName:
		var vSphereParams v1alpha1.VSphereClusterConfigParams
		err = yaml.Unmarshal(paramsData, &vSphereParams)
		if err != nil {
			return fmt.Errorf("unmarshal vSphereParams: %v", err)
		}

		if vSphereParams.ManagementClusterName != "" {
			// override the management cluster name with that from parameter file.
			managementClusterName = vSphereParams.ManagementClusterName
		}

		// set podsCidrBlocks and servicesCidrBlocks to the values from parameter file.
		podsCidrBlocks = vSphereParams.PodsCidrBlocks
		servicesCidrBlocks = vSphereParams.ServicesCidrBlocks

		if vSphereParams.CPEndpointHost != "" {
			// add control plane endpoint config with host from parameter file.
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpointHost(vSphereParams.CPEndpointHost))
		} else {
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		}

		// create datacenter config with values from parameter file
		datacenterConfig := v1alpha1.NewVSphereDatacenterConfigGenerate(clusterName, vSphereParams.Datacenter, vSphereParams.Network, vSphereParams.Server, vSphereParams.Thumbprint, vSphereParams.Insecure)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		// default counts of CP nodes, Etcd nodes and worker nodes.
		cpCount := 2
		etcdCount := 3
		workerCount := 2

		if vSphereParams.CPCount != 0 {
			// override counts of CP nodes with value from parameter file.
			cpCount = vSphereParams.CPCount
		}

		if vSphereParams.EtcdCount != 0 {
			// override counts of Etcd nodes with value from parameter file.
			etcdCount = vSphereParams.EtcdCount
		}

		if vSphereParams.WorkerCount != 0 {
			// override counts of Worker nodes with value from parameter file.
			workerCount = vSphereParams.WorkerCount
		}
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(cpCount),
			v1alpha1.ExternalETCDConfigCount(etcdCount),
			v1alpha1.WorkerNodeConfigCount(workerCount),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml
		var sshAuthorizedKey string
		if vSphereParams.SSHAuthorizedKeyFile != "" {
			b, err := os.ReadFile(vSphereParams.SSHAuthorizedKeyFile)
			if err != nil {
				return fmt.Errorf("open sshAuthorizedKeyFile file: %v", err)
			}
			sshAuthorizedKey = string(b)
		}

		kubernetesVersion = vSphereParams.KubernetesVersion
		// need to default control plane config name to something different from the cluster name based on assumption
		// in controller code
		cpMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName), vSphereParams.Datastore, vSphereParams.Folder, vSphereParams.ResourcePool, vSphereParams.Template, sshAuthorizedKey, vSphereParams.OSFamily, vSphereParams.CPDiskGiB, vSphereParams.CPNumCPUs, vSphereParams.CPMemoryMiB)
		workerMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(clusterName, vSphereParams.Datastore, vSphereParams.Folder, vSphereParams.ResourcePool, vSphereParams.Template, sshAuthorizedKey, vSphereParams.OSFamily, vSphereParams.WorkerDiskGiB, vSphereParams.WorkerNumCPUs, vSphereParams.WorkerMemoryMiB)
		etcdMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(providers.GetEtcdNodeName(clusterName), vSphereParams.Datastore, vSphereParams.Folder, vSphereParams.ResourcePool, vSphereParams.Template, sshAuthorizedKey, vSphereParams.OSFamily, vSphereParams.EtcdDiskGiB, vSphereParams.EtcdNumCPUs, vSphereParams.EtcdMemoryMiB)
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			v1alpha1.WithEtcdMachineGroupRef(etcdMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	case constants.SnowProviderName:
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		datacenterConfig := v1alpha1.NewSnowDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(3),
			v1alpha1.WorkerNodeConfigCount(3),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml

		cpMachineConfig := v1alpha1.NewSnowMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName))
		workerMachineConfig := v1alpha1.NewSnowMachineConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml)
	case constants.CloudStackProviderName:
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		datacenterConfig := v1alpha1.NewCloudStackDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(2),
			v1alpha1.ExternalETCDConfigCount(3),
			v1alpha1.WorkerNodeConfigCount(2),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml
		// need to default control plane config name to something different from the cluster name based on assumption
		// in controller code
		cpMachineConfig := v1alpha1.NewCloudStackMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName))
		workerMachineConfig := v1alpha1.NewCloudStackMachineConfigGenerate(clusterName)
		etcdMachineConfig := v1alpha1.NewCloudStackMachineConfigGenerate(providers.GetEtcdNodeName(clusterName))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			v1alpha1.WithEtcdMachineGroupRef(etcdMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	case constants.TinkerbellProviderName:
		var tinkerbellParams v1alpha1.TinkerbellClusterConfigParams
		err = yaml.Unmarshal(paramsData, &tinkerbellParams)
		if err != nil {
			return fmt.Errorf("unmarshal tinkerbellParams: %v", err)
		}

		if tinkerbellParams.ManagementClusterName != "" {
			// override the management cluster name with that from parameter file.
			managementClusterName = tinkerbellParams.ManagementClusterName
		}

		// set podsCidrBlocks and servicesCidrBlocks to the values from parameter file.
		podsCidrBlocks = tinkerbellParams.PodsCidrBlocks
		servicesCidrBlocks = tinkerbellParams.ServicesCidrBlocks

		if tinkerbellParams.CPEndpointHost != "" {
			// add control plane endpoint config with host from parameter file.
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpointHost(tinkerbellParams.CPEndpointHost))
		} else {
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		}

		kubernetesVersion = tinkerbellParams.KubernetesVersion

		adminIP := tinkerbellParams.AdminIP
		tinkerbellIP := tinkerbellParams.TinkerbellIP
		osImageURL := tinkerbellParams.OSImageURL

		// create datacenter config with values from parameter file
		datacenterConfig := v1alpha1.NewTinkerbellDatacenterConfigGenerate(clusterName, tinkerbellIP, osImageURL)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		// default counts of CP nodes, Etcd nodes and worker nodes.
		cpCount := 1
		workerCount := 1
		if tinkerbellParams.HardwareCSV != "" {
			// parse hardware.csv file to get counts of CP/worker nodes
			f, err := os.Open(tinkerbellParams.HardwareCSV)
			if err != nil {
				return fmt.Errorf("open hardware file: %v", err)
			}
			defer f.Close()
			csvReader := csv.NewReader(f)
			data, err := csvReader.ReadAll()
			if err != nil {
				return fmt.Errorf("read hardware file: %v", err)
			}
			macIndex := -1
			ipIndex := -1
			labelsIndex := -1
			cpCount = 0
			workerCount = 0
			for i, line := range data {
				if i == 0 {
					// from the header (first line), find the index of
					// MAC, IP, labels.
					for j, field := range line {
						if strings.EqualFold(field, "mac") {
							macIndex = j
						} else if strings.EqualFold(field, "ip_address") {
							ipIndex = j
						} else if strings.EqualFold(field, "labels") {
							labelsIndex = j
						}
					}
					if macIndex == -1 {
						return fmt.Errorf("no mac header found in hardware file")
					}
					if ipIndex == -1 {
						return fmt.Errorf("no ip header found in hardware file")
					}
					if labelsIndex == -1 {
						return fmt.Errorf("no labels header found in hardware file")
					}
				} else {
					// for rest lines, increase counts of CP nodes and worker nodes.
					if strings.ToLower(line[labelsIndex]) == "type=cp" {
						cpCount = cpCount + 1
					} else {
						workerCount = workerCount + 1
					}
				}
			}
		}

		if tinkerbellParams.CPCount != 0 {
			// override counts of CP nodes with value from parameter file.
			cpCount = tinkerbellParams.CPCount
		}

		if tinkerbellParams.WorkerCount != 0 {
			// override counts of Worker nodes with value from parameter file.
			workerCount = tinkerbellParams.WorkerCount
		}

		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(cpCount),
		)
		if workerCount > 0 {
			// only generate worker cluster when worker count > 0.
			clusterConfigOpts = append(clusterConfigOpts,
				v1alpha1.WorkerNodeConfigCount(workerCount),
				v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
			)
		}
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml

		var sshAuthorizedKey string
		if tinkerbellParams.SSHAuthorizedKeyFile != "" {
			b, err := os.ReadFile(tinkerbellParams.SSHAuthorizedKeyFile)
			if err != nil {
				return fmt.Errorf("open sshAuthorizedKeyFile file: %v", err)
			}
			sshAuthorizedKey = string(b)
		}

		cpMachineConfig := v1alpha1.NewTinkerbellMachineConfigGenerate(clusterName, providers.GetControlPlaneNodeName(clusterName), "cp", sshAuthorizedKey, tinkerbellParams.OSFamily)
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("generating cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml)

		if workerCount > 0 {
			workerMachineConfig := v1alpha1.NewTinkerbellMachineConfigGenerate(clusterName, clusterName, "worker", sshAuthorizedKey, tinkerbellParams.OSFamily)
			// only generate worker machine group reference when worker count > 0.
			clusterConfigOpts = append(clusterConfigOpts,
				v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			)
			// only generate worker machine config YAML when worker count > 0.
			workerMcYaml, err := yaml.Marshal(workerMachineConfig)
			if err != nil {
				return fmt.Errorf("generating cluster yaml: %v", err)
			}
			machineGroupYaml = append(machineGroupYaml, workerMcYaml)
		}

		if viper.IsSet("paramsFile") {
			if tinkerbellParams.TinkerbellTemplateConfigTemplateFile != "" {
				b, err := os.ReadFile(tinkerbellParams.TinkerbellTemplateConfigTemplateFile)
				if err != nil {
					if tinkerbellParams.OSFamily == v1alpha1.Ubuntu {
						return fmt.Errorf("open tinkerbellTemplateConfigTemplateFile file: %v\nSample TinkerbellTemplateConfigTemplateFile has content:%s", err, GetDefaultTinkerbellTemplateConfigTemplateUbuntu())
					} else if tinkerbellParams.OSFamily == v1alpha1.Bottlerocket {
						return fmt.Errorf("open tinkerbellTemplateConfigTemplateFile file: %v\nSample TinkerbellTemplateConfigTemplateFile has content:%s", err, GetDefaultTinkerbellTemplateConfigTemplateBottlerocket())
					}
					return fmt.Errorf("open tinkerbellTemplateConfigTemplateFile file: %v", err)
				}
				tinkerbellTemplateConfigTemplate = string(b)
			} else if tinkerbellParams.OSFamily == v1alpha1.Ubuntu {
				tinkerbellTemplateConfigTemplate = GetDefaultTinkerbellTemplateConfigTemplateUbuntu()
			} else if tinkerbellParams.OSFamily == v1alpha1.Bottlerocket {
				tinkerbellTemplateConfigTemplate = GetDefaultTinkerbellTemplateConfigTemplateBottlerocket()
			}

			tinkerbellTemplateConfigTemplate = strings.Replace(tinkerbellTemplateConfigTemplate, "$$NAME", clusterName, -1)
			tinkerbellTemplateConfigTemplate = strings.Replace(tinkerbellTemplateConfigTemplate, "$$IMG_URL", osImageURL, -1)
			tinkerbellTemplateConfigTemplate = strings.Replace(tinkerbellTemplateConfigTemplate, "$$ADMIN_IP", adminIP, -1)
			tinkerbellTemplateConfigTemplate = strings.Replace(tinkerbellTemplateConfigTemplate, "$$TINKERBELL_IP", tinkerbellIP, -1)
		}
	case constants.NutanixProviderName:
		datacenterConfig := v1alpha1.NewNutanixDatacenterConfigGenerate(clusterName)
		dcYaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		datacenterYaml = dcYaml
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(2),
			v1alpha1.ExternalETCDConfigCount(3),
			v1alpha1.WorkerNodeConfigCount(2),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)

		cpMachineConfig := v1alpha1.NewNutanixMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName))
		etcdMachineConfig := v1alpha1.NewNutanixMachineConfigGenerate(providers.GetEtcdNodeName(clusterName))
		workerMachineConfig := v1alpha1.NewNutanixMachineConfigGenerate(clusterName)

		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithEtcdMachineGroupRef(etcdMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
		)

		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}

		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}

		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}

		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	default:
		return fmt.Errorf("not a valid provider")
	}

	config := v1alpha1.NewClusterGenerate(clusterName, managementClusterName, kubernetesVersion, podsCidrBlocks, servicesCidrBlocks, clusterConfigOpts...)

	configMarshal, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("generating cluster yaml: %v", err)
	}
	clusterYaml, err := api.CleanupPathsFromYaml(configMarshal, removeFromDefaultConfig)
	if err != nil {
		return fmt.Errorf("cleaning up paths from yaml: %v", err)
	}
	resources = append(resources, clusterYaml, datacenterYaml)
	if len(machineGroupYaml) > 0 {
		resources = append(resources, machineGroupYaml...)
	}

	fmt.Print(string(templater.AppendYamlResources(resources...)))

	if tinkerbellTemplateConfigTemplate != "" {
		fmt.Println(tinkerbellTemplateConfigTemplate)
	} else {
		fmt.Println("")
	}

	return nil
}
