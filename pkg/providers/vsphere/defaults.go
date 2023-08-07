package vsphere

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/templates"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const minDiskGib int = 20

type Defaulter struct {
	govc ProviderGovcClient
}

func NewDefaulter(govc ProviderGovcClient) *Defaulter {
	return &Defaulter{
		govc: govc,
	}
}

func (d *Defaulter) setDefaultsForMachineConfig(ctx context.Context, spec *Spec) error {
	setDefaultsForEtcdMachineConfig(spec.etcdMachineConfig())

	for _, w := range spec.Cluster.Spec.WorkerNodeGroupConfigurations {
		if err := d.setWorkerDefaultTemplateIfMissing(ctx, spec, w); err != nil {
			return err
		}
	}

	for _, m := range spec.machineConfigs() {
		m.SetDefaults()
		m.SetUserDefaults()

		if err := d.setDefaultTemplateIfMissing(ctx, spec, m); err != nil {
			return err
		}

		if err := d.setTemplateFullPath(ctx, spec.VSphereDatacenter, m); err != nil {
			return err
		}

		if err := d.setCloneModeAndDiskSizeDefaults(ctx, m, spec.VSphereDatacenter.Spec.Datacenter); err != nil {
			return err
		}
	}

	return nil
}

func (d *Defaulter) SetDefaultsForDatacenterConfig(ctx context.Context, datacenterConfig *anywherev1.VSphereDatacenterConfig) error {
	datacenterConfig.SetDefaults()

	if datacenterConfig.Spec.Thumbprint != "" {
		if err := d.govc.ConfigureCertThumbprint(ctx, datacenterConfig.Spec.Server, datacenterConfig.Spec.Thumbprint); err != nil {
			return fmt.Errorf("failed configuring govc cert thumbprint: %v", err)
		}
	}

	return nil
}

func setDefaultsForEtcdMachineConfig(machineConfig *anywherev1.VSphereMachineConfig) {
	if machineConfig != nil && machineConfig.Spec.MemoryMiB < 8192 {
		logger.Info("Warning: VSphereMachineConfig MemoryMiB for etcd machines should not be less than 8192. Defaulting to 8192")
		machineConfig.Spec.MemoryMiB = 8192
	}
}

func (d *Defaulter) setWorkerDefaultTemplateIfMissing(ctx context.Context, spec *Spec, workerNodeGroup anywherev1.WorkerNodeGroupConfiguration) error {
	machineConfigName := workerNodeGroup.MachineGroupRef.Name
	machineConfig := spec.VSphereMachineConfigs[machineConfigName]
	if machineConfig == nil {
		return fmt.Errorf("cannot find VSphereMachineConfig %v for worker nodes", machineConfigName)
	}
	if machineConfig.Spec.Template == "" {
		logger.V(1).Info("Worker node VSphereMachineConfig template is not set. Using default template.")

		versionsBundle := spec.WorkerNodeGroupVersionsBundle(workerNodeGroup)
		if err := d.setupDefaultTemplate(ctx, spec, machineConfig, versionsBundle); err != nil {
			return err
		}
	}

	return nil
}

func (d *Defaulter) setDefaultTemplateIfMissing(ctx context.Context, spec *Spec, m *anywherev1.VSphereMachineConfig) error {
	if m.Spec.Template == "" {
		logger.V(1).Info("VSphereMachineConfig template is not set. Using default template.")
		versionsBundle := spec.RootVersionsBundle()
		if err := d.setupDefaultTemplate(ctx, spec, m, versionsBundle); err != nil {
			return err
		}
	}

	return nil
}

func (d *Defaulter) setupDefaultTemplate(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig, versionsBundle *cluster.VersionsBundle) error {
	osFamily := machineConfig.Spec.OSFamily
	eksd := versionsBundle.EksD
	var ova releasev1.Archive
	switch osFamily {
	case anywherev1.Bottlerocket:
		ova = eksd.Ova.Bottlerocket
	default:
		return fmt.Errorf("can not import ova for osFamily: %s, please use %s as osFamily for auto-importing or provide a valid template", osFamily, anywherev1.Bottlerocket)
	}

	templateName := fmt.Sprintf("%s-%s-%s-%s-%s", osFamily, eksd.KubeVersion, eksd.Name, strings.Join(ova.Arch, "-"), ova.SHA256[:7])
	machineConfig.Spec.Template = filepath.Join("/", spec.VSphereDatacenter.Spec.Datacenter, defaultTemplatesFolder, templateName)

	tags := requiredTemplateTagsByCategory(machineConfig, versionsBundle)

	// TODO: figure out if it's worth refactoring the factory to be able to reuse across machine configs.
	templateFactory := templates.NewFactory(d.govc, spec.VSphereDatacenter.Spec.Datacenter, machineConfig.Spec.Datastore, spec.VSphereDatacenter.Spec.Network, machineConfig.Spec.ResourcePool, defaultTemplateLibrary)

	// TODO: remove the factory's dependency on a machineConfig
	if err := templateFactory.CreateIfMissing(ctx, spec.VSphereDatacenter.Spec.Datacenter, machineConfig, ova.URI, tags); err != nil {
		return err
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (d *Defaulter) setCloneModeAndDiskSizeDefaults(ctx context.Context, machineConfig *anywherev1.VSphereMachineConfig, datacenter string) error {
	templateDiskSize, err := d.govc.GetVMDiskSizeInGB(ctx, machineConfig.Spec.Template, datacenter)
	if err != nil {
		return fmt.Errorf("getting disk size for template %s: %v", machineConfig.Spec.Template, err)
	}

	minDiskSize := max(minDiskGib, templateDiskSize)

	if machineConfig.Spec.DiskGiB < minDiskSize {
		errStr := fmt.Sprintf("Warning: VSphereMachineConfig DiskGiB cannot be less than %v. Defaulting to %v.", minDiskSize, minDiskSize)
		logger.Info(errStr)
		machineConfig.Spec.DiskGiB = minDiskSize
	}

	templateHasSnapshot, err := d.govc.TemplateHasSnapshot(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("getting template snapshot details: %v", err)
	}

	if machineConfig.Spec.CloneMode == anywherev1.FullClone {
		return nil
	}

	if machineConfig.Spec.CloneMode == anywherev1.LinkedClone {
		return validateMachineWithLinkedCloneMode(templateHasSnapshot, templateDiskSize, machineConfig)
	}

	if machineConfig.Spec.CloneMode == "" {
		return validateMachineWithNoCloneMode(templateHasSnapshot, templateDiskSize, machineConfig)
	}

	return fmt.Errorf("cloneMode %s is not supported for VSphereMachineConfig %s. Supported clone modes: [%s, %s]", machineConfig.Spec.CloneMode, machineConfig.Name, anywherev1.LinkedClone, anywherev1.FullClone)
}

func validateMachineWithNoCloneMode(templateHasSnapshot bool, templateDiskSize int, machineConfig *anywherev1.VSphereMachineConfig) error {
	if templateHasSnapshot && machineConfig.Spec.DiskGiB == templateDiskSize {
		logger.V(3).Info("CloneMode not set, defaulting to linkedClone", "VSphereMachineConfig", machineConfig.Name)
		machineConfig.Spec.CloneMode = anywherev1.LinkedClone
	} else {
		logger.V(3).Info("CloneMode not set, defaulting to fullClone", "VSphereMachineConfig", machineConfig.Name)
		machineConfig.Spec.CloneMode = anywherev1.FullClone
	}
	return nil
}

func validateMachineWithLinkedCloneMode(templateHasSnapshot bool, templateDiskSize int, machineConfig *anywherev1.VSphereMachineConfig) error {
	if !templateHasSnapshot {
		return fmt.Errorf(
			"cannot use 'linkedClone' for VSphereMachineConfig '%s' because its template (%s) has no snapshots; create snapshots or change the cloneMode to 'fullClone'",
			machineConfig.Name,
			machineConfig.Spec.Template,
		)
	}
	if machineConfig.Spec.DiskGiB != templateDiskSize {
		return fmt.Errorf(
			"diskGiB cannot be customized for VSphereMachineConfig '%s' when using 'linkedClone'; change the cloneMode to 'fullClone' or the diskGiB to match the template's (%s) disk size of %d GiB",
			machineConfig.Name,
			machineConfig.Spec.Template,
			templateDiskSize,
		)
	}
	return nil
}

func (d *Defaulter) setTemplateFullPath(ctx context.Context,
	datacenterConfig *anywherev1.VSphereDatacenterConfig,
	machine *anywherev1.VSphereMachineConfig,
) error {
	templateFullPath, err := d.govc.SearchTemplate(ctx, datacenterConfig.Spec.Datacenter, machine.Spec.Template)
	if err != nil {
		return fmt.Errorf("setting template full path: %v", err)
	}

	if len(templateFullPath) <= 0 {
		return fmt.Errorf("template <%s> not found", machine.Spec.Template)
	}

	machine.Spec.Template = templateFullPath
	return nil
}
