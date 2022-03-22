package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/types"
)

type Validator struct {
	govc      ProviderGovcClient
	netClient networkutils.NetClient
}

func NewValidator(govc ProviderGovcClient, netClient networkutils.NetClient) *Validator {
	return &Validator{
		govc:      govc,
		netClient: netClient,
	}
}

func (v *Validator) validateVCenterAccess(ctx context.Context, server string) error {
	if err := v.govc.ValidateVCenterConnection(ctx, server); err != nil {
		return fmt.Errorf("failed validating connection to vCenter: %v", err)
	}
	logger.MarkPass("Connected to server")

	if err := v.govc.ValidateVCenterAuthentication(ctx); err != nil {
		return fmt.Errorf("failed validating credentials for vCenter: %v", err)
	}
	logger.MarkPass("Authenticated to vSphere")

	return nil
}

func (v *Validator) ValidateVCenterConfig(ctx context.Context, datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	if err := v.validateVCenterAccess(ctx, datacenterConfig.Spec.Server); err != nil {
		return err
	}

	if err := v.validateThumbprint(ctx, datacenterConfig); err != nil {
		return err
	}

	if err := v.validateDatacenter(ctx, datacenterConfig.Spec.Datacenter); err != nil {
		return err
	}
	logger.MarkPass("Datacenter validated")

	if err := v.validateNetwork(ctx, datacenterConfig.Spec.Network); err != nil {
		return err
	}
	logger.MarkPass("Network validated")

	return nil
}

// TODO: dry out machine configs validations
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, vsphereClusterSpec *Spec) error {
	var etcdMachineConfig *anywherev1.VSphereMachineConfig

	// TODO: move this to api Cluster validations
	if len(vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}

	if vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for control plane")
	}

	controlPlaneMachineConfig := vsphereClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for control plane", vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}
	if len(controlPlaneMachineConfig.Spec.Datastore) <= 0 {
		return errors.New("VSphereMachineConfig datastore for control plane is not set or is empty")
	}
	if len(controlPlaneMachineConfig.Spec.Folder) <= 0 {
		logger.Info("VSphereMachineConfig folder for control plane is not set or is empty. Will default to root vSphere folder.")
	}
	if len(controlPlaneMachineConfig.Spec.ResourcePool) <= 0 {
		return errors.New("VSphereMachineConfig VM resourcePool for control plane is not set or is empty")
	}
	if controlPlaneMachineConfig.Spec.OSFamily != anywherev1.Bottlerocket && controlPlaneMachineConfig.Spec.OSFamily != anywherev1.Ubuntu {
		return fmt.Errorf("control plane osFamily: %s is not supported, please use one of the following: %s, %s", controlPlaneMachineConfig.Spec.OSFamily, anywherev1.Bottlerocket, anywherev1.Ubuntu)
	}

	workerNodeGroupConfigs := vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations
	for _, workerNodeGroupConfig := range workerNodeGroupConfigs {
		if workerNodeGroupConfig.Name == "" {
			return errors.New("must specify name for worker nodes")
		}
	}

	var workerNodeGroupMachineConfigs []*anywherev1.VSphereMachineConfig
	for _, workerNodeGroupConfiguration := range vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if workerNodeGroupConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for worker nodes")
		}
		workerNodeGroupMachineConfig := vsphereClusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		workerNodeGroupMachineConfigs = append(workerNodeGroupMachineConfigs, workerNodeGroupMachineConfig)
		if workerNodeGroupMachineConfig == nil {
			return fmt.Errorf("cannot find VSphereMachineConfig %v for worker nodes", workerNodeGroupConfiguration.MachineGroupRef.Name)
		}
		if len(workerNodeGroupMachineConfig.Spec.Datastore) <= 0 {
			return errors.New("VSphereMachineConfig datastore for worker nodes is not set or is empty")
		}
		if len(workerNodeGroupMachineConfig.Spec.Folder) <= 0 {
			logger.Info("VSphereMachineConfig folder for worker nodes is not set or is empty. Will default to root vSphere folder.")
		}
		if len(workerNodeGroupMachineConfig.Spec.ResourcePool) <= 0 {
			return errors.New("VSphereMachineConfig VM resourcePool for worker nodes is not set or is empty")
		}
		if workerNodeGroupMachineConfig.Spec.OSFamily != anywherev1.Bottlerocket && workerNodeGroupMachineConfig.Spec.OSFamily != anywherev1.Ubuntu {
			return fmt.Errorf("worker node osFamily: %s is not supported, please use one of the following: %s, %s", workerNodeGroupMachineConfig.Spec.OSFamily, anywherev1.Bottlerocket, anywherev1.Ubuntu)
		}
		if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
			return errors.New("control plane and worker nodes must have the same osFamily specified")
		}
		if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
			return errors.New("control plane and worker nodes must have the same template specified")
		}
	}
	if vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		if vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef == nil {
			return errors.New("must specify machineGroupRef for etcd machines")
		}
		etcdMachineConfig = vsphereClusterSpec.etcdMachineConfig()
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find VSphereMachineConfig %v for etcd machines", vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if len(etcdMachineConfig.Spec.Datastore) <= 0 {
			return errors.New("VSphereMachineConfig datastore for etcd machines is not set or is empty")
		}
		if len(etcdMachineConfig.Spec.Folder) <= 0 {
			logger.Info("VSphereMachineConfig folder for etcd machines is not set or is empty. Will default to root vSphere folder.")
		}
		if len(etcdMachineConfig.Spec.ResourcePool) <= 0 {
			return errors.New("VSphereMachineConfig VM resourcePool for etcd machines is not set or is empty")
		}
	}

	// TODO: move this to api Cluster validations
	if err := v.validateControlPlaneIp(vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err != nil {
		return err
	}

	for _, config := range vsphereClusterSpec.machineConfigsLookup {
		var b bool                                                                                            // Temporary until we remove the need to pass a bool pointer
		err := v.govc.ValidateVCenterSetupMachineConfig(ctx, vsphereClusterSpec.datacenterConfig, config, &b) // TODO: remove side effects from this implementation or directly move it to set defaults (pointer to bool is not needed)
		if err != nil {
			return fmt.Errorf("error validating vCenter setup for VSphereMachineConfig %v: %v", config.Name, err)
		}
	}

	if etcdMachineConfig != nil && etcdMachineConfig.Spec.OSFamily != anywherev1.Bottlerocket && etcdMachineConfig.Spec.OSFamily != anywherev1.Ubuntu {
		return fmt.Errorf("etcd node osFamily: %s is not supported, please use one of the following: %s, %s", etcdMachineConfig.Spec.OSFamily, anywherev1.Bottlerocket, anywherev1.Ubuntu)
	}

	if etcdMachineConfig != nil && controlPlaneMachineConfig.Spec.OSFamily != etcdMachineConfig.Spec.OSFamily {
		return errors.New("control plane and etcd machines must have the same osFamily specified")
	}

	if err := v.validateSSHUsername(controlPlaneMachineConfig); err == nil {
		for _, wnConfig := range workerNodeGroupMachineConfigs {
			if err = v.validateSSHUsername(wnConfig); err != nil {
				return fmt.Errorf("error validating SSHUsername for worker node VSphereMachineConfig %v: %v", wnConfig.Name, err)
			}
		}
		if etcdMachineConfig != nil {
			if err = v.validateSSHUsername(etcdMachineConfig); err != nil {
				return fmt.Errorf("error validating SSHUsername for etcd VSphereMachineConfig %v: %v", etcdMachineConfig.Name, err)
			}
		}
	} else {
		return fmt.Errorf("error validating SSHUsername for control plane VSphereMachineConfig %v: %v", controlPlaneMachineConfig.Name, err)
	}

	for _, machineConfig := range vsphereClusterSpec.machineConfigsLookup {
		if machineConfig.Namespace != vsphereClusterSpec.Cluster.Namespace {
			return errors.New("VSphereMachineConfig and Cluster objects must have the same namespace specified")
		}
	}

	if vsphereClusterSpec.datacenterConfig.Namespace != vsphereClusterSpec.Cluster.Namespace {
		return errors.New("VSphereDatacenterConfig and Cluster objects must have the same namespace specified")
	}

	if err := v.validateTemplate(ctx, vsphereClusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane template validation failed.")
		return err
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	return v.validateDatastoreUsage(ctx, vsphereClusterSpec, controlPlaneMachineConfig, etcdMachineConfig)
}

func (v *Validator) validateControlPlaneIp(ip string) error {
	// check if controlPlaneEndpointIp is valid
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s", ip)
	}
	return nil
}

func (v *Validator) validateSSHUsername(machineConfig *anywherev1.VSphereMachineConfig) error {
	if machineConfig.Spec.OSFamily == anywherev1.Bottlerocket && machineConfig.Spec.Users[0].Name != bottlerocketDefaultUser {
		return fmt.Errorf("SSHUsername %s is invalid. Please use 'ec2-user' for Bottlerocket", machineConfig.Spec.Users[0].Name)
	}
	return nil
}

func (v *Validator) validateTemplate(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	if err := v.validateTemplatePresence(ctx, spec.datacenterConfig.Spec.Datacenter, machineConfig); err != nil {
		return err
	}

	if err := v.validateTemplateTags(ctx, spec, machineConfig); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateTemplatePresence(ctx context.Context, datacenter string, machineConfig *anywherev1.VSphereMachineConfig) error {
	templateFullPath, err := v.govc.SearchTemplate(ctx, datacenter, machineConfig)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found. Has the template been imported?", machineConfig.Spec.Template)
	}

	return nil
}

func (v *Validator) validateTemplateTags(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	tags, err := v.govc.GetTags(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("error validating template tags: %v", err)
	}

	tagsLookup := types.SliceToLookup(tags)
	for _, t := range requiredTemplateTags(spec.Spec, machineConfig) {
		if !tagsLookup.IsPresent(t) {
			// TODO: maybe add help text about to how to tag a template?
			return fmt.Errorf("template %s is missing tag %s", machineConfig.Spec.Template, t)
		}
	}

	return nil
}

type datastoreUsage struct {
	availableSpace float64
	needGiBSpace   int
}

// TODO: cleanup this method signature
// TODO: dry out implementation
func (v *Validator) validateDatastoreUsage(ctx context.Context, vsphereClusterSpec *Spec, controlPlaneMachineConfig *anywherev1.VSphereMachineConfig, etcdMachineConfig *anywherev1.VSphereMachineConfig) error {
	usage := make(map[string]*datastoreUsage)
	controlPlaneAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, controlPlaneMachineConfig.Spec.Datastore) // TODO: remove dependency on machineConfig
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	controlPlaneNeedGiB := controlPlaneMachineConfig.Spec.DiskGiB * vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Count
	usage[controlPlaneMachineConfig.Spec.Datastore] = &datastoreUsage{
		availableSpace: controlPlaneAvailableSpace,
		needGiBSpace:   controlPlaneNeedGiB,
	}

	for _, workerNodeGroupConfiguration := range vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerMachineConfig := vsphereClusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		workerAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, workerMachineConfig.Spec.Datastore)
		if err != nil {
			return fmt.Errorf("error getting datastore details: %v", err)
		}
		workerNeedGiB := workerMachineConfig.Spec.DiskGiB * workerNodeGroupConfiguration.Count
		_, ok := usage[workerMachineConfig.Spec.Datastore]
		if ok {
			usage[workerMachineConfig.Spec.Datastore].needGiBSpace += workerNeedGiB
		} else {
			usage[workerMachineConfig.Spec.Datastore] = &datastoreUsage{
				availableSpace: workerAvailableSpace,
				needGiBSpace:   workerNeedGiB,
			}
		}
	}

	if etcdMachineConfig != nil {
		etcdAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, etcdMachineConfig.Spec.Datastore)
		if err != nil {
			return fmt.Errorf("error getting datastore details: %v", err)
		}
		etcdNeedGiB := etcdMachineConfig.Spec.DiskGiB * vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.Count
		if _, ok := usage[etcdMachineConfig.Spec.Datastore]; ok {
			usage[etcdMachineConfig.Spec.Datastore].needGiBSpace += etcdNeedGiB
		} else {
			usage[etcdMachineConfig.Spec.Datastore] = &datastoreUsage{
				availableSpace: etcdAvailableSpace,
				needGiBSpace:   etcdNeedGiB,
			}
		}
	}

	for datastore, usage := range usage {
		if float64(usage.needGiBSpace) > usage.availableSpace {
			return fmt.Errorf("not enough space in datastore %v for given diskGiB and count for respective machine groups", datastore)
		}
	}
	return nil
}

func (v *Validator) validateThumbprint(ctx context.Context, datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	// No need to validate thumbprint in insecure mode
	if datacenterConfig.Spec.Insecure {
		return nil
	}

	// If cert is not self signed, thumbprint is ignored
	if !v.govc.IsCertSelfSigned(ctx) {
		return nil
	}

	if datacenterConfig.Spec.Thumbprint == "" {
		return fmt.Errorf("thumbprint is required for secure mode with self-signed certificates")
	}

	thumbprint, err := v.govc.GetCertThumbprint(ctx)
	if err != nil {
		return err
	}

	if thumbprint != datacenterConfig.Spec.Thumbprint {
		return fmt.Errorf("thumbprint mismatch detected, expected: %s, actual: %s", datacenterConfig.Spec.Thumbprint, thumbprint)
	}

	return nil
}

func (v *Validator) validateDatacenter(ctx context.Context, datacenter string) error {
	exists, err := v.govc.DatacenterExists(ctx, datacenter)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("datacenter %s not found", datacenter)
	}

	return nil
}

func (v *Validator) validateNetwork(ctx context.Context, network string) error {
	exists, err := v.govc.NetworkExists(ctx, network)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("network %s not found", network)
	}

	return nil
}

func (v *Validator) validateControlPlaneIpUniqueness(spec *Spec) error {
	ip := spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if networkutils.IsIPInUse(v.netClient, ip) {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	}
	return nil
}
