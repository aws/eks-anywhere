package oras

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
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
	host, namespace := splitRegistryNamespace(fr.registry)
	artifacts := ReadFilesFromBundles(bundles)
	for _, a := range UniqueCharts(artifacts) {
		updatedChartURL := urls.ReplaceHost(a, host)
		if namespace != "" {
			updatedChartURL = insertNamespace(updatedChartURL, namespace)
		}
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

// splitRegistryNamespace separates a registry string into host:port and optional namespace.
func splitRegistryNamespace(registry string) (host, namespace string) {
	slashIdx := strings.Index(registry, "/")
	if slashIdx == -1 {
		return registry, ""
	}
	return registry[:slashIdx], registry[slashIdx+1:]
}

// insertNamespace inserts a namespace path segment after the host in a URL.
func insertNamespace(url, namespace string) string {
	prefix := oci.OCIPrefix
	if !strings.HasPrefix(url, prefix) {
		slashIdx := strings.Index(url, "/")
		if slashIdx == -1 {
			return url + "/" + namespace
		}
		return url[:slashIdx] + "/" + namespace + url[slashIdx:]
	}
	withoutScheme := strings.TrimPrefix(url, prefix)
	slashIdx := strings.Index(withoutScheme, "/")
	if slashIdx == -1 {
		return prefix + withoutScheme + "/" + namespace
	}
	return prefix + withoutScheme[:slashIdx] + "/" + namespace + withoutScheme[slashIdx:]
}

func ChartFileName(chart string) string {
	return strings.Replace(filepath.Base(chart), ":", "-", 1) + ".yaml"
}
