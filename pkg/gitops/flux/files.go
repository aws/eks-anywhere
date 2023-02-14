package flux

import (
	_ "embed"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	eksaSystemDirName     = "eksa-system"
	kustomizeFileName     = "kustomization.yaml"
	clusterConfigFileName = "eksa-cluster.yaml"
	hardwareFileName      = "hardware.yaml"
	fluxSyncFileName      = "gotk-sync.yaml"
	fluxPatchFileName     = "gotk-patches.yaml"
)

//go:embed manifests/eksa-system/kustomization.yaml
var eksaKustomizeContent string

//go:embed manifests/flux-system/kustomization.yaml
var fluxKustomizeContent string

//go:embed manifests/flux-system/gotk-sync.yaml
var fluxSyncContent string

//go:embed manifests/flux-system/gotk-patches.yaml
var fluxPatchContent string

type Templater interface {
	WriteToFile(templateContent string, data interface{}, fileName string, f ...filewriter.FileOptionsFunc) (filePath string, err error)
}

type FileGenerator struct {
	fluxWriter, eksaWriter       filewriter.FileWriter
	fluxTemplater, eksaTemplater Templater
}

func NewFileGenerator() *FileGenerator {
	return &FileGenerator{}
}

// NewFileGeneratorWithWriterTemplater takes flux and eksa writer and templater interface to build the generator.
// This is only for testing.
func NewFileGeneratorWithWriterTemplater(fluxWriter, eksaWriter filewriter.FileWriter, fluxTemplater, eksaTemplater Templater) *FileGenerator {
	return &FileGenerator{
		fluxWriter:    fluxWriter,
		eksaWriter:    eksaWriter,
		fluxTemplater: fluxTemplater,
		eksaTemplater: eksaTemplater,
	}
}

func (g *FileGenerator) Init(writer filewriter.FileWriter, eksaSystemDir, fluxSystemDir string) error {
	eksaWriter, err := writer.WithDir(eksaSystemDir)
	if err != nil {
		return fmt.Errorf("initializing eks-a system writer: %v", err)
	}
	eksaWriter.CleanUpTemp()

	fluxWriter, err := writer.WithDir(fluxSystemDir)
	if err != nil {
		return fmt.Errorf("initializing flux system writer: %v", err)
	}
	fluxWriter.CleanUpTemp()

	g.eksaWriter = eksaWriter
	g.fluxWriter = fluxWriter
	g.eksaTemplater = templater.New(eksaWriter)
	g.fluxTemplater = templater.New(fluxWriter)

	return nil
}

// WriteEksaFiles writes the files defining objects in the eksa-system namespace for the cluster to the git repo.
func (g *FileGenerator) WriteEksaFiles(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig, hardwareCSVPath string) error {
	if datacenterConfig == nil && machineConfigs == nil {
		return nil
	}

	if err := g.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs); err != nil {
		return err
	}

	kuztomizationValues := map[string]string{
		"ConfigFileName": clusterConfigFileName,
	}

	if hardwareCSVPath != "" {
		if err := g.WriteHardwareYAML(hardwareCSVPath); err != nil {
			return err
		}

		// Allow edits to a hardware.yaml file to be recognized by flux by
		// adding it to resources in the kuztomization.yaml file
		kuztomizationValues["HardwareFileName"] = hardwareFileName
	}

	if err := g.WriteEksaKustomization(kuztomizationValues); err != nil {
		return err
	}

	return nil
}

func (g *FileGenerator) WriteFluxSystemFiles(clusterSpec *cluster.Spec) error {
	if err := g.WriteFluxKustomization(clusterSpec); err != nil {
		return err
	}

	if err := g.WriteFluxSync(); err != nil {
		return err
	}

	if err := g.WriteFluxPatch(clusterSpec); err != nil {
		return err
	}

	return nil
}

func (g *FileGenerator) WriteClusterConfig(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error {
	specs, err := clustermarshaller.MarshalClusterSpec(clusterSpec, datacenterConfig, machineConfigs)
	if err != nil {
		return err
	}
	if filePath, err := g.eksaWriter.Write(clusterConfigFileName, specs, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing eks-a cluster config file into %s: %v", filePath, err)
	}

	return nil
}

// WriteEksaKustomization writes the eks-a kustomization manifest to the eksa-system folder in the repository path.
func (g *FileGenerator) WriteEksaKustomization(values map[string]string) error {
	if path, err := g.eksaTemplater.WriteToFile(eksaKustomizeContent, values, kustomizeFileName, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing eks-a kustomization manifest file into %s: %v", path, err)
	}
	return nil
}

// WriteHardwareYAML writes the hardware manifest kustomization manifest to the eksa-system folder in the repository path.
func (g *FileGenerator) WriteHardwareYAML(hardwareCSVPath string) error {
	hardwareSpec, err := hardware.BuildHardwareYAML(hardwareCSVPath)
	if err != nil {
		return fmt.Errorf("building hardware manifest from csv file %s: %v", hardwareCSVPath, err)
	}
	if filePath, err := g.eksaWriter.Write(hardwareFileName, hardwareSpec, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing eks-a hardware manifest file into %s: %v", filePath, err)
	}
	return nil
}

func (g *FileGenerator) WriteFluxKustomization(clusterSpec *cluster.Spec) error {
	values := map[string]string{
		"Namespace": clusterSpec.FluxConfig.Spec.SystemNamespace,
	}

	if path, err := g.fluxTemplater.WriteToFile(fluxKustomizeContent, values, kustomizeFileName, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("creating flux-system kustomization manifest file into %s: %v", path, err)
	}
	return nil
}

func (g *FileGenerator) WriteFluxSync() error {
	if path, err := g.fluxTemplater.WriteToFile(fluxSyncContent, nil, fluxSyncFileName, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("creating flux-system sync manifest file into %s: %v", path, err)
	}
	return nil
}

func (g *FileGenerator) WriteFluxPatch(clusterSpec *cluster.Spec) error {
	values := map[string]string{
		"Namespace":                   clusterSpec.FluxConfig.Spec.SystemNamespace,
		"SourceControllerImage":       clusterSpec.VersionsBundle.Flux.SourceController.VersionedImage(),
		"KustomizeControllerImage":    clusterSpec.VersionsBundle.Flux.KustomizeController.VersionedImage(),
		"HelmControllerImage":         clusterSpec.VersionsBundle.Flux.HelmController.VersionedImage(),
		"NotificationControllerImage": clusterSpec.VersionsBundle.Flux.NotificationController.VersionedImage(),
	}
	if path, err := g.fluxTemplater.WriteToFile(fluxPatchContent, values, fluxPatchFileName, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("creating flux-system patch manifest file into %s: %v", path, err)
	}
	return nil
}
