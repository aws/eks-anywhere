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

func TestChartRegistryImporterImport(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password)
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", "oci://registry.com:443/project")
	client.EXPECT().PushChart(ctx, "folder/chart2-v2.2.0.tgz", "oci://registry.com:443/project")

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(Succeed())
}

func TestChartRegistryImporterImportWithNamespacePath(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	// Registry with a namespace/path component (OCI namespace scenario)
	registry := "192.168.1.1:443/eks-a-test"
	srcFolder := "folder"
	charts := []string{"public.ecr.aws/eks/cilium/cilium:1.17.15-1", "public.ecr.aws/w9m0f3l5/eks-anywhere-packages:0.4.9-eks-a-121"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	// RegistryLogin should only receive host:port, NOT the path
	client.EXPECT().RegistryLogin(ctx, "192.168.1.1:443", user, password)
	// PushChart should receive OCI URL with namespace injected between host and original path
	// "public.ecr.aws/eks/cilium/cilium:1.17.15-1" -> ChartPushURL -> "oci://public.ecr.aws/eks/cilium"
	// ReplaceHost with "192.168.1.1:443" -> "oci://192.168.1.1:443/eks/cilium"
	// injectNamespace with "eks-a-test" -> "oci://192.168.1.1:443/eks-a-test/eks/cilium"
	client.EXPECT().PushChart(ctx, "folder/cilium-1.17.15-1.tgz", "oci://192.168.1.1:443/eks-a-test/eks/cilium")
	// "public.ecr.aws/w9m0f3l5/eks-anywhere-packages:0.4.9-eks-a-121" -> ChartPushURL -> "oci://public.ecr.aws/w9m0f3l5"
	// ReplaceHost with "192.168.1.1:443" -> "oci://192.168.1.1:443/w9m0f3l5"
	// injectNamespace with "eks-a-test" -> "oci://192.168.1.1:443/eks-a-test/w9m0f3l5"
	client.EXPECT().PushChart(ctx, "folder/eks-anywhere-packages-0.4.9-eks-a-121.tgz", "oci://192.168.1.1:443/eks-a-test/w9m0f3l5")

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(Succeed())
}

func TestChartRegistryImporterImportWithMultiLevelNamespacePath(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	// Registry with multi-level namespace path
	registry := "10.0.0.1:5000/org/team"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	// RegistryLogin should only receive host:port
	client.EXPECT().RegistryLogin(ctx, "10.0.0.1:5000", user, password)
	// Chart push URL should include the full namespace path
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", "oci://10.0.0.1:5000/org/team/project")

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(Succeed())
}

func TestChartRegistryImporterImportLoginError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password).Return(errors.New("logging error"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("importing charts: logging error")))
}

func TestChartRegistryImporterImportLoginErrorWithNamespacePath(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "192.168.1.1:443/eks-a-test"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	// Login should be attempted with host:port only
	client.EXPECT().RegistryLogin(ctx, "192.168.1.1:443", user, password).Return(errors.New("login failed"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("importing charts: login failed")))
}

func TestChartRegistryImporterImportPushError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "registry.com:443"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0", "ecr.com/project/chart2:v2.2.0", "ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, registry, user, password)
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", "oci://registry.com:443/project").Return(errors.New("pushing error"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("pushing chart [ecr.com/project/chart1:v1.1.0] to registry [registry.com:443]: pushing error")))
}

func TestChartRegistryImporterImportPushErrorWithNamespacePath(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	user := "u"
	password := "pass"
	registry := "192.168.1.1:443/eks-a-test"
	srcFolder := "folder"
	charts := []string{"ecr.com/project/chart1:v1.1.0"}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	client.EXPECT().RegistryLogin(ctx, "192.168.1.1:443", user, password)
	client.EXPECT().PushChart(ctx, "folder/chart1-v1.1.0.tgz", "oci://192.168.1.1:443/eks-a-test/project").Return(errors.New("pushing error"))

	i := helm.NewChartRegistryImporter(client, srcFolder, registry, user, password)
	g.Expect(i.Import(ctx, charts...)).To(MatchError(ContainSubstring("pushing chart [ecr.com/project/chart1:v1.1.0] to registry [192.168.1.1:443/eks-a-test]: pushing error")))
}

func TestSplitRegistryHostAndPath(t *testing.T) {
	tests := []struct {
		name         string
		registry     string
		expectedHost string
		expectedPath string
	}{
		{
			name:         "host:port only",
			registry:     "192.168.1.1:443",
			expectedHost: "192.168.1.1:443",
			expectedPath: "",
		},
		{
			name:         "host:port with single namespace",
			registry:     "192.168.1.1:443/eks-a-test",
			expectedHost: "192.168.1.1:443",
			expectedPath: "eks-a-test",
		},
		{
			name:         "host:port with multi-level namespace",
			registry:     "10.0.0.1:5000/org/team",
			expectedHost: "10.0.0.1:5000",
			expectedPath: "org/team",
		},
		{
			name:         "hostname without port",
			registry:     "myregistry.com",
			expectedHost: "myregistry.com",
			expectedPath: "",
		},
		{
			name:         "hostname with path",
			registry:     "myregistry.com/namespace",
			expectedHost: "myregistry.com",
			expectedPath: "namespace",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			host, path := helm.SplitRegistryHostAndPath(tt.registry)
			g.Expect(host).To(Equal(tt.expectedHost))
			g.Expect(path).To(Equal(tt.expectedPath))
		})
	}
}

func TestInjectNamespace(t *testing.T) {
	tests := []struct {
		name      string
		ociURL    string
		namespace string
		expected  string
	}{
		{
			name:      "inject single namespace",
			ociURL:    "oci://192.168.1.1:443/eks/cilium",
			namespace: "eks-a-test",
			expected:  "oci://192.168.1.1:443/eks-a-test/eks/cilium",
		},
		{
			name:      "inject multi-level namespace",
			ociURL:    "oci://10.0.0.1:5000/project",
			namespace: "org/team",
			expected:  "oci://10.0.0.1:5000/org/team/project",
		},
		{
			name:      "url without path",
			ociURL:    "oci://192.168.1.1:443",
			namespace: "eks-a-test",
			expected:  "oci://192.168.1.1:443/eks-a-test",
		},
		{
			name:      "empty namespace returns original",
			ociURL:    "oci://192.168.1.1:443/project",
			namespace: "",
			expected:  "oci://192.168.1.1:443/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			result := helm.InjectNamespace(tt.ociURL, tt.namespace)
			g.Expect(result).To(Equal(tt.expected))
		})
	}
}
