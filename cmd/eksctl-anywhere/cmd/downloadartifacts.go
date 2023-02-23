package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/manifests/bundles"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	"github.com/aws/eks-anywhere/pkg/version"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type downloadArtifactsOptions struct {
	downloadDir     string
	fileName        string
	bundlesOverride string
	dryRun          bool
	retainDir       bool
}

var downloadArtifactsopts = &downloadArtifactsOptions{}

func init() {
	downloadCmd.AddCommand(downloadArtifactsCmd)
	downloadArtifactsCmd.Flags().StringVarP(&downloadArtifactsopts.bundlesOverride, "bundles-override", "", "", "Override default Bundles manifest (not recommended)")
	downloadArtifactsCmd.Flags().StringVarP(&downloadArtifactsopts.fileName, "filename", "f", "", "[Deprecated] Filename that contains EKS-A cluster configuration")
	downloadArtifactsCmd.Flags().StringVarP(&downloadArtifactsopts.downloadDir, "download-dir", "d", "eks-anywhere-downloads", "Directory to download the artifacts to")
	downloadArtifactsCmd.Flags().BoolVarP(&downloadArtifactsopts.dryRun, "dry-run", "", false, "Print the manifest URIs without downloading them")
	downloadArtifactsCmd.Flags().BoolVarP(&downloadArtifactsopts.retainDir, "retain-dir", "r", false, "Do not delete the download folder after creating a tarball")
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
	factory := dependencies.NewFactory()
	deps, err := factory.
		WithFileReader().
		WithManifestReader().
		Build(context)
	if err != nil {
		return err
	}

	reader := deps.FileReader

	var b *releasev1.Bundles
	if opts.bundlesOverride != "" {
		b, err = bundles.Read(reader, opts.bundlesOverride)
		if err != nil {
			return err
		}
	} else {
		b, err = deps.ManifestReader.ReadBundlesForVersion(version.Get().GitVersion)
		if err != nil {
			return err
		}
	}

	// download the eks-a-release.yaml
	if !opts.dryRun {
		releaseManifestURL := releases.ManifestURL()
		if err := downloadArtifact(filepath.Join(opts.downloadDir, filepath.Base(releaseManifestURL)), releaseManifestURL, reader); err != nil {
			return fmt.Errorf("downloading release manifest: %v", err)
		}
	}

	versionBundles := b.Spec.VersionsBundles
	for i, bundle := range versionBundles {
		for component, manifestList := range bundle.Manifests() {
			for _, manifest := range manifestList {
				if *manifest == "" {
					// This can happen if the provider is not GA and not added to the bundle-release corresponding to an EKS-A release
					continue
				}
				if opts.dryRun {
					logger.Info(fmt.Sprintf("Found artifact: %s\n", *manifest))
					continue
				}

				filePath := filepath.Join(opts.downloadDir, bundle.KubeVersion, component, filepath.Base(*manifest))
				if err = downloadArtifact(filePath, *manifest, reader); err != nil {
					return fmt.Errorf("downloading artifact for component %s: %v", component, err)
				}
				*manifest = filePath
			}
		}
		b.Spec.VersionsBundles[i] = bundle
	}

	bundleReleaseContent, err := yaml.Marshal(b)
	if err != nil {
		return fmt.Errorf("marshaling bundle-release.yaml: %v", err)
	}
	bundleReleaseFilePath := filepath.Join(opts.downloadDir, "bundle-release.yaml")
	if err = os.WriteFile(bundleReleaseFilePath, bundleReleaseContent, 0o644); err != nil {
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
	if err = os.WriteFile(filePath, contents, 0o644); err != nil {
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
