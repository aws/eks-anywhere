package helm

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/utils/oci"
)

type ChartRegistryDownloader struct {
	client    Client
	dstFolder string

	Retrier retrier.Retrier
}

func NewChartRegistryDownloader(client Client, dstFolder string) *ChartRegistryDownloader {
	return &ChartRegistryDownloader{
		client:    client,
		dstFolder: dstFolder,
		Retrier:   *retrier.NewWithMaxRetries(5, 200*time.Second),
	}
}

func (d *ChartRegistryDownloader) Download(ctx context.Context, charts ...string) error {
	for _, chart := range uniqueCharts(charts) {
		chartURL, chartVersion := oci.ChartURLAndVersion(chart)
		logger.Info("Saving helm chart to disk", "chart", chart)
		err := d.Retrier.Retry(func() error {
			return d.client.SaveChart(ctx, chartURL, chartVersion, d.dstFolder)
		})
		if err != nil {
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
