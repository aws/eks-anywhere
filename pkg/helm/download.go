package helm

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

type ChartRegistryDownloader struct {
	client    Client
	dstFolder string
}

func NewChartRegistryDownloader(client Client, dstFolder string) *ChartRegistryDownloader {
	return &ChartRegistryDownloader{
		client:    client,
		dstFolder: dstFolder,
	}
}

func (d *ChartRegistryDownloader) Download(ctx context.Context, charts ...string) error {
	for _, chart := range uniqueCharts(charts) {
		chartURL, chartVersion := oci.ChartURLAndVersion(chart)
		logger.Info("Saving helm chart to disk", "chart", chart)
		if err := d.client.SaveChart(ctx, chartURL, chartVersion, d.dstFolder); err != nil {
			return fmt.Errorf("downloading chart [%s] from registry: %v", chart, err)
		}
	}

	return nil
}

func uniqueCharts(charts []string) []string {
	c := types.SliceToLookup(charts).ToSlice()
	// TODO: maybe optimize this, avoiding the sort and just following the same order as the original slice
	sort.Strings(c)
	return c
}
