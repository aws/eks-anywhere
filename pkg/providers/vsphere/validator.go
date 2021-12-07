package vsphere

import (
	"context"
	"errors"
	"fmt"
	"net"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

type validator struct {
	govc       ProviderGovcClient
	selfSigned bool // TODO: remove/update once
}

func newValidator(govc ProviderGovcClient) *validator {
	return &validator{
		govc:       govc,
		selfSigned: false,
	}
}

// TODO: dry out machine configs validations
func (p *validator) validateCluster(ctx context.Context, vsphereClusterSpec *spec) error {
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
	if err := p.validateControlPlaneIp(vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err != nil {
		return err
	}

	if err := p.govc.ValidateVCenterSetup(ctx, vsphereClusterSpec.datacenterConfig, &p.selfSigned); err != nil { // TODO: remove side effects from this implementation or directly move it to set defaults (try not to pass a pointer to bool to update its value, it's very difficult to follow)
		return fmt.Errorf("error validating vCenter setup: %v", err)
	}

	for _, config := range vsphereClusterSpec.machineConfigsLookup {
		err := p.govc.ValidateVCenterSetupMachineConfig(ctx, vsphereClusterSpec.datacenterConfig, config, &p.selfSigned) // TODO: remove side effects from this implementation or directly move it to set defaults (pointer to bool is not needed)
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

	if err := p.validateSSHUsername(controlPlaneMachineConfig); err == nil {
		if err = p.validateSSHUsername(workerNodeGroupMachineConfig); err != nil {
			return fmt.Errorf("error validating SSHUsername for worker node VSphereMachineConfig %v: %v", workerNodeGroupMachineConfig.Name, err)
		}
		if etcdMachineConfig != nil {
			if err = p.validateSSHUsername(etcdMachineConfig); err != nil {
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

	if err := p.validateTemplate(ctx, vsphereClusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane template validation failed.")
		return err
	}

	// TODO: not sure if this makes any sense since we later validate than the two templates are the same?
	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		if err := p.validateTemplate(ctx, vsphereClusterSpec, workerNodeGroupMachineConfig); err != nil {
			logger.V(1).Info("Workload template validation failed.")
			return err
		}
	}

	if controlPlaneMachineConfig.Spec.Template != workerNodeGroupMachineConfig.Spec.Template {
		return errors.New("control plane and worker nodes must have the same template specified")
	}
	logger.MarkPass("Control plane and Workload templates validated")

	if etcdMachineConfig != nil {
		if err := p.validateTemplate(ctx, vsphereClusterSpec, etcdMachineConfig); err != nil {
			logger.V(1).Info("Etcd template validation failed.")
			return err
		}
		if etcdMachineConfig.Spec.Template != controlPlaneMachineConfig.Spec.Template {
			return errors.New("control plane and etcd machines must have the same template specified")
		}
	}

	return p.validateDatastoreUsage(ctx, vsphereClusterSpec.Spec, controlPlaneMachineConfig, workerNodeGroupMachineConfig, etcdMachineConfig)
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

func (p *validator) validateTemplatePresence(ctx context.Context, datacenter string, machineConfig *anywherev1.VSphereMachineConfig) error {
	templateFullPath, err := p.govc.SearchTemplate(ctx, datacenter, machineConfig)
	if err != nil {
		return fmt.Errorf("error validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found. Has the template been imported?", machineConfig.Spec.Template)
	}

	machineConfig.Spec.Template = templateFullPath // TODO: this is a side effect, it should be in defaults

	return nil
}

func (p *validator) validateTemplateTags(ctx context.Context, spec *spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	tags, err := p.govc.GetTags(ctx, machineConfig.Spec.Template)
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
func (p *validator) validateDatastoreUsage(ctx context.Context, clusterSpec *cluster.Spec, controlPlaneMachineConfig *anywherev1.VSphereMachineConfig, workerNodeGroupMachineConfig *anywherev1.VSphereMachineConfig, etcdMachineConfig *anywherev1.VSphereMachineConfig) error {
	usage := make(map[string]*datastoreUsage)
	var etcdAvailableSpace float64
	controlPlaneAvailableSpace, err := p.govc.GetWorkloadAvailableSpace(ctx, controlPlaneMachineConfig) // TODO: remove dependency on machineConfig
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	workerAvailableSpace, err := p.govc.GetWorkloadAvailableSpace(ctx, workerNodeGroupMachineConfig)
	if err != nil {
		return fmt.Errorf("error getting datastore details: %v", err)
	}
	etcdAvailableSpace, err = p.govc.GetWorkloadAvailableSpace(ctx, etcdMachineConfig)
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
