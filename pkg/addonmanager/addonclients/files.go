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

	if !nfc.filesUpdateNeeded() {
		logger.V(1).Info("Git repo file structure is already up-to-date")
		return nil
	}

	if err := nfc.syncGitRepo(ctx); err != nil {
		return err
	}

	eksaSpec, err := nfc.updateGitOpsConfig(filepath.Join(ofc.gitOpts.Writer.Dir(), ofc.eksaSystemDir(), clusterConfigFileName))
	if err != nil {
		return err
	}

	err = ofc.gitOpts.Git.Remove(ofc.path())
	if err != nil {
		return &ConfigVersionControlFailedError{Err: fmt.Errorf("error when removing %s in git: %v", ofc.path(), err)}
	}

	if err := nfc.writeEksaUpgradeFiles(eksaSpec); err != nil {
		return err
	}

	nWriter, err := nfc.initFluxWriter()
	if err != nil {
		return err
	}

	if err = ofc.generateFluxSystemFiles(nWriter); err != nil {
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

		gitopsYaml, err := yaml.Marshal(gitopsconfig)
		if err != nil {
			return nil, fmt.Errorf("error outputting yaml: %v", err)
		}
		resources = append(resources, gitopsYaml)
	}

	return templater.AppendYamlResources(resources...), nil
}
