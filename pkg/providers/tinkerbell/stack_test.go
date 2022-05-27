package tinkerbell

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestTinkerbellStackInstallWithAllOptionsSuccess(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	ctx := context.Background()

	tinkManifestName := "tink.yaml"
	bootsImageURI := "public.ecr.aws/eks-anywhere/tinkerbell/boots:latest"
	hegelManifestName := "hegel.yaml"

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			Manifest: releasev1alpha1.Manifest{URI: tinkManifestName},
		},
		Boots: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: bootsImageURI},
		},
		Hegel: releasev1alpha1.TinkerbellServiceBundle{
			Manifest: releasev1alpha1.Manifest{URI: hegelManifestName},
		},
	}

	stack := NewStackInstaller(clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack, docker, writer, helm, "1.2.3.4").
		WithNamespace(constants.EksaSystemNamespace, true).
		WithTinkController().
		WithTinkServer().
		WithBootsOnKubernetes().
		WithHegel()

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)

	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, helmChartOci, helmChartVersion, cluster.KubeconfigFile, overridesFileName)

	err := stack.Install(ctx, cluster.KubeconfigFile)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackInstallWithBootsOnDockerSuccess(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	ctx := context.Background()

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			TinkWorker: releasev1alpha1.Image{URI: "tink-worker:latest"},
		},
		Boots: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: "boots:latest"},
		},
		Hegel: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: "hegel:latest"},
		},
	}

	stack := NewStackInstaller(clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack, docker, writer, helm, "1.2.3.4").WithBootsOnDocker()

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)

	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, helmChartOci, helmChartVersion, cluster.KubeconfigFile, overridesFileName)

	docker.EXPECT().Run(ctx, "boots:latest",
		boots,
		[]string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67"},
		"-v", gomock.Any(),
		"--network", "host",
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
	)

	err := stack.Install(ctx, cluster.KubeconfigFile)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackInstallWithBootsOnKubernetes(t *testing.T) {
	t.Setenv(features.TinkerbellProviderEnvVar, "true")
	clusterSpecManifest := "cluster_tinkerbell_stacked_etcd.yaml"
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}

	clusterSpec := givenClusterSpec(t, clusterSpecManifest)
	ctx := context.Background()

	clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack = releasev1alpha1.TinkerbellStackBundle{
		Tink: releasev1alpha1.TinkBundle{
			TinkWorker: releasev1alpha1.Image{URI: "tink-worker:latest"},
		},
		Boots: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: "boots:latest"},
		},
		Hegel: releasev1alpha1.TinkerbellServiceBundle{
			Image: releasev1alpha1.Image{URI: "hegel:latest"},
		},
	}

	stack := NewStackInstaller(clusterSpec.VersionsBundle.Tinkerbell.TinkerbellStack, docker, writer, helm, "1.2.3.4").WithBootsOnKubernetes()

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)

	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, helmChartOci, helmChartVersion, cluster.KubeconfigFile, overridesFileName)

	err := stack.Install(ctx, cluster.KubeconfigFile)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}
