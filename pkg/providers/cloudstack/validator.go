package cloudstack

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

type Validator struct {
	cmk       ProviderCmkClient
	machineConfigs         map[string]*anywherev1.CloudStackMachineConfig
	netClient networkutils.NetClient
}

func NewValidator(cmk ProviderCmkClient, machineConfigs map[string]*anywherev1.CloudStackMachineConfig, netClient networkutils.NetClient) *Validator {
	return &Validator{
		cmk:       cmk,
		machineConfigs: machineConfigs,
		netClient: netClient,
	}
}

func (v *Validator) validateCloudStackAccess(ctx context.Context) error {
	if err := v.cmk.ValidateCloudStackConnection(ctx); err != nil {
		return fmt.Errorf("failed validating connection to vCenter: %v", err)
	}
	logger.MarkPass("Connected to server")

	return nil
}

func (v *Validator) ValidateCloudStackDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.CloudStackDatacenterConfig) error {
	if err := v.cmk.ValidateCloudStackConnection(ctx); err != nil {
		return fmt.Errorf("failed validating connection to cloudstack: %v", err)
	}

	if err := v.cmk.ValidateZonePresent(ctx, datacenterConfig.Spec.Zone); err != nil {
		return err
	}
	logger.MarkPass("Datacenter validated")

	return nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, clusterSpec *Spec) error {
	var etcdMachineConfig *anywherev1.CloudStackMachineConfig

	if len(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if len(clusterSpec.datacenterConfig.Spec.Domain) <= 0 {
		return errors.New("CloudStackDatacenterConfig domain is not set or is empty")
	}
	if len(clusterSpec.datacenterConfig.Spec.Zone.Value) <= 0 {
		return errors.New("CloudStackDatacenterConfig zone is not set or is empty")
	}
	if len(clusterSpec.datacenterConfig.Spec.Network.Value) <= 0 {
		return errors.New("CloudStackDatacenterConfig network is not set or is empty")
	}
	if clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}
	controlPlaneMachineConfig := clusterSpec.controlPlaneMachineConfig()

	if clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig, ok := v.machineConfigs[clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name]
	if !ok {
		return fmt.Errorf("cannot find CloudStackMachineConfig %v for worker nodes", clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
	}

	if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for etcd machines")
		}
		etcdMachineConfig = clusterSpec.etcdMachineConfig()
		if len(etcdMachineConfig.Spec.Users) <= 0 {
			etcdMachineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
		}
		if len(etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
			etcdMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
		}
	}

	if len(controlPlaneMachineConfig.Spec.Users) <= 0 {
		controlPlaneMachineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
	}
	if len(workerNodeGroupMachineConfig.Spec.Users) <= 0 {
		workerNodeGroupMachineConfig.Spec.Users = []anywherev1.UserConfiguration{{}}
	}
	if len(controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		controlPlaneMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}
	if len(workerNodeGroupMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
		workerNodeGroupMachineConfig.Spec.Users[0].SshAuthorizedKeys = []string{""}
	}

	err := v.validateControlPlaneHost(&clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
	if err != nil {
		return err
	}

	for _, machineConfig := range v.machineConfigs {
		if machineConfig.Namespace != clusterSpec.Namespace {
			return errors.New("CloudStackMachineConfig and Cluster objects must have the same namespace specified")
		}
	}
	if clusterSpec.datacenterConfig.Namespace != clusterSpec.Namespace {
		return errors.New("CloudStackDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	if controlPlaneMachineConfig.Spec.Template.Value == "" {
		return fmt.Errorf("control plane CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
	}

	if err = v.validateMachineConfig(ctx, clusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane machine config validation failed.")
		return err
	}
	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		if workerNodeGroupMachineConfig.Spec.Template.Value == "" {
			return fmt.Errorf("worker CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
		}
		if err = v.validateMachineConfig(ctx, clusterSpec, workerNodeGroupMachineConfig); err != nil {
			logger.V(1).Info("Workload machine config validation failed.")
			return err
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return errors.New("control plane and worker nodes must have the same template specified")
		}
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if etcdMachineConfig.Spec.Template.Value == "" {
			return fmt.Errorf("etcd CloudStackMachineConfig template is not set. Default template is not supported in CloudStack, please provide a template name")
		}
		if err = v.validateMachineConfig(ctx, clusterSpec, etcdMachineConfig); err != nil {
			logger.V(1).Info("Etcd machine config validation failed.")
			return err
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	return nil
}

func (v *Validator) validateMachineConfig(ctx context.Context, clusterSpec *Spec, machineConfig *anywherev1.CloudStackMachineConfig) error {
	domain := clusterSpec.datacenterConfig.Spec.Domain
	zone := clusterSpec.datacenterConfig.Spec.Zone
	account := clusterSpec.datacenterConfig.Spec.Account
	err := v.cmk.ValidateTemplatePresent(ctx, domain, zone, account, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}
	if err := v.cmk.ValidateServiceOfferingPresent(ctx, domain, zone, account, machineConfig.Spec.ComputeOffering); err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	return nil
}

func (v *Validator) validateControlPlaneIp(ip string) error {
	// TODO: check if controlPlaneEndpointIp is valid
	return nil
}


func (v *Validator) validateControlPlaneHost(pHost *string) error {
	_, port, err := net.SplitHostPort(*pHost)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			port = controlEndpointDefaultPort
			*pHost = fmt.Sprintf("%s:%s", *pHost, port)
		} else {
			return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s (%s)", *pHost, err.Error())
		}
	}
	_, err = strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host has an invalid port: %s (%s)", *pHost, err.Error())
	}
	return nil
}

func (v *Validator) validateTemplatePresence(ctx context.Context, deploymentConfig *anywherev1.CloudStackDatacenterConfig, template anywherev1.CloudStackResourceRef) error {
	domain := deploymentConfig.Spec.Domain
	zone := deploymentConfig.Spec.Zone
	account := deploymentConfig.Spec.Account
	err := v.cmk.ValidateTemplatePresent(ctx, domain, zone, account, template)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	return nil
}

func (v *Validator) validateControlPlaneIpUniqueness(spec *Spec) error {
	// TODO: Implement in CloudStack
	return nil
}
