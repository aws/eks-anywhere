package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/types"
)

type validator struct {
	govc      ProviderGovcClient
	netClient networkutils.NetClient
}

func newValidator(govc ProviderGovcClient, netClient networkutils.NetClient) *validator {
	return &validator{
		govc:      govc,
		netClient: netClient,
	}
}

func (v *validator) validateVCenterAccess(ctx context.Context, server string) error {
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

// TODO: dry out machine configs validations
func (v *validator) validateCluster(ctx context.Context, vsphereClusterSpec *spec) error {
	if len(vsphereClusterSpec.datacenterConfig.Spec.Server) <= 0 {
		return errors.New("VSphereDatacenterConfig server is not set or is empty")
	}

	var etcdMachineConfig *anywherev1.VSphereMachineConfig

	// TODO: move this to api Cluster validations
	if len(vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host) <= 0 {
		return errors.New("cluster controlPlaneConfiguration.Endpoint.Host is not set or is empty")
	}
	if len(vsphereClusterSpec.datacenterConfig.Spec.Datacenter) <= 0 {
		return errors.New("VSphereDatacenterConfig datacenter is not set or is empty")
	}
	if len(vsphereClusterSpec.datacenterConfig.Spec.Network) <= 0 {
		return errors.New("VSphereDatacenterConfig VM network is not set or is empty")
	}
	if err := validatePath(networkFolderType, vsphereClusterSpec.datacenterConfig.Spec.Network, vsphereClusterSpec.datacenterConfig.Spec.Datacenter); err != nil {
		return err
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
	if vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef == nil {
		return errors.New("must specify machineGroupRef for worker nodes")
	}

	workerNodeGroupMachineConfig := vsphereClusterSpec.firstWorkerMachineConfig()
	if workerNodeGroupMachineConfig == nil {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for worker nodes", vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].MachineGroupRef.Name)
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

	if err := v.validateThumbprint(ctx, vsphereClusterSpec); err != nil {
		return err
	}

	if err := v.validateDatacenter(ctx, vsphereClusterSpec); err != nil {
		return err
	}
	logger.MarkPass("Datacenter validated")

	if err := v.validateNetwork(ctx, vsphereClusterSpec); err != nil {
		return err
	}
	logger.MarkPass("Network validated")

	for _, config := range vsphereClusterSpec.machineConfigsLookup {
		var b bool                                                                                            // Temporary until we remove the need to pass a bool pointer
		err := v.govc.ValidateVCenterSetupMachineConfig(ctx, vsphereClusterSpec.datacenterConfig, config, &b) // TODO: remove side effects from this implementation or directly move it to set defaults (pointer to bool is not needed)
		if err != nil {
			return fmt.Errorf("error validating vCenter setup for VSphereMachineConfig %v: %v", config.Name, err)
		}
	}

	if controlPlaneMachineConfig.Spec.OSFamily != workerNodeGroupMachineConfig.Spec.OSFamily {
		return errors.New("control plane and worker nodes must have the same osFamily specified")
	}

	if etcdMachineConfig != nil && controlPlaneMachineConfig.Spec.OSFamily != etcdMachineConfig.Spec.OSFamily {
		return errors.New("control plane and etcd machines must have the same osFamily specified")
	}

	if err := v.validateSSHUsername(controlPlaneMachineConfig); err == nil {
		if err = v.validateSSHUsername(workerNodeGroupMachineConfig); err != nil {
			return fmt.Errorf("error validating SSHUsername for worker node VSphereMachineConfig %v: %v", workerNodeGroupMachineConfig.Name, err)
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

	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		return errors.New("control plane and worker nodes must have the same template specified")
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	return v.validateDatastoreUsage(ctx, vsphereClusterSpec.Spec, controlPlaneMachineConfig, workerNodeGroupMachineConfig, etcdMachineConfig)
}

func (p *validator) validateControlPlaneIp(ip string) error {
	// check if controlPlaneEndpointIp is valid
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s", ip)
	}
	return nil
}

func (p *validator) validateSSHUsername(machineConfig *anywherev1.VSphereMachineConfig) error {
	if machineConfig.Spec.OSFamily == anywherev1.Bottlerocket && machineConfig.Spec.Users[0].Name != bottlerocketDefaultUser {
		return fmt.Errorf("SSHUsername %s is invalid. Please use 'ec2-user' for Bottlerocket", machineConfig.Spec.Users[0].Name)
	}
	return nil
}

func (p *validator) validateTemplate(ctx context.Context, spec *spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	if err := p.validateTemplatePresence(ctx, spec.datacenterConfig.Spec.Datacenter, machineConfig); err != nil {
		return err
	}

	if err := p.validateTemplateTags(ctx, spec, machineConfig); err != nil {
		return err
	}

	return nil
}

func (v *validator) validateTemplatePresence(ctx context.Context, datacenter string, machineConfig *anywherev1.VSphereMachineConfig) error {
	templateFullPath, err := v.govc.SearchTemplate(ctx, datacenter, machineConfig)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found. Has the template been imported?", machineConfig.Spec.Template)
	}

	machineConfig.Spec.Template = templateFullPath // TODO: this is a side effect, it should be in defaults

	return nil
}

func (v *validator) validateTemplateTags(ctx context.Context, spec *spec, machineConfig *anywherev1.VSphereMachineConfig) error {
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
func (v *validator) validateDatastoreUsage(ctx context.Context, clusterSpec *cluster.Spec, controlPlaneMachineConfig *anywherev1.VSphereMachineConfig, workerNodeGroupMachineConfig *anywherev1.VSphereMachineConfig, etcdMachineConfig *anywherev1.VSphereMachineConfig) error {
	usage := make(map[string]*datastoreUsage)
	controlPlaneAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, controlPlaneMachineConfig) // TODO: remove dependency on machineConfig
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	workerAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, workerNodeGroupMachineConfig)
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}

	controlPlaneNeedGiB := controlPlaneMachineConfig.Spec.DiskGiB * clusterSpec.Spec.ControlPlaneConfiguration.Count
	usage[controlPlaneMachineConfig.Spec.Datastore] = &datastoreUsage{
		availableSpace: controlPlaneAvailableSpace,
		needGiBSpace:   controlPlaneNeedGiB,
	}
	workerNeedGiB := workerNodeGroupMachineConfig.Spec.DiskGiB * clusterSpec.Spec.WorkerNodeGroupConfigurations[0].Count
	_, ok := usage[workerNodeGroupMachineConfig.Spec.Datastore]
	if ok {
		usage[workerNodeGroupMachineConfig.Spec.Datastore].needGiBSpace += workerNeedGiB
	} else {
		usage[workerNodeGroupMachineConfig.Spec.Datastore] = &datastoreUsage{
			availableSpace: workerAvailableSpace,
			needGiBSpace:   workerNeedGiB,
		}
	}

	if etcdMachineConfig != nil {
		etcdAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, etcdMachineConfig)
		if err != nil {
			return fmt.Errorf("error getting datastore details: %v", err)
		}
		etcdNeedGiB := etcdMachineConfig.Spec.DiskGiB * clusterSpec.Spec.ExternalEtcdConfiguration.Count
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

func (v *validator) validateThumbprint(ctx context.Context, spec *spec) error {
	// No need to validate thumbprint in insecure mode
	if spec.datacenterConfig.Spec.Insecure {
		return nil
	}

	// If cert is not self signed, thumbprint is ignored
	if !v.govc.IsCertSelfSigned(ctx) {
		return nil
	}

	if spec.datacenterConfig.Spec.Thumbprint == "" {
		return fmt.Errorf("thumbprint is required for secure mode with self-signed certificates")
	}

	thumbprint, err := v.govc.GetCertThumbprint(ctx)
	if err != nil {
		return err
	}

	if thumbprint != spec.datacenterConfig.Spec.Thumbprint {
		return fmt.Errorf("thumbprint mismatch detected, expected: %s, actual: %s", spec.datacenterConfig.Spec.Thumbprint, thumbprint)
	}

	return nil
}

func (v *validator) validateDatacenter(ctx context.Context, spec *spec) error {
	datacenter := spec.datacenterConfig.Spec.Datacenter
	exists, err := v.govc.DatacenterExists(ctx, datacenter)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("datacenter %s not found", datacenter)
	}

	return nil
}

func (v *validator) validateNetwork(ctx context.Context, spec *spec) error {
	network := spec.datacenterConfig.Spec.Network
	exists, err := v.govc.NetworkExists(ctx, network)
	if err != nil {
		return err
	}

	if !exists {
		return fmt.Errorf("network %s not found", network)
	}

	return nil
}

func (v *validator) validateControlPlaneIpUniqueness(spec *spec) error {
	ip := spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host
	if !networkutils.NewIPGenerator(v.netClient).IsIPUnique(ip) {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host <%s> is already in use, please provide a unique IP", ip)
	}
	return nil
}
