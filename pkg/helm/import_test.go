package helm_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/helm"
	"github.com/aws/eks-anywhere/pkg/helm/mocks"
)

func TestChartRegistryImporterImport(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	namespace := constants.DefaultRegistryMirrorNamespace
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password)
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", fmt.Sprintf("oci://registry.com:443/%s/project", namespace))
	client.EXPECT().PushChart(ctx, "folder/chart2-v2.2.0.tgz", fmt.Sprintf("oci://registry.com:443/%s/project", namespace))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, namespace, user, password)
	g.Expect(i.Import(ctx, charts...)).To(Succeed())
}

func TestChartRegistryImporterImportLoginError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	namespace := ""
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password).Return(errors.New("logging error"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, namespace, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("importing charts: logging error")))
}

func TestChartRegistryImporterImportPushError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	namespace := "custom"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password)
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", fmt.Sprintf("oci://registry.com:443/%s/project", namespace)).Return(errors.New("pushing error"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, namespace, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("pushing chart [ecr.com/project/chart1:v1.1.0] to registry [oci://registry.com:443/custom/project]: pushing error")))
}
