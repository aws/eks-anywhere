package oras

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/urls"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type FileRegistryImporter struct {
	registry           string
	username, password string
	srcFolder          string
}

func NewFileRegistryImporter(registry, username, password, srcFolder string) *FileRegistryImporter {
	return &FileRegistryImporter{
		registry:  registry,
		username:  username,
		password:  password,
		srcFolder: srcFolder,
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
		err = curatedpackages.PushBundle(ctx, updatedChartURL, fileName, data)
		if err != nil {
			logger.Info("Warning: Failed  to push to registry", "error", err)
		}
	}
}

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".yaml"
}
