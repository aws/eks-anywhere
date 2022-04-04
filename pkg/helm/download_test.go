package helm_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/helm/mocks"
)

func TestChartRegistryDownloaderDownload(t *testing.T) {
	g := NewWithT(t)
	charts := []string{"ecr.com/chart1:v1.1.0", "ecr.com/chart2:v2.2.0", "ecr.com/chart1:v1.1.0"}
	ctx := context.Background()
	folder := "folder"
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	client.EXPECT().SaveChart(ctx, "oci://ecr.com/chart1", "v1.1.0", folder)
	client.EXPECT().SaveChart(ctx, "oci://ecr.com/chart2", "v2.2.0", folder)

	d := helm.NewChartRegistryDownloader(client, folder)
	g.Expect(d.Download(ctx, charts...)).To(Succeed())
}

func TestChartRegistryDownloaderDownloadError(t *testing.T) {
	g := NewWithT(t)
	charts := []string{"ecr.com/chart1:v1.1.0", "ecr.com/chart2:v2.2.0", "ecr.com/chart1:v1.1.0"}
	ctx := context.Background()
	folder := "folder"
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	client.EXPECT().SaveChart(ctx, "oci://ecr.com/chart1", "v1.1.0", folder)
	client.EXPECT().SaveChart(ctx, "oci://ecr.com/chart2", "v2.2.0", folder).Return(errors.New("failed downloading"))

	d := helm.NewChartRegistryDownloader(client, folder)
	g.Expect(d.Download(ctx, charts...)).To(MatchError(ContainSubstring("downloading chart [ecr.com/chart2:v2.2.0] from registry: failed downloading")))
}
