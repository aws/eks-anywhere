package vsphere

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"path/filepath"

	"gopkg.in/yaml.v2"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/govmomi"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	vsphereRootPath = "/"
)

type PrivAssociation struct {
	objectType   string
	privsContent string
	path         string
}

type missingPriv struct {
	Username    string   `yaml:"username"`
	ObjectType  string   `yaml:"objectType"`
	Path        string   `yaml:"path"`
	Permissions []string `yaml:"permissions"`
}

type VSphereClientBuilder interface {
	Build(ctx context.Context, host string, username string, password string, insecure bool, datacenter string) (govmomi.VSphereClient, error)
}

type Validator struct {
	govc                 ProviderGovcClient
	vSphereClientBuilder VSphereClientBuilder
}

// NewValidator initializes the client for VSphere provider validations.
func NewValidator(govc ProviderGovcClient, vscb VSphereClientBuilder) *Validator {
	return &Validator{
		govc:                 govc,
		vSphereClientBuilder: vscb,
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

// TODO: dry out machine configs validations.
func (v *Validator) ValidateClusterMachineConfigs(ctx context.Context, vsphereClusterSpec *Spec) error {
	var etcdMachineConfig *anywherev1.VSphereMachineConfig

	controlPlaneMachineConfig := vsphereClusterSpec.controlPlaneMachineConfig()
	if controlPlaneMachineConfig == nil {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for control plane", vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name)
	}

	for _, workerNodeGroupConfiguration := range vsphereClusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineConfig := vsphereClusterSpec.workerMachineConfig(workerNodeGroupConfiguration)
		if workerNodeGroupMachineConfig == nil {
			return fmt.Errorf("cannot find VSphereMachineConfig %v for worker nodes", workerNodeGroupConfiguration.MachineGroupRef.Name)
		}
	}
	if vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig = vsphereClusterSpec.etcdMachineConfig()
		if etcdMachineConfig == nil {
			return fmt.Errorf("cannot find VSphereMachineConfig %v for etcd machines", vsphereClusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name)
		}
		if !v.sameOSFamily(vsphereClusterSpec.VSphereMachineConfigs) {
			return errors.New("all VSphereMachineConfigs must have the same osFamily specified")
		}
		if !v.sameTemplate(vsphereClusterSpec.VSphereMachineConfigs) {
			return errors.New("all VSphereMachineConfigs must have the same template specified")
		}
	}

	// TODO: move this to api Cluster validations
	if err := v.validateControlPlaneIp(vsphereClusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err != nil {
		return err
	}

	for _, config := range vsphereClusterSpec.VSphereMachineConfigs {
		var b bool                                                                                             // Temporary until we remove the need to pass a bool pointer
		err := v.govc.ValidateVCenterSetupMachineConfig(ctx, vsphereClusterSpec.VSphereDatacenter, config, &b) // TODO: remove side effects from this implementation or directly move it to set defaults (pointer to bool is not needed)
		if err != nil {
			return fmt.Errorf("validating vCenter setup for VSphereMachineConfig %v: %v", config.Name, err)
		}
	}

	if err := v.validateTemplate(ctx, vsphereClusterSpec, controlPlaneMachineConfig); err != nil {
		logger.V(1).Info("Control plane template validation failed.")
		return err
	}
	logger.MarkPass("Control plane and Workload templates validated")

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

func (v *Validator) validateTemplate(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	if err := v.validateTemplatePresence(ctx, spec.VSphereDatacenter.Spec.Datacenter, machineConfig); err != nil {
		return err
	}

	if err := v.validateTemplateTags(ctx, spec, machineConfig); err != nil {
		return err
	}

	return nil
}

func (v *Validator) validateTemplatePresence(ctx context.Context, datacenter string, machineConfig *anywherev1.VSphereMachineConfig) error {
	templateFullPath, err := v.govc.SearchTemplate(ctx, datacenter, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found. Has the template been imported?", machineConfig.Spec.Template)
	}

	return nil
}

func (v *Validator) validateTemplateTags(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	tags, err := v.govc.GetTags(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("validating template tags: %v", err)
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
// TODO: dry out implementation.
func (v *Validator) validateDatastoreUsage(ctx context.Context, vsphereClusterSpec *Spec, controlPlaneMachineConfig *anywherev1.VSphereMachineConfig, etcdMachineConfig *anywherev1.VSphereMachineConfig) error {
	usage := make(map[string]*datastoreUsage)
	controlPlaneAvailableSpace, err := v.govc.GetWorkloadAvailableSpace(ctx, controlPlaneMachineConfig.Spec.Datastore) // TODO: remove dependency on machineConfig
	if err != nil {
		return fmt.Errorf("getting datastore details: %v", err)
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
			return fmt.Errorf("getting datastore details: %v", err)
		}
		workerNeedGiB := workerMachineConfig.Spec.DiskGiB * *workerNodeGroupConfiguration.Count
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
			return fmt.Errorf("getting datastore details: %v", err)
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

func (v *Validator) collectSpecMachineConfigs(ctx context.Context, spec *Spec) ([]*anywherev1.VSphereMachineConfig, error) {
	controlPlaneMachineConfig := spec.controlPlaneMachineConfig()
	machineConfigs := []*anywherev1.VSphereMachineConfig{controlPlaneMachineConfig}

	for _, workerNodeGroupConfiguration := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		workerNodeGroupMachineConfig := spec.workerMachineConfig(workerNodeGroupConfiguration)
		machineConfigs = append(machineConfigs, workerNodeGroupMachineConfig)
	}

	if spec.Cluster.Spec.ExternalEtcdConfiguration != nil {
		etcdMachineConfig := spec.etcdMachineConfig()
		machineConfigs = append(machineConfigs, etcdMachineConfig)
	}

	return machineConfigs, nil
}

func (v *Validator) validateUserPrivs(ctx context.Context, spec *Spec, vuc *config.VSphereUserConfig) (bool, error) {
	machineConfigs, err := v.collectSpecMachineConfigs(ctx, spec)
	if err != nil {
		return false, err
	}

	requiredPrivAssociations := []PrivAssociation{
		// validate global root priv settings are correct
		{
			objectType:   govmomi.VSphereTypeFolder,
			privsContent: config.VSphereGlobalPrivsFile,
			path:         vsphereRootPath,
		},
		{
			objectType:   govmomi.VSphereTypeNetwork,
			privsContent: config.VSphereUserPrivsFile,
			path:         spec.VSphereDatacenter.Spec.Network,
		},
	}

	seen := map[string]interface{}{}
	for _, mc := range machineConfigs {

		if _, ok := seen[mc.Spec.Datastore]; !ok {
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeDatastore,
				privsContent: config.VSphereUserPrivsFile,
				path:         mc.Spec.Datastore,
			},
			)
			seen[mc.Spec.Datastore] = 1
		}
		if _, ok := seen[mc.Spec.ResourcePool]; !ok {
			// do something here
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeResourcePool,
				privsContent: config.VSphereUserPrivsFile,
				path:         mc.Spec.ResourcePool,
			})
			seen[mc.Spec.ResourcePool] = 1
		}
		if _, ok := seen[mc.Spec.Folder]; !ok {
			// validate Administrator role (all privs) on VM folder and Template folder
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeFolder,
				privsContent: config.VSphereAdminPrivsFile,
				path:         mc.Spec.Folder,
			})
			seen[mc.Spec.Folder] = 1
		}

		if _, ok := seen[mc.Spec.Template]; !ok {
			// ToDo: add more sophisticated validation around a scenario where someone has uploaded templates
			// on their own and does not want to allow EKSA user write access to templates
			// Verify privs on the template
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeVirtualMachine,
				privsContent: config.VSphereAdminPrivsFile,
				path:         mc.Spec.Template,
			})
			seen[mc.Spec.Template] = 1
		}

		if _, ok := seen[filepath.Dir(mc.Spec.Template)]; !ok {
			// Verify privs on the template directory
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeFolder,
				privsContent: config.VSphereAdminPrivsFile,
				path:         filepath.Dir(mc.Spec.Template),
			})

			seen[filepath.Dir(mc.Spec.Template)] = 1
		}
	}

	host := spec.VSphereDatacenter.Spec.Server
	datacenter := spec.VSphereDatacenter.Spec.Datacenter

	vsc, err := v.vSphereClientBuilder.Build(
		ctx,
		host,
		vuc.EksaVsphereUsername,
		vuc.EksaVspherePassword,
		spec.VSphereDatacenter.Spec.Insecure,
		datacenter,
	)
	if err != nil {
		return false, err
	}

	return v.validatePrivs(ctx, requiredPrivAssociations, vsc)
}

func (v *Validator) validateCSIUserPrivs(ctx context.Context, spec *Spec, vuc *config.VSphereUserConfig) (bool, error) {
	requiredPrivAssociations := []PrivAssociation{
		{ // CNS-SEARCH-AND-SPBM role
			objectType:   govmomi.VSphereTypeFolder,
			privsContent: config.VSphereCnsSearchAndSpbmPrivsFile,
			path:         vsphereRootPath,
		},
	}

	machineConfigs, err := v.collectSpecMachineConfigs(ctx, spec)
	if err != nil {
		return false, err
	}

	var pas []PrivAssociation
	seen := map[string]interface{}{}
	for _, mc := range machineConfigs {
		if _, ok := seen[mc.Spec.Datastore]; !ok {
			requiredPrivAssociations = append(
				requiredPrivAssociations,
				// CNS-Datastore role
				PrivAssociation{
					objectType:   govmomi.VSphereTypeDatastore,
					privsContent: config.VSphereCnsDatastorePrivsFile,
					path:         mc.Spec.Datastore,
				},
				// CNS-HOST-CONFIG-STORAGE role
				PrivAssociation{
					objectType:   govmomi.VSphereTypeDatastore,
					privsContent: config.VSphereCnsHostConfigStorageFile,
					path:         mc.Spec.Datastore,
				},
			)
			seen[mc.Spec.Datastore] = 1
		}
		if _, ok := seen[mc.Spec.Folder]; !ok {
			// CNS-VM role
			requiredPrivAssociations = append(requiredPrivAssociations, PrivAssociation{
				objectType:   govmomi.VSphereTypeFolder,
				privsContent: config.VSphereCnsVmPrivsFile,
				path:         mc.Spec.Folder,
			})
			seen[mc.Spec.Folder] = 1
		}

		requiredPrivAssociations = append(requiredPrivAssociations, pas...)
	}

	host := spec.VSphereDatacenter.Spec.Server
	datacenter := spec.VSphereDatacenter.Spec.Datacenter

	vsc, err := v.vSphereClientBuilder.Build(
		ctx,
		host,
		vuc.EksaVsphereCSIUsername,
		vuc.EksaVsphereCSIPassword,
		spec.VSphereDatacenter.Spec.Insecure,
		datacenter,
	)
	if err != nil {
		return false, err
	}

	return v.validatePrivs(ctx, requiredPrivAssociations, vsc)
}

func (v *Validator) validateCPUserPrivs(ctx context.Context, spec *Spec, vuc *config.VSphereUserConfig) (bool, error) {
	// CP role just needs read only
	privObjs := []PrivAssociation{
		{
			objectType:   govmomi.VSphereTypeFolder,
			privsContent: config.VSphereReadOnlyPrivs,
			path:         vsphereRootPath,
		},
	}

	host := spec.VSphereDatacenter.Spec.Server
	datacenter := spec.VSphereDatacenter.Spec.Datacenter

	vsc, err := v.vSphereClientBuilder.Build(
		ctx,
		host,
		vuc.EksaVsphereCPUsername,
		vuc.EksaVsphereCPPassword,
		spec.VSphereDatacenter.Spec.Insecure,
		datacenter,
	)
	if err != nil {
		return false, err
	}

	return v.validatePrivs(ctx, privObjs, vsc)
}

func (v *Validator) validatePrivs(ctx context.Context, privObjs []PrivAssociation, vsc govmomi.VSphereClient) (bool, error) {
	var privs []string
	var err error
	missingPrivs := []missingPriv{}
	passed := false

	for _, obj := range privObjs {
		path := obj.path
		privsContent := obj.privsContent
		t := obj.objectType
		username := vsc.Username()
		privs, err = v.getMissingPrivs(ctx, vsc, path, t, privsContent, username)
		if err != nil {
			return passed, err
		} else if len(privs) > 0 {
			mp := missingPriv{
				Username:    username,
				ObjectType:  t,
				Path:        path,
				Permissions: privs,
			}
			missingPrivs = append(missingPrivs, mp)
			content, err := yaml.Marshal(mp)
			if err == nil {
				s := fmt.Sprintf("  Warning: User %s missing %d vSphere permissions on %s, cluster creation may fail.\nRe-run create cluster with --verbosity=3 to see specific missing permissions.", username, len(privs), path)
				logger.MarkWarning(s)
				s = fmt.Sprintf("Missing Permissions:\n%s", string(content))
				logger.V(3).Info(s)
			} else {
				s := fmt.Sprintf("  Warning: failed to list missing privs: %v", err)
				logger.MarkWarning(s)
			}
		}
	}

	if len(missingPrivs) == 0 {
		passed = true
	}

	return passed, nil
}

func checkRequiredPrivs(requiredPrivs []string, hasPrivs []string) []string {
	hp := map[string]interface{}{}
	for _, val := range hasPrivs {
		hp[val] = 1
	}

	missingPrivs := []string{}
	for _, p := range requiredPrivs {
		if _, ok := hp[p]; !ok {
			missingPrivs = append(missingPrivs, p)
		}
	}

	return missingPrivs
}

func (v *Validator) getMissingPrivs(ctx context.Context, vsc govmomi.VSphereClient, path string, objType string, requiredPrivsContent string, username string) ([]string, error) {
	var requiredPrivs []string
	err := json.Unmarshal([]byte(requiredPrivsContent), &requiredPrivs)
	if err != nil {
		return nil, err
	}

	hasPrivs, err := vsc.GetPrivsOnEntity(ctx, path, objType, username)
	if err != nil {
		return nil, err
	}

	missingPrivs := checkRequiredPrivs(requiredPrivs, hasPrivs)

	return missingPrivs, nil
}

func (v *Validator) sameOSFamily(configs map[string]*anywherev1.VSphereMachineConfig) bool {
	c := getRandomMachineConfig(configs)
	osFamily := c.Spec.OSFamily

	for _, machineConfig := range configs {
		if machineConfig.Spec.OSFamily != osFamily {
			return false
		}
	}
	return true
}

func (v *Validator) sameTemplate(configs map[string]*anywherev1.VSphereMachineConfig) bool {
	c := getRandomMachineConfig(configs)
	template := c.Spec.Template

	for _, machineConfig := range configs {
		if machineConfig.Spec.Template != template {
			return false
		}
	}
	return true
}

func getRandomMachineConfig(configs map[string]*anywherev1.VSphereMachineConfig) *anywherev1.VSphereMachineConfig {
	var machineConfig *anywherev1.VSphereMachineConfig
	for _, c := range configs {
		machineConfig = c
		break
	}
	return machineConfig
}
