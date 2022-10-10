package helm

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/aws/eks-anywhere/pkg/docker"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

type ChartRegistryImporter struct {
	client             Client
	registry           string
	ociNamespace       string
	username, password string
	srcFolder          string
}

func NewChartRegistryImporter(client Client, srcFolder, registry, ociNamespace, username, password string) *ChartRegistryImporter {
	return &ChartRegistryImporter{
		client:       client,
		srcFolder:    srcFolder,
		registry:     registry,
		ociNamespace: ociNamespace,
		username:     username,
		password:     password,
	}
}

func (i *ChartRegistryImporter) Import(ctx context.Context, charts ...string) error {
	if err := i.client.RegistryLogin(ctx, i.registry, i.username, i.password); err != nil {
		return fmt.Errorf("importing charts: %v", err)
	}

	for _, chart := range uniqueCharts(charts) {
		pushChartURL := oci.ChartPushURL(chart)
		pushChartURL = docker.ReplaceHostWithNamespacedEndpoint(pushChartURL, i.registry, i.ociNamespace)

		chartFilepath := filepath.Join(i.srcFolder, ChartFileName(chart))
		if err := i.client.PushChart(ctx, chartFilepath, pushChartURL); err != nil {
			return fmt.Errorf("pushing chart [%s] to registry [%s]: %v", chart, pushChartURL, err)
		}
	}

	return nil
}
