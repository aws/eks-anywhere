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

	"github.com/aws/eks-anywhere/pkg/cluster"
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
	clusterSpec, err := cluster.NewSpec(opts.fileName, cliVersion)
	if err != nil {
		return err
	}

	release, err := clusterSpec.GetRelease(cliVersion)
	if err != nil {
		return err
	}
	bundlesManifestUrl := release.BundleManifestUrl

	specManifests := []string{
		bundlesManifestUrl,
		clusterSpec.GetReleaseManifestUrl(),
	}
	for _, manifestURI := range specManifests {
		if opts.dryRun {
			logger.Info(fmt.Sprintf("Found artifact: %s\n", manifestURI))
			continue
		}
		if err = downloadArtifact("", opts.downloadDir, manifestURI, clusterSpec); err != nil {
			return fmt.Errorf("error downloading artifact: %v", err)
		}
	}

	bundle := clusterSpec.VersionsBundle
	for component, manifestList := range bundle.Manifests() {
		for _, manifest := range manifestList {
			if opts.dryRun {
				logger.Info(fmt.Sprintf("Found artifact: %s\n", manifest.URI))
				continue
			}
			if err = downloadArtifact(component, opts.downloadDir, manifest.URI, clusterSpec); err != nil {
				return fmt.Errorf("error downloading artifact for component %s: %v", component, err)
			}
		}
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

func downloadArtifact(component, downloadDir, artifactUri string, s *cluster.Spec) error {
	logger.V(3).Info(fmt.Sprintf("Downloading artifact: %s", artifactUri))

	fileName := filepath.Base(artifactUri)

	var filePath string
	if component != "" {
		filePath = filepath.Join(downloadDir, component, fileName)
	} else {
		filePath = filepath.Join(downloadDir, fileName)
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}

	logger.V(3).Info(fmt.Sprintf("Creating local artifact file: %s", filePath))

	contents, err := s.ReadFile(artifactUri)
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

	if _, err = io.Copy(tarFile, &buf); err != nil {
		return err
	}
	logger.V(3).Info(fmt.Sprintf("Successfully created downloads tarball %s", tarFileName))

	return nil
}
