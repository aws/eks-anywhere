package vsphere

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

type OVFDeployOptions struct {
	Name             string           `json:"Name"`
	PowerOn          bool             `json:"PowerOn"`
	DiskProvisioning string           `json:"DiskProvisioning"`
	WaitForIP        bool             `json:"WaitForIP"`
	NetworkMappings  []NetworkMapping `json:"NetworkMapping"`
	Annotation       string           `json:"Annotation"`
	PropertyMapping  []OVFProperty    `json:"PropertyMapping"`
	InjectOvfEnv     bool             `json:"InjectOvfEnv"`
}

type OVFProperty struct {
	Key   string `json:"Key"`
	Value string `json:"Value"`
}

type NetworkMapping struct {
	Name    string `json:"Name"`
	Network string `json:"Network"`
}

func DeployTemplate(envMap map[string]string, library, templateName, vmName, deployFolder, datacenter, datastore, resourcePool string, opts OVFDeployOptions) error {
	context := context.Background()
	executableBuilder, close, err := executables.NewExecutableBuilder(context, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	defer close.CheckErr(context)
	tmpWriter, _ := filewriter.NewWriter(vmName)
	govc := executableBuilder.BuildGovcExecutable(tmpWriter, executables.WithGovcEnvMap(envMap))
	defer govc.Close(context)

	deployOptions, err := json.Marshal(opts)
	if err != nil {
		return fmt.Errorf("failed to marshall vm deployment options: %v", err)
	}

	// deploy template
	if err := govc.DeployTemplate(context, library, templateName, vmName, deployFolder, datacenter, datastore, opts.NetworkMappings[0].Network, resourcePool, deployOptions); err != nil {
		return fmt.Errorf("failed to deploy vm from library template: %v", err)
	}

	return nil
}

func TagVirtualMachine(envMap map[string]string, vmPath, tag string) error {
	context := context.Background()
	executableBuilder, close, err := executables.NewExecutableBuilder(context, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}

	defer close.CheckErr(context)
	tmpWriter, _ := filewriter.NewWriter(vmPath)
	govc := executableBuilder.BuildGovcExecutable(tmpWriter, executables.WithGovcEnvMap(envMap))
	defer govc.Close(context)

	if err := govc.AddTag(context, vmPath, tag); err != nil {
		return fmt.Errorf("failed to tag vm: %v", err)
	}
	return nil
}
