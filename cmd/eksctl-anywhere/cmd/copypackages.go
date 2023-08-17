package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/aws/eks-anywhere/pkg/config"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/registry"
)

// copyPackagesCmd is the context for the copy packages command.
var copyPackagesCmd = &cobra.Command{
	Use:          "packages <destination-registry>",
	Short:        "Copy curated package images and charts from a source to a destination",
	Long:         `Copy all the EKS Anywhere curated package images and helm charts from a source to a destination.`,
	SilenceUsage: true,
	RunE:         runCopyPackages,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("A destination must be specified as an argument")
		}
		return nil
	},
}

func init() {
	copyCmd.AddCommand(copyPackagesCmd)

	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.bundleFile, "bundle", "b", "", "EKS-A bundle file to read artifact dependencies from")
	if err := copyPackagesCmd.MarkFlagRequired("bundle"); err != nil {
		log.Fatalf("Cannot mark 'bundle' flag as required: %s", err)
	}
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.dstCert, "dst-cert", "", "", "TLS certificate for destination registry")
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.srcCert, "src-cert", "", "", "TLS certificate for source registry")
	copyPackagesCmd.Flags().BoolVar(&copyPackagesCommand.insecure, "insecure", false, "Skip TLS verification while copying images and charts")
	copyPackagesCmd.Flags().BoolVar(&copyPackagesCommand.dryRun, "dry-run", false, "Dry run copy to print images that would be copied")
	copyPackagesCmd.Flags().StringVarP(&copyPackagesCommand.awsRegion, "aws-region", "", os.Getenv(config.EksaRegionEnv), "Region to copy images from")
}

var copyPackagesCommand = CopyPackagesCommand{}

var publicPackages = []string{"ecr-token-refresher", "eks-anywhere-packages", "credential-provider-package"}

// CopyPackagesCommand copies packages specified in a bundle to a destination.
type CopyPackagesCommand struct {
	destination   string
	bundleFile    string
	srcCert       string
	dstCert       string
	insecure      bool
	dryRun        bool
	awsRegion     string
	registryCache *registry.Cache
}

func runCopyPackages(_ *cobra.Command, args []string) error {
	ctx := context.Background()
	copyPackagesCommand.destination = args[0]

	credentialStore := registry.NewCredentialStore()
	err := credentialStore.Init()
	if err != nil {
		return err
	}

	return copyPackagesCommand.call(ctx, credentialStore)
}

func (c CopyPackagesCommand) call(ctx context.Context, credentialStore *registry.CredentialStore) error {
	factory := dependencies.NewFactory()
	deps, err := factory.
		WithExecutableImage().
		WithManifestReader().
		WithHelm().
		Build(ctx)
	if err != nil {
		return err
	}
	defer deps.Close(ctx)

	eksaBundle, err := bundles.Read(deps.ManifestReader, c.bundleFile)
	if err != nil {
		return err
	}

	c.registryCache = registry.NewCache()
	bundleReader := curatedpackages.NewPackageReader(c.registryCache, credentialStore, c.awsRegion)

	// Note: package bundle yaml file is included by charts below
	charts := bundleReader.ReadChartsFromBundles(ctx, eksaBundle)
	tags, err := getTagsFromCharts(ctx, deps.Helm, charts)
	if err != nil {
		return err
	}

	certificates, err := registry.GetCertificates(c.dstCert)
	if err != nil {
		return err
	}

	dstContext := registry.NewStorageContext(c.destination, credentialStore, certificates, c.insecure)
	dstRegistry, err := c.registryCache.Get(dstContext)
	if err != nil {
		return fmt.Errorf("error with repository %s: %v", c.destination, err)
	}

	logger.V(0).Info("Copying curated packages helm charts from public ECR to destination", "destination", c.destination)
	if err = c.copyArtifacts(ctx, func(a registry.Artifact) registry.StorageClient {
		return dstRegistry
	}, credentialStore, charts); err != nil {
		return err
	}

	imageList, err := bundleReader.ReadImagesFromBundles(ctx, eksaBundle)
	if err != nil {
		return err
	}
	addTags(imageList, tags)

	logger.V(0).Info("Copying curated packages images from private ECR to destination", "destination", c.destination)
	publicRepoPrefix := strings.Split(charts[0].Repository, "/")[0] + "/"
	return c.copyArtifacts(ctx, func(a registry.Artifact) registry.StorageClient {
		// private curated packages should go to curated-packages
		dstRegistry.SetProject("curated-packages/")
		for _, pp := range publicPackages {
			if strings.HasSuffix(a.Repository, pp) {
				// public curated packages images should go to publicRepo/*
				dstRegistry.SetProject(publicRepoPrefix)
			}
		}
		return dstRegistry
	}, credentialStore, imageList)
}

func addTags(images []registry.Artifact, tags map[string]string) {
	for idx, i := range images {
		if tag, ok := tags[i.Digest]; ok {
			images[idx].Tag = tag
		}
	}
}

func (c CopyPackagesCommand) copyArtifacts(ctx context.Context, getDstRegistry func(registry.Artifact) registry.StorageClient, credentialStore *registry.CredentialStore, artifacts []registry.Artifact) error {
	certificates, err := registry.GetCertificates(c.srcCert)
	if err != nil {
		return err
	}

	for _, a := range artifacts {
		host := a.Registry

		srcContext := registry.NewStorageContext(host, credentialStore, certificates, c.insecure)
		srcRegistry, err := c.registryCache.Get(srcContext)
		if err != nil {
			return fmt.Errorf("error with repository %s: %v", host, err)
		}

		dstRegistry := getDstRegistry(a)

		artifact := registry.NewArtifact(a.Registry, a.Repository, a.Tag, a.Digest)
		logger.V(0).Info("Copying image to destination", "destination", dstRegistry.Destination(artifact))
		if c.dryRun {
			continue
		}

		err = registry.Copy(ctx, srcRegistry, dstRegistry, artifact)
		if err != nil {
			return err
		}
	}
	return nil
}

// Since package bundle doesn't contain tags, we get tags from chart values.
func getTagsFromCharts(ctx context.Context, helm *executables.Helm, charts []registry.Artifact) (map[string]string, error) {
	tags := make(map[string]string)
	for _, chart := range charts {
		for _, pp := range publicPackages {
			// only public package charts may contain tags info
			if strings.HasSuffix(chart.Repository, "/"+pp) {
				url := "oci://" + chart.Registry + "/" + chart.Repository
				values, err := helm.ShowValues(ctx, url, chart.Tag)
				if err != nil {
					return nil, err
				}

				err = getTagsFromChartValues(values.Bytes(), tags)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return tags, nil
}

func getTagsFromChartValues(chartValues []byte, res map[string]string) error {
	type nodeType = map[interface{}]interface{}
	m := make(nodeType)

	if err := yaml.Unmarshal(chartValues, &m); err != nil {
		return err
	}

	var dfs func(nodeType, map[string]string)
	dfs = func(node nodeType, res map[string]string) {
		for _, v := range node {
			switch vv := v.(type) {
			case nodeType:
				_, hasTag := vv["tag"]
				_, hasDigest := vv["digest"]
				if hasTag && hasDigest && vv["tag"] != "" && vv["digest"] != "" && vv["tag"] != nil {
					if strings.HasPrefix(vv["tag"].(string), "sha256:") {
						continue
					}
					res[vv["digest"].(string)] = vv["tag"].(string)
				} else {
					dfs(vv, res)
				}
			}
		}
	}

	dfs(m, res)
	return nil
}
