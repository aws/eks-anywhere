package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
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
			return fmt.Errorf("failed to generate eks-a cluster config: %v", err) // need to have better error handling here in own func
		}
		return nil
	},
}

func preRunGenerateClusterConfig(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("failed initializing flags: %v", err)
		}
	})
}

func init() {
	generateCmd.AddCommand(generateClusterConfigCmd)
	generateClusterConfigCmd.Flags().StringP("provider", "p", "", "Provider to use (vsphere or docker)")
	err := generateClusterConfigCmd.MarkFlagRequired("provider")
	if err != nil {
		log.Fatalf("failed marking flag as required: %v", err)
	}
}

func generateClusterConfig(clusterName string) error {
	var resources [][]byte
	var datacenterYaml []byte
	var machineGroupYaml [][]byte
	var clusterConfigOpts []v1alpha1.ClusterGenerateOpt
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
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml
	case constants.VSphereProviderName:
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		datacenterConfig := v1alpha1.NewVSphereDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(2),
			v1alpha1.ExternalETCDConfigCount(3),
			v1alpha1.WorkerNodeConfigCount(2),
			v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		datacenterYaml = dcyaml
		// need to default control plane config name to something different from the cluster name based on assumption
		// in controller code
		cpMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName))
		workerMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(clusterName)
		etcdMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(providers.GetEtcdNodeName(clusterName))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			v1alpha1.WithEtcdMachineGroupRef(etcdMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	case constants.SnowProviderName:
		if !features.IsActive(features.SnowProvider()) {
			return fmt.Errorf("the snow infrastructure provider is still under development")
		}
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
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
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
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml)
	case constants.CloudStackProviderName:
		if !features.IsActive(features.CloudStackProvider()) {
			return fmt.Errorf("the cloudstack infrastructure provider is still under development")
		}
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		datacenterConfig := v1alpha1.NewCloudStackDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.ControlPlaneConfigCount(2),
			v1alpha1.ExternalETCDConfigCount(3),
			v1alpha1.WorkerNodeConfigCount(2),
		)
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
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
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("failed to generate cluster yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	case constants.TinkerbellProviderName:
		if features.IsActive(features.TinkerbellProvider()) {
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
			datacenterConfig := v1alpha1.NewTinkerbellDatacenterConfigGenerate(clusterName)
			clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
			clusterConfigOpts = append(clusterConfigOpts,
				v1alpha1.ControlPlaneConfigCount(1),
				v1alpha1.WorkerNodeConfigCount(1),
				v1alpha1.WorkerNodeConfigName(constants.DefaultWorkerNodeGroupName),
			)
			dcyaml, err := yaml.Marshal(datacenterConfig)
			if err != nil {
				return fmt.Errorf("failed to generate cluster yaml: %v", err)
			}
			datacenterYaml = dcyaml

			cpMachineConfig, err := v1alpha1.NewTinkerbellMachineConfigGenerate(providers.GetControlPlaneNodeName(clusterName))
			if err != nil {
				return fmt.Errorf("failed to generate cluster yaml: %v", err)
			}
			workerMachineConfig, err := v1alpha1.NewTinkerbellMachineConfigGenerate(clusterName)
			if err != nil {
				return fmt.Errorf("failed to generate cluster yaml: %v", err)
			}
			clusterConfigOpts = append(clusterConfigOpts,
				v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
				v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			)
			cpMcYaml, err := yaml.Marshal(cpMachineConfig)
			if err != nil {
				return fmt.Errorf("failed to generate cluster yaml: %v", err)
			}
			workerMcYaml, err := yaml.Marshal(workerMachineConfig)
			if err != nil {
				return fmt.Errorf("failed to generate cluster yaml: %v", err)
			}
			machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml)
		} else {
			return fmt.Errorf("the tinkerbell infrastructure provider under still under development")
		}
	default:
		return fmt.Errorf("not a valid provider")
	}
	config := v1alpha1.NewClusterGenerate(clusterName, clusterConfigOpts...)

	configMarshal, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to generate cluster yaml: %v", err)
	}
	clusterYaml, err := api.CleanupPathsFromYaml(configMarshal, removeFromDefaultConfig)
	if err != nil {
		return fmt.Errorf("failed to clean up paths from yaml: %v", err)
	}
	resources = append(resources, clusterYaml, datacenterYaml)
	if len(machineGroupYaml) > 0 {
		resources = append(resources, machineGroupYaml...)
	}
	fmt.Println(string(templater.AppendYamlResources(resources...)))
	return nil
}
