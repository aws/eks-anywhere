package cmd

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/chart/loader"
	helmRegistry "helm.sh/helm/v3/pkg/registry"
	"oras.land/oras-go/v2"
	"oras.land/oras-go/v2/registry/remote"
	"oras.land/oras-go/v2/registry/remote/auth"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/registry"
)

// copyPackagesCmd is the context for the copy packages command.
var copyPackagesCmd = &cobra.Command{
	Use:          "packages <destination-registry>",
	Short:        "Copy curated package images and charts from source registries to a destination registry",
	Long:         `Copy all the EKS Anywhere curated package images and helm charts from source registries to a destination registry. Registry credentials are fetched from docker config.`,
	SilenceUsage: true,
	RunE:         runCopyPackages,
	Args: func(cmd *cobra.Command, args []string) error {
		if err := cobra.ExactArgs(1)(cmd, args); err != nil {
			return fmt.Errorf("A destination registry must be specified as an argument")
		}
		return nil
	},
}

var cs = registry.NewCredentialStore()

func init() {
	copyCmd.AddCommand(copyPackagesCmd)

	copyPackagesCmd.Flags().StringVar(&cpc.srcImageRegistry, "src-image-registry", "", "The source registry that stores container images")
	if err := copyPackagesCmd.MarkFlagRequired("src-image-registry"); err != nil {
		logger.Fatal(err, "Cannot mark flag as required")
	}
	copyPackagesCmd.Flags().StringVar(&cpc.kubeVersion, "kube-version", "", "The kubernetes version of the package bundle to copy")
	if err := copyPackagesCmd.MarkFlagRequired("kube-version"); err != nil {
		logger.Fatal(err, "Cannot mark flag as required")
	}
	copyPackagesCmd.Flags().StringVar(&cpc.srcChartRegistry, "src-chart-registry", "", "The source registry that stores helm charts (default src-image-registry)")
	copyPackagesCmd.Flags().BoolVar(&cpc.dstPlainHTTP, "dst-plain-http", false, "Whether or not to use plain http for destination registry")
	copyPackagesCmd.Flags().BoolVar(&cpc.dstInsecure, "dst-insecure", false, "Skip TLS verification against the destination registry")
	copyPackagesCmd.Flags().BoolVar(&cpc.dryRun, "dry-run", false, "Dry run will show what artifacts would be copied, but not actually copy them")

	// making oras client to use dockerconfig
	if err := cs.Init(); err != nil {
		panic(err)
	}
	auth.DefaultClient.Credential = func(ctx context.Context, registry string) (auth.Credential, error) {
		return cs.Credential(registry)
	}
}

var cpc = copyPackagesConfig{}

var publicPackages = []string{"ecr-token-refresher", "eks-anywhere-packages", "credential-provider-package"}

// copyPackagesConfig copies packages specified in a bundle to a destination.
type copyPackagesConfig struct {
	destRegistry     string
	srcImageRegistry string
	srcChartRegistry string
	kubeVersion      string
	dstPlainHTTP     bool
	dstInsecure      bool
	dryRun           bool
}

func runCopyPackages(_ *cobra.Command, args []string) error {
	cpc.destRegistry = args[0]
	if cpc.srcChartRegistry == "" {
		cpc.srcChartRegistry = cpc.srcImageRegistry
	}
	ctx := context.Background()
	bundle, err := getPackageBundle(ctx, cpc.srcChartRegistry, cpc.kubeVersion)
	if err != nil {
		return fmt.Errorf("cannot fetch package bundle: %w", err)
	}
	if err := copyArtifacts(context.Background(), bundle); err != nil {
		return err
	}

	// copy package bundle yaml after charts and images
	tag := getPackageBundleTag(cpc.kubeVersion)
	_, err = orasCopy(ctx, curatedpackages.ImageRepositoryName, cpc.srcChartRegistry, tag, cpc.destRegistry, tag)
	return err
}

func getTagsFromChartValues(chartValues map[string]any, res map[string]string) error {
	type nodeType = map[string]interface{}

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
				}
				dfs(vv, res)
			}
		}
	}

	dfs(chartValues, res)
	return nil
}

func getPackageBundleTag(kubeVersion string) string {
	return "v" + strings.Replace(kubeVersion, ".", "-", -1) + "-latest"
}

func getPackageBundle(ctx context.Context, registry, kubeVersion string) (*packagesv1.PackageBundle, error) {
	repo, err := remote.NewRepository(registry + "/" + curatedpackages.ImageRepositoryName)
	if err != nil {
		return nil, err
	}
	tag := getPackageBundleTag(kubeVersion)
	_, data, err := oras.FetchBytes(ctx, repo, tag, oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, err
	}

	var mani ocispec.Manifest
	if err := json.Unmarshal(data, &mani); err != nil {
		return nil, fmt.Errorf("unmarshal manifest: %v", err)
	}
	if len(mani.Layers) < 1 {
		return nil, fmt.Errorf("missing layer")
	}

	_, data, err = oras.FetchBytes(ctx, repo.Blobs(), string(mani.Layers[0].Digest), oras.DefaultFetchBytesOptions)
	if err != nil {
		return nil, err
	}

	bundle := packagesv1.PackageBundle{}
	err = yaml.Unmarshal(data, &bundle)
	if err != nil {
		return nil, err
	}
	return &bundle, nil
}

func copyArtifacts(ctx context.Context, bundle *packagesv1.PackageBundle) error {
	for _, p := range bundle.Spec.Packages {
		for _, v := range p.Source.Versions {
			chartTag := v.Name
			url := cpc.srcChartRegistry + "/" + p.Source.Repository
			values, err := getChartValues(url + ":" + chartTag)
			if err != nil {
				return fmt.Errorf("cannot get chart values %s: %w", url+":"+chartTag, err)
			}

			tags := make(map[string]string)
			if err = getTagsFromChartValues(values, tags); err != nil {
				return fmt.Errorf("cannot get tags from chart values: %w", err)
			}
			_, err = orasCopy(ctx, p.Source.Repository, cpc.srcChartRegistry, chartTag, cpc.destRegistry, chartTag)
			if err != nil {
				return fmt.Errorf("cannot copy chart to repo: %w", err)
			}
			if err := copyImages(ctx, v.Images, tags); err != nil {
				return fmt.Errorf("cannot process images: %w", err)
			}
		}
	}
	return nil
}

func copyImages(ctx context.Context, images []packagesv1.VersionImages, tags map[string]string) error {
	for _, i := range images {
		dstRef := i.Digest
		if t, ok := tags[i.Digest]; ok {
			logger.V(0).Info("Using tag as the reference for digest", "tag", t, "digest", i.Digest)
			dstRef = t
		}
		_, err := orasCopy(ctx, i.Repository, cpc.srcImageRegistry, i.Digest, cpc.destRegistry, dstRef)
		if err != nil {
			return fmt.Errorf("cannot copy image to repo: %w", err)
		}
	}
	return nil
}

func getChartValues(chartURL string) (map[string]interface{}, error) {
	helmClient, err := helmRegistry.NewClient()
	if err != nil {
		return nil, err
	}
	res, err := helmClient.Pull(chartURL)
	if err != nil {
		return nil, err
	}
	chart, _ := loader.LoadArchive(bytes.NewReader(res.Chart.Data))
	return chart.Values, nil
}

func orasCopy(ctx context.Context, repo, srcRegistry, srcRef, dstRegistry, dstRef string) (ocispec.Descriptor, error) {
	logger.V(0).Info("Copying artifact", "from", srcRegistry+"/"+repo, "to", dstRegistry+"/"+repo, "dstRef", dstRef)

	if cpc.dryRun {
		return ocispec.Descriptor{}, nil
	}

	src, err := remote.NewRepository(srcRegistry + "/" + repo)
	if err != nil {
		return ocispec.Descriptor{}, err
	}

	dst, err := remote.NewRepository(dstRegistry + "/" + repo)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	setUpDstRepo(dst, &cpc)

	return oras.Copy(ctx, src, srcRef, dst, dstRef, oras.DefaultCopyOptions)
}

func setUpDstRepo(dst *remote.Repository, c *copyPackagesConfig) {
	dst.PlainHTTP = c.dstPlainHTTP
	dst.Client = &auth.Client{
		Client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: c.dstInsecure},
			},
		},
		Header: http.Header{
			"User-Agent": {"oras-go"},
		},
		Cache: auth.DefaultCache,
		Credential: func(ctx context.Context, registry string) (auth.Credential, error) {
			return cs.Credential(registry)
		},
	}
}
