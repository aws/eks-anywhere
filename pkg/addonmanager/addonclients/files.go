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

	if !nfc.filesUpdateNeeded() {
		logger.V(1).Info("Git repo file structure is already up-to-date")
		return nil
	}

	if err := nfc.updateEksaSystemFiles(ofc.eksaSystemDir()); err != nil {
		return err
	}

	if err := nfc.moveFluxSystemFiles(filepath.Join(ofc.path(), ofc.namespace())); err != nil {
		return err
	}

	p := nfc.clusterSpec.GitOpsConfig.Spec.Flux.Github.ClusterRootPath()
	if err := nfc.gitOpts.Git.Add(p); err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when adding %s to git: %v", p, err)}
	}

	if err := nfc.FluxAddonClient.pushToRemoteRepo(ctx, p, upgradeFluxconfigCommitMessage); err != nil {
		return err
	}
	logger.V(3).Info("Finished updating file structure to git",
		"repository", nfc.repository())

	return nil
}

func (fc *fluxForCluster) filesUpdateNeeded() bool {
	fluxSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.fluxSystemDir())
	eksaSystemPath := filepath.Join(fc.gitOpts.Writer.Dir(), fc.eksaSystemDir())
	return !(validations.FileExists(fluxSystemPath) && validations.FileExists(eksaSystemPath))
}

func (fc *fluxForCluster) moveFluxSystemFiles(oldPath string) error {
	if oldPath == fc.fluxSystemDir() {
		logger.V(3).Info("Directory [flux-system] is already up-to-date")
		return nil
	}

	w := fc.gitOpts.Writer

	if err := w.Copy(filepath.Join(oldPath, kustomizeFileName), filepath.Join(fc.fluxSystemDir(), kustomizeFileName)); err != nil {
		return err
	}

	if err := w.Copy(filepath.Join(oldPath, fluxSyncFileName), filepath.Join(fc.fluxSystemDir(), fluxSyncFileName)); err != nil {
		return err
	}

	if err := w.Copy(filepath.Join(oldPath, fluxPatchFileName), filepath.Join(fc.fluxSystemDir(), fluxPatchFileName)); err != nil {
		return err
	}

	if err := w.Copy(filepath.Join(oldPath, fluxComponentsFileName), filepath.Join(fc.fluxSystemDir(), fluxComponentsFileName)); err != nil {
		return err
	}

	err := fc.gitOpts.Git.Remove(oldPath)
	if err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when removing %s in git: %v", oldPath, err)}
	}
	return nil
}

func (fc *fluxForCluster) updateEksaSystemFiles(oldPath string) error {
	if oldPath == fc.eksaSystemDir() {
		logger.V(3).Info("Directory [eksa-system] is already up-to-date")
		return nil
	}
	eksaSpec, err := fc.updateGitOpsConfig(filepath.Join(fc.gitOpts.Writer.Dir(), oldPath, clusterConfigFileName))
	if err != nil {
		return err
	}

	if err := fc.writeEksaUpgradeFiles(eksaSpec); err != nil {
		return err
	}

	err = fc.gitOpts.Git.Remove(oldPath)
	if err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when removing %s in git: %v", oldPath, err)}
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

func (fc *fluxForCluster) updateGitOpsConfig(fileName string) ([]byte, error) {
	logger.V(3).Info("Updating eks-a cluster config content - gitopsconfig clusterConfigPath...")
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("unable to read file due to: %v", err)
	}

	var resources [][]byte
	for _, c := range strings.Split(string(content), v1alpha1.YamlSeparator) {
		var gitopsconfig v1alpha1.GitOpsConfig
		if err = yaml.Unmarshal([]byte(c), &gitopsconfig); err != nil {
			return nil, fmt.Errorf("unable to parse %s\nyaml: %s\n %v", fileName, c, err)
		}

		if gitopsconfig.Kind() != gitopsconfig.ExpectedKind() {
			if len(c) > 0 {
				resources = append(resources, []byte(c))
			}
			continue
		}

		if err := yaml.UnmarshalStrict([]byte(c), gitopsconfig); err != nil {
			return nil, err
		}

		gitopsconfig.Spec.Flux.Github.ClusterConfigPath = fc.path()

		gitopsYaml, err := yaml.Marshal(gitopsconfig.ConvertConfigToConfigGenerateStruct())
		if err != nil {
			return nil, fmt.Errorf("error outputting yaml: %v", err)
		}
		resources = append(resources, gitopsYaml)
	}

	return templater.AppendYamlResources(resources...), nil
}
