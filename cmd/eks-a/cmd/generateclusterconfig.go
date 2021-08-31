package cmd

import (
	"fmt"
	"log"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/validations"
)

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
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func init() {
	generateCmd.AddCommand(generateClusterConfigCmd)
	generateClusterConfigCmd.Flags().StringP("provider", "p", "", "Provider to use")
	err := generateClusterConfigCmd.MarkFlagRequired("provider")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func generateClusterConfig(clusterName string) error {
	var resources [][]byte
	var datacenterYaml []byte
	var machineGroupYaml [][]byte
	var clusterConfigOpts []v1alpha1.ClusterGenerateOpt
	switch strings.ToLower(viper.GetString("provider")) {
	case docker.ProviderName:
		datacenterConfig := v1alpha1.NewDockerDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("error outputting yaml: %v", err)
		}
		datacenterYaml = dcyaml
	case vsphere.ProviderName:
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithClusterEndpoint())
		datacenterConfig := v1alpha1.NewVSphereDatacenterConfigGenerate(clusterName)
		clusterConfigOpts = append(clusterConfigOpts, v1alpha1.WithDatacenterRef(datacenterConfig))
		dcyaml, err := yaml.Marshal(datacenterConfig)
		if err != nil {
			return fmt.Errorf("error outputting yaml: %v", err)
		}
		datacenterYaml = dcyaml
		// need to default control plane config name to something different from the cluster name based on assumption
		// in controller code
		cpMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(clusterName + "-cp")
		workerMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(clusterName)
		etcdMachineConfig := v1alpha1.NewVSphereMachineConfigGenerate(fmt.Sprintf("%s-etcd", clusterName))
		clusterConfigOpts = append(clusterConfigOpts,
			v1alpha1.WithCPMachineGroupRef(cpMachineConfig),
			v1alpha1.WithWorkerMachineGroupRef(workerMachineConfig),
			v1alpha1.WithEtcdMachineGroupRef(etcdMachineConfig),
		)
		cpMcYaml, err := yaml.Marshal(cpMachineConfig)
		if err != nil {
			return fmt.Errorf("error outputting yaml: %v", err)
		}
		workerMcYaml, err := yaml.Marshal(workerMachineConfig)
		if err != nil {
			return fmt.Errorf("error outputting yaml: %v", err)
		}
		etcdMcYaml, err := yaml.Marshal(etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("error outputting yaml: %v", err)
		}
		machineGroupYaml = append(machineGroupYaml, cpMcYaml, workerMcYaml, etcdMcYaml)
	default:
		return fmt.Errorf("not a valid provider")
	}
	config := v1alpha1.NewClusterGenerate(clusterName, clusterConfigOpts...)

	clusterYaml, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("error outputting yaml: %v", err)
	}
	resources = append(resources, clusterYaml, datacenterYaml)
	if len(machineGroupYaml) > 0 {
		resources = append(resources, machineGroupYaml...)
	}
	fmt.Println(string(templater.AppendYamlResources(resources...)))
	return nil
}
