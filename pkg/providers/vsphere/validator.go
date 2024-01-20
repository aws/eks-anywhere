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

func (v *Validator) validateMachineConfigTagsExist(ctx context.Context, machineConfigs []*anywherev1.VSphereMachineConfig) error {
	tags, err := v.govc.ListTags(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if tags exists in vSphere: %v", err)
	}

	tagIDs := make([]string, 0, len(tags))
	for _, t := range tags {
		tagIDs = append(tagIDs, t.Id)
	}

	idLookup := types.SliceToLookup(tagIDs)
	for _, machineConfig := range machineConfigs {
		for _, tagID := range machineConfig.Spec.TagIDs {
			if !idLookup.IsPresent(tagID) {
				return fmt.Errorf("tag (%s) does not exist in vSphere. please provide a valid tag id in the urn format (example: urn:vmomi:InventoryServiceTag:8e0ce079-0677-48d6-8865-19ada4e6dabd:GLOBAL)", tagID)
			}
		}
	}
	logger.MarkPass("Machine config tags validated")

	return nil
}

// ValidateClusterMachineConfigs validates all the attributes of etcd, control plane, and worker node VSphereMachineConfigs.
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
		if etcdMachineConfig.Spec.HostOSConfiguration != nil && etcdMachineConfig.Spec.HostOSConfiguration.BottlerocketConfiguration != nil && etcdMachineConfig.Spec.HostOSConfiguration.BottlerocketConfiguration.Kubernetes != nil {
			logger.Info("Bottlerocket Kubernetes settings are not supported for etcd machines. Ignoring Kubernetes settings for etcd machines.", "etcdMachineConfig", etcdMachineConfig.Name)
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

	if err := v.validateTemplates(ctx, vsphereClusterSpec); err != nil {
		return err
	}

	if err := v.validateMachineConfigTagsExist(ctx, vsphereClusterSpec.machineConfigs()); err != nil {
		return err
	}

	logger.MarkPass("Control plane and Workload templates validated")

	for _, mc := range vsphereClusterSpec.VSphereMachineConfigs {
		if mc.OSFamily() == anywherev1.Bottlerocket {
			if err := v.validateBRHardDiskSize(ctx, vsphereClusterSpec, mc); err != nil {
				return fmt.Errorf("failed validating BR Hard Disk size: %v", err)
			}
		}
	}

	return nil
}

func (v *Validator) validateControlPlaneIp(ip string) error {
	// check if controlPlaneEndpointIp is valid
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		return fmt.Errorf("cluster controlPlaneConfiguration.Endpoint.Host is invalid: %s", ip)
	}
	return nil
}

func (v *Validator) validateTemplates(ctx context.Context, spec *Spec) error {
	tagsForTemplates := make(map[string][]string)
	rootVersionsBundle := spec.RootVersionsBundle()
	for _, m := range sliceIfNotNil(spec.controlPlaneMachineConfig(), spec.etcdMachineConfig()) {
		currentTags := tagsForTemplates[m.Spec.Template]
		tagsForTemplates[m.Spec.Template] = append(
			currentTags,
			requiredTemplateTags(m, rootVersionsBundle)...,
		)
	}

	for _, w := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		machineConfig := spec.VSphereMachineConfigs[w.MachineGroupRef.Name]
		versionsBundle := spec.WorkerNodeGroupVersionsBundle(w)

		currentTags := tagsForTemplates[machineConfig.Spec.Template]
		tagsForTemplates[machineConfig.Spec.Template] = append(
			currentTags,
			requiredTemplateTags(machineConfig, versionsBundle)...,
		)
	}

	for template, requiredTags := range tagsForTemplates {
		datacenter := spec.VSphereDatacenter.Spec.Datacenter

		templatePath, err := v.getTemplatePath(ctx, datacenter, template)
		if err != nil {
			return err
		}

		if err := v.validateTemplateTags(ctx, templatePath, requiredTags); err != nil {
			return err
		}
	}

	return nil
}

func (v *Validator) getTemplatePath(ctx context.Context, datacenter, templatePath string) (string, error) {
	templateFullPath, err := v.govc.SearchTemplate(ctx, datacenter, templatePath)
	if err != nil {
		return "", fmt.Errorf("validating template: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return "", fmt.Errorf("template <%s> not found. Has the template been imported?", templatePath)
	}

	return templateFullPath, nil
}

func (v *Validator) validateTemplateTags(ctx context.Context, templatePath string, requiredTags []string) error {
	tags, err := v.govc.GetTags(ctx, templatePath)
	if err != nil {
		return fmt.Errorf("validating template tags: %v", err)
	}

	tagsLookup := types.SliceToLookup(tags)
	for _, t := range requiredTags {
		if !tagsLookup.IsPresent(t) {
			// TODO: maybe add help text about to how to tag a template?
			return fmt.Errorf("template %s is missing tag %s", templatePath, t)
		}
	}

	return nil
}

func (v *Validator) validateBRHardDiskSize(ctx context.Context, spec *Spec, machineConfigSpec *anywherev1.VSphereMachineConfig) error {
	dataCenter := spec.Config.VSphereDatacenter.Spec.Datacenter
	template := machineConfigSpec.Spec.Template
	hardDiskMap, err := v.govc.GetHardDiskSize(ctx, template, dataCenter)
	if err != nil {
		return fmt.Errorf("validating hard disk size: %v", err)
	}
	if len(hardDiskMap) == 0 {
		return fmt.Errorf("no hard disks found for template: %v", template)
	} else if len(hardDiskMap) > 1 {
		if hardDiskMap[disk1] != 2097152 { // 2GB in KB to avoid roundoff errors
			return fmt.Errorf("Incorrect disk size for disk1 - expected: 2097152 kB got: %v", hardDiskMap[disk1])
		} else if hardDiskMap[disk2] != 20971520 { // 20GB in KB to avoid roundoff errors
			return fmt.Errorf("Incorrect disk size for disk2 - expected: 20971520 kB got: %v", hardDiskMap[disk2])
		}
	} else if hardDiskMap[disk1] != 23068672 { // 22GB in KB to avoid roundoff errors
		return fmt.Errorf("Incorrect disk size for disk1 - expected: 23068672 kB got: %v", hardDiskMap[disk1])
	}
	logger.V(5).Info("Bottlerocket Disk size validated: ", "diskMap", hardDiskMap)
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

func (v *Validator) validateVsphereUserPrivs(ctx context.Context, vSphereClusterSpec *Spec) error {
	var passed bool
	var err error
	vuc := config.NewVsphereUserConfig()

	if passed, err = v.validateUserPrivs(ctx, vSphereClusterSpec, vuc); err != nil {
		return err
	}
	markPrivsValidationPass(passed, vuc.EksaVsphereUsername)

	if len(vuc.EksaVsphereCPUsername) > 0 && vuc.EksaVsphereCPUsername != vuc.EksaVsphereUsername {
		if passed, err = v.validateCPUserPrivs(ctx, vSphereClusterSpec, vuc); err != nil {
			return err
		}
		markPrivsValidationPass(passed, vuc.EksaVsphereCPUsername)
	}

	return nil
}

func markPrivsValidationPass(passed bool, username string) {
	if passed {
		s := fmt.Sprintf("%s user vSphere privileges validated", username)
		logger.MarkPass(s)
	}
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
	username := vsc.Username()

	for _, obj := range privObjs {
		path := obj.path
		privsContent := obj.privsContent
		t := obj.objectType
		privs, err = v.getMissingPrivs(ctx, vsc, path, t, privsContent, username)
		if err != nil {
			return passed, fmt.Errorf("failed to get missing privileges: %v", err)
		} else if len(privs) > 0 {
			mp := missingPriv{
				Username:    username,
				ObjectType:  t,
				Path:        path,
				Permissions: privs,
			}
			missingPrivs = append(missingPrivs, mp)
		}
	}

	if len(missingPrivs) != 0 {
		content, err := yaml.Marshal(missingPrivs)
		if err != nil {
			return passed, fmt.Errorf("failed to marshal missing permissions: %v", err)
		}

		errMsg := fmt.Sprintf("user %s missing vSphere permissions", username)
		logger.V(3).Info(errMsg, "Permissions", string(content))

		return passed, fmt.Errorf("user %s missing vSphere permissions", username)
	}

	passed = true

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

func sliceIfNotNil(machines ...*anywherev1.VSphereMachineConfig) []*anywherev1.VSphereMachineConfig {
	var notNil []*anywherev1.VSphereMachineConfig
	for _, m := range machines {
		if m != nil {
			notNil = append(notNil, m)
		}
	}

	return notNil
}
