package oras

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type BundlePusher interface {
	PushBundle(ctx context.Context, ref, fileName string, fileContent []byte) error
}
type FileRegistryImporter struct {
	registry           string
	username, password string
	srcFolder          string
	bundlePusher       BundlePusher
}

func NewFileRegistryImporter(registry, username, password, srcFolder string, bundlePusher BundlePusher) *FileRegistryImporter {
	return &FileRegistryImporter{
		registry:     registry,
		username:     username,
		password:     password,
		srcFolder:    srcFolder,
		bundlePusher: bundlePusher,
	}
}

func (fr *FileRegistryImporter) Push(ctx context.Context, bundles *releasev1.Bundles) {
	artifacts := ReadFilesFromBundles(bundles)
	for _, a := range UniqueCharts(artifacts) {
		updatedChartURL := urls.ReplaceHost(a, fr.registry)
		fileName := ChartFileName(a)
		chartFilepath := filepath.Join(fr.srcFolder, fileName)
		data, err := os.ReadFile(chartFilepath)
		if err != nil {
			logger.Info("Warning: reading file", "error", err)
			continue
		}
		err = fr.bundlePusher.PushBundle(ctx, updatedChartURL, fileName, data)
		if err != nil {
			logger.Info("Warning: Failed  to push to registry", "error", err)
		}
	}
}

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".yaml"
}
