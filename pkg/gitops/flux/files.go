package flux

import (
	_ "embed"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clustermarshaller"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	eksaSystemDirName     = "eksa-system"
	kustomizeFileName     = "kustomization.yaml"
	clusterConfigFileName = "eksa-cluster.yaml"
	fluxSyncFileName      = "gotk-sync.yaml"
	fluxPatchFileName     = "gotk-patches.yaml"
)

//go:embed manifests/eksa-system/kustomization.yaml
var eksaKustomizeContent string

//go:embed manifests/flux-system/kustomization.yaml
var fluxKustomizeContent string

//go:embed manifests/flux-system/gotk-sync.yaml
var fluxSyncContent string

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

func (g *FileGenerator) WriteEksaFiles(clusterSpec *cluster.Spec, datacenterConfig providers.DatacenterConfig, machineConfigs []providers.MachineConfig) error {
	if datacenterConfig == nil && machineConfigs == nil {
		return nil
	}

	if err := g.WriteClusterConfig(clusterSpec, datacenterConfig, machineConfigs); err != nil {
		return err
	}

	if err := g.WriteEksaKustomization(); err != nil {
		return err
	}

	return nil
}

// WriteFluxSystemFiles writes the flux-system files into the flux system git directory.
func (g *FileGenerator) WriteFluxSystemFiles(managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec) error {
	if err := g.WriteFluxKustomization(managementComponents, clusterSpec); err != nil {
		return err
	}

	if err := g.WriteFluxSync(); err != nil {
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

func (g *FileGenerator) WriteEksaKustomization() error {
	values := map[string]string{
		"ConfigFileName": clusterConfigFileName,
	}

	if path, err := g.eksaTemplater.WriteToFile(eksaKustomizeContent, values, kustomizeFileName, filewriter.PersistentFile); err != nil {
		return fmt.Errorf("writing eks-a kustomization manifest file into %s: %v", path, err)
	}
	return nil
}

// WriteFluxKustomization writes the flux-system kustomization file into the flux system git directory.
func (g *FileGenerator) WriteFluxKustomization(managementComponents *cluster.ManagementComponents, clusterSpec *cluster.Spec) error {
	values := map[string]string{
		"Namespace":                   clusterSpec.FluxConfig.Spec.SystemNamespace,
		"SourceControllerImage":       managementComponents.Flux.SourceController.VersionedImage(),
		"KustomizeControllerImage":    managementComponents.Flux.KustomizeController.VersionedImage(),
		"HelmControllerImage":         managementComponents.Flux.HelmController.VersionedImage(),
		"NotificationControllerImage": managementComponents.Flux.NotificationController.VersionedImage(),
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
