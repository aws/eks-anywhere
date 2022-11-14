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

		if err := d.setTemplateFullPath(ctx, spec.datacenterConfig, m); err != nil {
			return err
		}

		if err := d.setDiskDefaults(ctx, m); err != nil {
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
	machineConfig.Spec.Template = filepath.Join("/", spec.datacenterConfig.Spec.Datacenter, defaultTemplatesFolder, templateName)

	tags := requiredTemplateTagsByCategory(spec.Spec, machineConfig)

	// TODO: figure out if it's worth refactoring the factory to be able to reuse across machine configs.
	templateFactory := templates.NewFactory(d.govc, spec.datacenterConfig.Spec.Datacenter, machineConfig.Spec.Datastore, spec.datacenterConfig.Spec.Network, machineConfig.Spec.ResourcePool, defaultTemplateLibrary)

	// TODO: remove the factory's dependency on a machineConfig
	if err := templateFactory.CreateIfMissing(ctx, spec.datacenterConfig.Spec.Datacenter, machineConfig, ova.URI, tags); err != nil {
		return err
	}

	return nil
}

func (d *Defaulter) setDiskDefaults(ctx context.Context, machineConfig *anywherev1.VSphereMachineConfig) error {
	templateHasSnapshot, err := d.govc.TemplateHasSnapshot(ctx, machineConfig.Spec.Template)
	if err != nil {
		return fmt.Errorf("getting template details: %v", err)
	}

	if !templateHasSnapshot {
		logger.Info("Warning: Your VM template has no snapshots. Defaulting to FullClone mode. VM provisioning might take longer.")
		if machineConfig.Spec.DiskGiB < 20 {
			logger.Info("Warning: VSphereMachineConfig DiskGiB cannot be less than 20. Defaulting to 20.")
			machineConfig.Spec.DiskGiB = 20
		}
	} else if machineConfig.Spec.DiskGiB != 25 {
		logger.Info("Warning: Your VM template includes snapshot(s). LinkedClone mode will be used. DiskGiB cannot be customizable as disks cannot be expanded when using LinkedClone mode. Using default of 25 for DiskGiBs.")
		machineConfig.Spec.DiskGiB = 25
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
