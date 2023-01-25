package vsphere

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/internal/templates"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

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
	for _, m := range spec.machineConfigs() {
		m.SetDefaults()
		if err := d.setDefaultTemplateIfMissing(ctx, spec, m); err != nil {
			return err
		}

		if err := d.setTemplateFullPath(ctx, spec.VSphereDatacenter, m); err != nil {
			return err
		}

		if err := d.setCloneModeDefaults(ctx, m, spec.VSphereDatacenter.Spec.Datacenter); err != nil {
			return err
		}

		d.setDiskSizeDefaults(m)
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

func (d *Defaulter) setDefaultTemplateIfMissing(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	if machineConfig.Spec.Template == "" {
		logger.V(1).Info("Control plane VSphereMachineConfig template is not set. Using default template.")
		if err := d.setupDefaultTemplate(ctx, spec, machineConfig); err != nil {
			return err
		}
	}

	return nil
}

func (d *Defaulter) setupDefaultTemplate(ctx context.Context, spec *Spec, machineConfig *anywherev1.VSphereMachineConfig) error {
	osFamily := machineConfig.Spec.OSFamily
	eksd := spec.VersionsBundle.EksD
	var ova releasev1.Archive
	switch osFamily {
	case anywherev1.Bottlerocket:
		ova = eksd.Ova.Bottlerocket
	default:
		return fmt.Errorf("can not import ova for osFamily: %s, please use %s as osFamily for auto-importing or provide a valid template", osFamily, anywherev1.Bottlerocket)
	}

	templateName := fmt.Sprintf("%s-%s-%s-%s-%s", osFamily, eksd.KubeVersion, eksd.Name, strings.Join(ova.Arch, "-"), ova.SHA256[:7])
	machineConfig.Spec.Template = filepath.Join("/", spec.VSphereDatacenter.Spec.Datacenter, defaultTemplatesFolder, templateName)

	tags := requiredTemplateTagsByCategory(spec.Spec, machineConfig)

	// TODO: figure out if it's worth refactoring the factory to be able to reuse across machine configs.
	templateFactory := templates.NewFactory(d.govc, spec.VSphereDatacenter.Spec.Datacenter, machineConfig.Spec.Datastore, spec.VSphereDatacenter.Spec.Network, machineConfig.Spec.ResourcePool, defaultTemplateLibrary)

	// TODO: remove the factory's dependency on a machineConfig
	if err := templateFactory.CreateIfMissing(ctx, spec.VSphereDatacenter.Spec.Datacenter, machineConfig, ova.URI, tags); err != nil {
		return err
	}

	return nil
}

func (d *Defaulter) setCloneModeDefaults(ctx context.Context, machineConfig *anywherev1.VSphereMachineConfig, datacenter string) error {
	diskSize, err := d.govc.GetVMDiskSizeInGB(ctx, machineConfig.Spec.Template, datacenter)
	if err != nil {
		return fmt.Errorf("getting disk size for template %s: %v", machineConfig.Spec.Template, err)
	}

	templateHasSnapshot, err := d.govc.TemplateHasSnapshot(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("getting template snapshot details: %v", err)
	}

	switch machineConfig.Spec.CloneMode {
	case "":
		if templateHasSnapshot && machineConfig.Spec.DiskGiB == diskSize {
			logger.V(3).Info("CloneMode not set, defaulting to linkedClone", "VSphereMachineConfig", machineConfig.Name)
			machineConfig.Spec.CloneMode = anywherev1.LinkedClone
		} else {
			logger.V(3).Info("CloneMode not set, defaulting to fullClone", "VSphereMachineConfig", machineConfig.Name)
			machineConfig.Spec.CloneMode = anywherev1.FullClone
		}

	case anywherev1.FullClone:
		// do nothing

	case anywherev1.LinkedClone:
		if !templateHasSnapshot {
			return fmt.Errorf(
				"cannot use 'linkedClone' for VSphereMachineConfig '%s' because its template (%s) has no snapshots; create snapshots or change the cloneMode to 'fullClone'",
				machineConfig.Name,
				machineConfig.Spec.Template,
			)
		}
		if machineConfig.Spec.DiskGiB != diskSize {
			return fmt.Errorf(
				"diskGiB cannot be customized for VSphereMachineConfig '%s' when using 'linkedClone'; change the cloneMode to 'fullClone' or the diskGiB to match the template's (%s) disk size of %d GiB",
				machineConfig.Name,
				machineConfig.Spec.Template,
				diskSize,
			)
		}

	default:
		return fmt.Errorf("cloneMode %s is not supported for VSphereMachineConfig %s. Supported clone modes: [%s, %s]", machineConfig.Spec.CloneMode, machineConfig.Name, anywherev1.LinkedClone, anywherev1.FullClone)
	}

	return nil
}

func (d *Defaulter) setDiskSizeDefaults(machineConfig *anywherev1.VSphereMachineConfig) {
	if machineConfig.Spec.DiskGiB < 20 {
		logger.Info("Warning: VSphereMachineConfig DiskGiB cannot be less than 20. Defaulting to 20.")
		machineConfig.Spec.DiskGiB = 20
	}
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
