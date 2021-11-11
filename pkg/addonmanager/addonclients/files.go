package addonclients

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/templater"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func (f *FluxAddonClient) UpdateLegacyFileStructure(ctx context.Context, spec *cluster.Spec) error {
	logger.V(1).Info("Checking for Flux repo file structure...")

	if spec.GitOpsConfig == nil {
		logger.V(1).Info("Skipping Flux repo file structure update, GitOps not enabled")
		return nil
	}

	fc := &fluxForCluster{
		FluxAddonClient: f,
		clusterSpec:     spec,
	}

	if err := fc.syncGitRepo(ctx); err != nil {
		return err
	}

	updateNeeded, err := fc.filesUpdateNeeded()
	if err != nil {
		return fmt.Errorf("error updating git repo: %v", err)
	}
	if !updateNeeded {
		logger.V(1).Info("Git repo file structure is already up-to-date")
		return nil
	}

	if err := fc.updateEksaSystemFiles(); err != nil {
		return err
	}

	if err := fc.gitOpts.Git.Add(fc.path()); err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when adding %s to git: %v", fc.path(), err)}
	}

	if err := fc.FluxAddonClient.pushToRemoteRepo(ctx, fc.path(), upgradeFluxconfigCommitMessage); err != nil {
		return err
	}
	logger.V(3).Info("Finished updating file structure to git",
		"repository", fc.repository())

	return nil
}

func (fc *fluxForCluster) filesUpdateNeeded() (bool, error) {
	fluxSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.fluxSystemDir())
	eksaSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.eksaSystemDir())
	if !validations.FileExists(fluxSystemPath) {
		return false, fmt.Errorf("unrecognized file structure, missing flux-system path at %s", fluxSystemPath)
	}
	return !validations.FileExists(eksaSystemPath), nil
}

func (fc *fluxForCluster) updateEksaSystemFiles() error {
	// in oler version (<= 0.5.0), flux-system and eksa-system folders are under the same parent directory
	oldEksaPath := filepath.Join(fc.path(), eksaSystemDirName)
	content, err := ioutil.ReadFile(filepath.Join(fc.gitOpts.Writer.Dir(), oldEksaPath, clusterConfigFileName))
	if err != nil {
		return fmt.Errorf("unable to read eksa-cluster.yaml file due to: %v", err)
	}

	if err := fc.writeEksaUpgradeFiles(content); err != nil {
		return err
	}

	err = fc.gitOpts.Git.Remove(oldEksaPath)
	if err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when removing %s in git: %v", oldEksaPath, err)}
	}
	return nil
}

func (fc *fluxForCluster) writeEksaUpgradeFiles(resourcesSpec []byte) error {
	w, err := fc.initEksaWriter()
	if err != nil {
		return err
	}

	t := templater.New(w)
	logger.V(3).Info("Updating eksa-system cluster config file...")

	if _, err = t.WriteBytesToFile(resourcesSpec, clusterConfigFileName, filewriter.PersistentFile); err != nil {
		return err
	}

	return fc.generateEksaKustomizeFile()
}
