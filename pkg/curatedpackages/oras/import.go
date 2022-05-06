package oras

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/logger"
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
		chartName := filepath.Base(a)
		fileName := ChartFileName(a)
		chartFilepath := filepath.Join(fr.srcFolder, fileName)
		data, err := os.ReadFile(chartFilepath)
		if err != nil {
			logger.MarkFail("Error reading file", chartFilepath, "error", err)
			continue
		}
		ref := fmt.Sprintf("%s/%s", fr.registry, chartName)
		err = curatedpackages.Push(ctx, a, ref, fileName, data)
		if err != nil {
			logger.MarkFail("Failed  to push registry", ref, "error", err)
		}
	}
}

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".yaml"
}
