package addonclients

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (f *FluxAddonClient) UpdateLegacyFileStructure(ctx context.Context, currentSpec *cluster.Spec, newSpec *cluster.Spec) error {
	logger.V(1).Info("Checking for Flux repo file structure...")

	if newSpec.GitOpsConfig == nil {
		logger.V(1).Info("Skipping Flux repo file structure update, GitOps not enabled")
		return nil
	}

	ofc := &fluxForCluster{
		FluxAddonClient: f,
		clusterSpec:     currentSpec,
	}

	nfc := &fluxForCluster{
		FluxAddonClient: f,
		clusterSpec:     newSpec,
	}

	if err := nfc.syncGitRepo(ctx); err != nil {
		return err
	}

	oldClusterConfigPath := ofc.path()
	updateNeeded, err := nfc.filesUpdateNeeded(oldClusterConfigPath)
	if err != nil {
		return fmt.Errorf("updating git repo: %v", err)
	}
	if !updateNeeded {
		logger.V(1).Info("Git repo file structure is already up-to-date")
		return nil
	}

	if err := updateEksaSystemFiles(ofc, nfc); err != nil {
		return err
	}

	if err := nfc.gitOpts.Git.Add(nfc.path()); err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("adding %s to git: %v", nfc.path(), err)}
	}

	if err := nfc.FluxAddonClient.pushToRemoteRepo(ctx, nfc.path(), upgradeFluxconfigCommitMessage); err != nil {
		return err
	}
	logger.V(3).Info("Finished updating file structure to git",
		"repository", nfc.repository())

	return nil
}

func (fc *fluxForCluster) filesUpdateNeeded(oldPath string) (bool, error) {
	fluxSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.fluxSystemDir())
	eksaSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.eksaSystemDir())
	if !validations.FileExists(fluxSystemPath) {
		return false, fmt.Errorf("unrecognized file structure, missing flux-system path at %s", fluxSystemPath)
	}
	return oldPath != fc.path() || !validations.FileExists(eksaSystemPath), nil
}

func updateEksaSystemFiles(ofc, nfc *fluxForCluster) error {
	// in older version (<= 0.5.0), flux-system and eksa-system folders are under the same parent directory
	oldEksaPath := filepath.Join(ofc.path(), eksaSystemDirName)
	eksaSpec, err := nfc.updateGitOpsConfig(filepath.Join(nfc.gitOpts.Writer.Dir(), oldEksaPath, clusterConfigFileName))
	if err != nil {
		return err
	}

	w, err := nfc.initEksaWriter()
	if err != nil {
		return err
	}

	logger.V(3).Info("Updating eksa-system files...")
	if _, err = w.Write(clusterConfigFileName, eksaSpec, filewriter.PersistentFile); err != nil {
		return err
	}
	if err := ofc.generateEksaKustomizeFile(w); err != nil {
		return err
	}

	if oldEksaPath != nfc.eksaSystemDir() {
		err = nfc.gitOpts.Git.Remove(oldEksaPath)
		if err != nil {
			return &ConfigVersionControlFailedError{Err: fmt.Errorf("removing %s in git: %v", oldEksaPath, err)}
		}
	}

	return nil
}

func (fc *fluxForCluster) updateGitOpsConfig(fileName string) ([]byte, error) {
	logger.V(3).Info("Updating eks-a cluster config content - gitopsconfig clusterConfigPath...")
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	var resources [][]byte
	for _, c := range strings.Split(string(content), v1alpha1.YamlSeparator) {
		var gitopsconfig v1alpha1.GitOpsConfig
		if err := yaml.Unmarshal([]byte(c), &gitopsconfig); err != nil {
			return nil, fmt.Errorf("unable to parse %s\nyaml: %s\n %v", fileName, c, err)
		}

		if gitopsconfig.Kind() != gitopsconfig.ExpectedKind() {
			if len(c) > 0 {
				resources = append(resources, []byte(c))
			}
			continue
		}

		gitopsconfig.Spec.Flux.Github.ClusterConfigPath = fc.path()

		gitopsYaml, err := yaml.Marshal(gitopsconfig.ConvertConfigToConfigGenerateStruct())
		if err != nil {
			return nil, fmt.Errorf("outputting yaml: %v", err)
		}
		resources = append(resources, gitopsYaml)
	}

	return templater.AppendYamlResources(resources...), nil
}
