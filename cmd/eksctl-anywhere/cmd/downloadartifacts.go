package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/version"
)

type downloadArtifactsOptions struct {
	downloadDir string
	fileName    string
	dryRun      bool
	retainDir   bool
}

var downloadArtifactsopts = &downloadArtifactsOptions{}

func init() {
	downloadCmd.AddCommand(downloadArtifactsCmd)
	downloadArtifactsCmd.Flags().StringVarP(&downloadArtifactsopts.fileName, "filename", "f", "", "Filename that contains EKS-A cluster configuration")
	downloadArtifactsCmd.Flags().StringVarP(&downloadArtifactsopts.downloadDir, "download-dir", "d", "eks-anywhere-downloads", "Directory to download the artifacts to")
	downloadArtifactsCmd.Flags().BoolVarP(&downloadArtifactsopts.dryRun, "dry-run", "", false, "Print the manifest URIs without downloading them")
	downloadArtifactsCmd.Flags().BoolVarP(&downloadArtifactsopts.retainDir, "retain-dir", "r", false, "Do not delete the download folder after creating a tarball")
	err := downloadArtifactsCmd.MarkFlagRequired("filename")
	if err != nil {
		log.Fatalf("Error marking filename flag as required: %v", err)
	}
}

var downloadArtifactsCmd = &cobra.Command{
	Use:          "artifacts",
	Short:        "Download EKS Anywhere artifacts/manifests to a tarball on disk",
	Long:         "This command is used to download the S3 artifacts from an EKS Anywhere bundle manifest and package them into a tarball",
	PreRunE:      preRunDownloadArtifactsCmd,
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := downloadArtifacts(cmd.Context(), downloadArtifactsopts); err != nil {
			return err
		}
		return nil
	},
}

func downloadArtifacts(context context.Context, opts *downloadArtifactsOptions) error {
	cliVersion := version.Get()
	clusterSpec, err := cluster.NewSpecFromClusterConfig(opts.fileName, cliVersion)
	if err != nil {
		return err
	}

	if clusterSpec.Cluster.Spec.RegistryMirrorConfiguration == nil || clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint == "" {
		return fmt.Errorf("endpoint not set. It is necessary to define a valid endpoint in your spec (registryMirrorConfiguration.endpoint)")
	}
	endpoint := clusterSpec.Cluster.Spec.RegistryMirrorConfiguration.Endpoint

	release, err := clusterSpec.GetRelease(cliVersion)
	if err != nil {
		return err
	}
	bundlesManifestUrl := release.BundleManifestUrl

	reader := files.NewReader(files.WithUserAgent(fmt.Sprintf("eks-a-cli-download/%s", version.Get().GitVersion)))

	// download the eks-a-release.yaml
	if !opts.dryRun {
		releaseManifestURL := clusterSpec.GetReleaseManifestUrl()
		if err := downloadArtifact(filepath.Join(opts.downloadDir, filepath.Base(releaseManifestURL)), releaseManifestURL, reader); err != nil {
			return fmt.Errorf("error downloading release manifest: %v", err)
		}
	}

	versionBundles := clusterSpec.Bundles.Spec.VersionsBundles
	for i, bundle := range versionBundles {
		for component, manifestList := range bundle.Manifests() {
			for _, manifest := range manifestList {
				if opts.dryRun {
					logger.Info(fmt.Sprintf("Found artifact: %s\n", *manifest))
					continue
				}

				filePath := filepath.Join(opts.downloadDir, bundle.KubeVersion, component, filepath.Base(*manifest))
				if err = downloadArtifact(filePath, *manifest, reader); err != nil {
					return fmt.Errorf("error downloading artifact for component %s: %v", component, err)
				}
				*manifest = filePath
			}
		}
		for component, chart := range bundle.Charts() {
			chartRegistry := fmt.Sprintf("%s/%s/%s", endpoint, chart.Name, component)
			chart.URI = fmt.Sprintf("%s:%s", chartRegistry, chart.Tag())
		}
		clusterSpec.Bundles.Spec.VersionsBundles[i] = bundle
	}

	bundleReleaseContent, err := yaml.Marshal(clusterSpec.Bundles)
	if err != nil {
		return fmt.Errorf("error marshaling bundle-release.yaml: %v", err)
	}
	bundleReleaseFilePath := filepath.Join(opts.downloadDir, filepath.Base(bundlesManifestUrl))
	if err = ioutil.WriteFile(bundleReleaseFilePath, bundleReleaseContent, 0o644); err != nil {
		return err
	}

	if !opts.dryRun {
		if err = createTarball(opts.downloadDir); err != nil {
			return err
		}

		if !opts.retainDir {
			if err = os.RemoveAll(opts.downloadDir); err != nil {
				return err
			}
		}
	}

	return nil
}

func preRunDownloadArtifactsCmd(cmd *cobra.Command, args []string) error {
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if err := viper.BindPFlag(flag.Name, flag); err != nil {
			log.Fatalf("Error initializing flags: %v", err)
		}
	})
	return nil
}

func downloadArtifact(filePath, artifactUri string, reader *files.Reader) error {
	logger.V(3).Info(fmt.Sprintf("Downloading artifact: %s", artifactUri))

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	logger.V(3).Info(fmt.Sprintf("Creating local artifact file: %s", filePath))

	contents, err := reader.ReadFile(artifactUri)
	if err != nil {
		return err
	}
	if err = ioutil.WriteFile(filePath, contents, 0o644); err != nil {
		return err
	}

	logger.V(3).Info(fmt.Sprintf("Successfully downloaded artifact %s to %s", artifactUri, filePath))

	return nil
}

func createTarball(downloadDir string) error {
	var buf bytes.Buffer
	tarFileName := fmt.Sprintf("%s.tar.gz", downloadDir)
	tarFile, err := os.Create(tarFileName)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	gzipWriter := gzip.NewWriter(&buf)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	err = filepath.Walk(downloadDir, func(file string, fileInfo os.FileInfo, walkErr error) error {
		header, err := tar.FileInfoHeader(fileInfo, file)
		if err != nil {
			return err
		}

		header.Name = filepath.ToSlash(file)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}
		if !fileInfo.IsDir() {
			data, err := os.Open(file)
			if err != nil {
				return err
			}
			if _, err := io.Copy(tarWriter, data); err != nil {
				return err
			}
			logger.V(3).Info(fmt.Sprintf("Added file %s to tarball", file))
		}
		return nil
	})
	if err != nil {
		return err
	}

	tarWriter.Close()
	gzipWriter.Close()
	if _, err = io.Copy(tarFile, &buf); err != nil {
		return err
	}
	logger.V(3).Info(fmt.Sprintf("Successfully created downloads tarball %s", tarFileName))

	return nil
}
