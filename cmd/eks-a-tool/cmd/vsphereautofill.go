package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/validations"
)

var autofillCmd = &cobra.Command{
	Use:    "autofill",
	Short:  "Autofill provider config",
	Long:   "Fills provider config with values set in environment variables",
	PreRun: preRunAutofill,
	RunE: func(cmd *cobra.Command, args []string) error {
		err := autofill(cmd.Context())
		if err != nil {
			log.Fatalf("Error filling the provider config: %v", err)
		}
		return nil
	},
}

func init() {
	vsphereCmd.AddCommand(autofillCmd)
	autofillCmd.Flags().StringP("filename", "f", "", "Cluster config yaml filepath")
	err := autofillCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking flag as required: %v", err)
	}
}

func preRunAutofill(cmd *cobra.Command, args []string) {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		err := viper.BindPFlag(flag.Name, flag)
		if err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
}

func autofill(ctx context.Context) error {
	clusterConfigFileName := viper.GetString("filename")
	clusterConfigFileExist := validations.FileExists(clusterConfigFileName)
	if !clusterConfigFileExist {
		return fmt.Errorf("the cluster config file %s does not exist", clusterConfigFileName)
	}
	clusterConfig, err := v1alpha1.GetAndValidateClusterConfig(clusterConfigFileName)
	if err != nil {
		return fmt.Errorf("unable to get cluster config from file: %v", err)
	}
	datacenterConfig, err := v1alpha1.GetVSphereDatacenterConfig(clusterConfigFileName)
	if err != nil {
		return fmt.Errorf("unable to get datacenter config from file: %v", err)
	}
	machineConfig, err := v1alpha1.GetVSphereMachineConfigs(clusterConfigFileName)
	if err != nil {
		return fmt.Errorf("unable to get machine config from file: %v", err)
	}
	controlPlaneMachineConfig := machineConfig[clusterConfig.Spec.ControlPlaneConfiguration.MachineGroupRef.Name]
	workerMachineConfig := machineConfig[clusterConfig.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	var updatedFields []string
	updateField := func(envName string, field *string) {
		if value, set := os.LookupEnv(envName); set && len(value) > 0 {
			*field = value
			updatedFields = append(updatedFields, envName)
		}
	}

	updateFieldInt := func(envName string, field *int) {
		if value, set := os.LookupEnv(envName); set && len(value) > 0 {
			val, _ := strconv.Atoi(value)
			*field = val
			updatedFields = append(updatedFields, envName)
		}
	}
	tlsInsecure := strconv.FormatBool(datacenterConfig.Spec.Insecure)
	updateField("CONTROL_PLANE_ENDPOINT_IP", &clusterConfig.Spec.ControlPlaneConfiguration.Endpoint.Host)
	updateField("DATACENTER", &datacenterConfig.Spec.Datacenter)
	updateField("NETWORK", &datacenterConfig.Spec.Network)
	updateField("SERVER", &datacenterConfig.Spec.Server)
	updateField("INSECURE", &tlsInsecure)
	updateField("THUMBPRINT", &datacenterConfig.Spec.Thumbprint)

	updateFieldInt("CONTROL_PLANE_COUNT", &clusterConfig.Spec.ControlPlaneConfiguration.Count)
	updateFieldInt("WORKER_NODE_COUNT", &clusterConfig.Spec.WorkerNodeGroupConfigurations[0].Count)

	updateField("SSH_AUTHORIZED_KEY", &controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys[0])
	updateField("SSH_USERNAME", &controlPlaneMachineConfig.Spec.Users[0].Name)
	updateField("TEMPLATE", &controlPlaneMachineConfig.Spec.Template)
	updateField("DATASTORE", &controlPlaneMachineConfig.Spec.Datastore)
	updateField("FOLDER", &controlPlaneMachineConfig.Spec.Folder)
	updateField("RESOURCE_POOL", &controlPlaneMachineConfig.Spec.ResourcePool)
	updateField("STORAGE_POLICY_NAME", &controlPlaneMachineConfig.Spec.StoragePolicyName)

	updateField("SSH_AUTHORIZED_KEY", &workerMachineConfig.Spec.Users[0].SshAuthorizedKeys[0])
	updateField("SSH_USERNAME", &workerMachineConfig.Spec.Users[0].Name)
	updateField("TEMPLATE", &workerMachineConfig.Spec.Template)
	updateField("DATASTORE", &workerMachineConfig.Spec.Datastore)
	updateField("FOLDER", &workerMachineConfig.Spec.Folder)
	updateField("RESOURCE_POOL", &workerMachineConfig.Spec.ResourcePool)
	updateField("STORAGE_POLICY_NAME", &workerMachineConfig.Spec.StoragePolicyName)

	clusterOutput, err := yaml.Marshal(clusterConfig)
	if err != nil {
		return fmt.Errorf("outputting yaml: %v", err)
	}
	datacenterOutput, err := yaml.Marshal(datacenterConfig)
	if err != nil {
		return fmt.Errorf("outputting yaml: %v", err)
	}
	controlPlaneMachineOutput, err := yaml.Marshal(controlPlaneMachineConfig)
	if err != nil {
		return fmt.Errorf("outputting yaml: %v", err)
	}
	workerMachineOutput, err := yaml.Marshal(workerMachineConfig)
	if err != nil {
		return fmt.Errorf("outputting yaml: %v", err)
	}
	result := strings.ReplaceAll(string(datacenterOutput), "  aws: {}\n", "")
	result = strings.ReplaceAll(result, "  vsphere: {}\n", "")
	result = string(clusterOutput) + "\n---\n" + result + "\n---\n" + string(controlPlaneMachineOutput) + "\n---\n" + string(workerMachineOutput)

	writer, err := filewriter.NewWriter(filepath.Dir(clusterConfig.Name))
	if err != nil {
		return err
	}
	_, err = writer.Write(filepath.Base(clusterConfig.Name), []byte(result))
	if err != nil {
		return fmt.Errorf("writing to file %s: %v", clusterConfig.Name, err)
	}
	fmt.Printf("The following fields were updated: %v\n", updatedFields)
	return nil
}
