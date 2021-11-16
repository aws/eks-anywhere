package addonclients

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/types"
	"sigs.k8s.io/yaml"
)

const upgradeFluxconfigCommitMessage = "Upgrade commit of flux configuration; generated by EKS-A CLI"

func (f *FluxAddonClient) Upgrade(ctx context.Context, managementCluster *types.Cluster, currentSpec *cluster.Spec, newSpec *cluster.Spec) (*types.ChangeDiff, error) {
	logger.V(1).Info("Checking for Flux upgrades")
	if !newSpec.Cluster.IsSelfManaged() {
		logger.V(1).Info("Skipping Flux upgrades, not a self-managed cluster")
		return nil, nil
	}

	if newSpec.GitOpsConfig == nil {
		logger.V(1).Info("Skipping Flux upgrades, GitOps not enabled")
		return nil, nil
	}

	changeDiff := f.fluxChangeDiff(currentSpec, newSpec)
	if changeDiff == nil {
		logger.V(1).Info("Nothing to upgrade for Flux")
		return nil, nil
	}

	logger.V(1).Info("Starting Flux upgrades")
	if err := f.upgradeFilesAndCommit(ctx, newSpec); err != nil {
		return nil, fmt.Errorf("failed upgrading Flux from bundles %d to bundles %d: %v", currentSpec.Bundles.Spec.Number, newSpec.Bundles.Spec.Number, err)
	}
	if err := f.flux.BootstrapToolkitsComponents(ctx, managementCluster, newSpec.GitOpsConfig); err != nil {
		return nil, fmt.Errorf("failed upgrading Flux components: %v", err)
	}
	if err := f.flux.Reconcile(ctx, managementCluster, newSpec.GitOpsConfig); err != nil {
		return nil, fmt.Errorf("failed reconciling Flux components: %v", err)
	}

	return types.NewChangeDiff(changeDiff), nil
}

func (f *FluxAddonClient) fluxChangeDiff(currentSpec, newSpec *cluster.Spec) *types.ComponentChangeDiff {
	oldVersion := currentSpec.VersionsBundle.Flux.Version
	newVersion := newSpec.VersionsBundle.Flux.Version
	if oldVersion != newVersion {
		logger.V(1).Info("Flux change diff ", "oldVersion", oldVersion, "newVersion", newVersion)
		return &types.ComponentChangeDiff{
			ComponentName: "Flux",
			NewVersion:    newVersion,
			OldVersion:    oldVersion,
		}

	}
	return nil
}

func (f *FluxAddonClient) upgradeFilesAndCommit(ctx context.Context, newSpec *cluster.Spec) error {
	fc := &fluxForCluster{
		FluxAddonClient: f,
		clusterSpec:     newSpec,
	}

	if err := fc.syncGitRepo(ctx); err != nil {
		return err
	}

	if err := fc.commitFluxUpgradeFilesToGit(ctx); err != nil {
		return err
	}

	return nil
}

func (fc *fluxForCluster) commitFluxUpgradeFilesToGit(ctx context.Context) error {
	logger.Info("Adding flux configuration files to Git")

	logger.V(3).Info("Generating eks-a cluster config file...")
	if err := fc.writeEksaUpgradeFiles(); err != nil {
		return err
	}

	logger.V(3).Info("Generating flux custom manifest files...")
	if err := fc.writeFluxUpgradeFiles(); err != nil {
		return err
	}

	if err := fc.gitOpts.Git.Add(fc.path()); err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when adding %s to git: %v", fc.path(), err)}
	}

	if err := fc.FluxAddonClient.pushToRemoteRepo(ctx, fc.path(), upgradeFluxconfigCommitMessage); err != nil {
		return err
	}
	logger.V(3).Info("Finished pushing flux custom manifest files to git",
		"repository", fc.repository())
	return nil
}

func (fc *fluxForCluster) writeFluxUpgradeFiles() error {
	w, err := fc.initFluxWriter()
	if err != nil {
		return err
	}

	t := templater.New(w)
	logger.V(3).Info("Generating flux-system patch file...")
	if err = fc.generateFluxPatchFile(t); err != nil {
		return err
	}

	return nil
}

func (fc *fluxForCluster) writeEksaUpgradeFiles() error {
	eksaSpec, err := fc.generateUpdatedEksaConfig(filepath.Join(fc.gitOpts.Writer.Dir(), fc.eksaSystemDir(), clusterConfigFileName))
	if err != nil {
		return err
	}

	w, err := fc.initEksaWriter()
	if err != nil {
		return err
	}

	logger.V(3).Info("Updating eksa-system eksa-cluster.yaml")
	if _, err = w.Write(clusterConfigFileName, eksaSpec, filewriter.PersistentFile); err != nil {
		return err
	}
	return nil
}

func (fc *fluxForCluster) generateUpdatedEksaConfig(fileName string) ([]byte, error) {
	logger.V(3).Info("Updating eks-a cluster config content")
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	var resources [][]byte
	for _, c := range strings.Split(string(content), v1alpha1.YamlSeparator) {
		var resource unstructured.Unstructured
		if err := yaml.Unmarshal([]byte(c), &resource); err != nil {
			return nil, fmt.Errorf("unable to parse %s\nyaml: %s\n %v", fileName, c, err)
		}
		if resource.GetKind() == "" {
			continue
		}
		if resource.GetNamespace() == "" {
			logger.V(4).Info("Namespace is not presented, set to default", "resource", resource.GetKind())
			resource.SetNamespace("default")
		}
		resourceYaml, err := yaml.Marshal(resource.Object)
		if err != nil {
			return nil, fmt.Errorf("error outputting yaml: %v", err)
		}
		resources = append(resources, resourceYaml)
	}

	return templater.AppendYamlResources(resources...), nil
}
