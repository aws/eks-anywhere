package stack_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/aws/eks-anywhere/pkg/constants"
	filewritermocks "github.com/aws/eks-anywhere/pkg/filewriter/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/stack/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	overridesFileName = "tinkerbell-chart-overrides.yaml"
	boots             = "boots"
	testIP            = "1.2.3.4"

	helmChartPath    = "public.ecr.aws/eks-anywhere/tinkerbell/tinkerbell-chart"
	helmChartName    = "tinkerbell-chart"
	helmChartVersion = "0.1.0"
)

var helmChartURI = fmt.Sprintf("%s:%s", helmChartPath, helmChartVersion)

func getTinkBundle() releasev1alpha1.TinkerbellBundle {
	return releasev1alpha1.TinkerbellBundle{
		TinkerbellStack: releasev1alpha1.TinkerbellStackBundle{
			Tink: releasev1alpha1.TinkBundle{
				TinkWorker: releasev1alpha1.Image{URI: "tink-worker:latest"},
			},
			Boots: releasev1alpha1.TinkerbellServiceBundle{
				Image: releasev1alpha1.Image{URI: "boots:latest"},
			},
			Hegel: releasev1alpha1.TinkerbellServiceBundle{
				Image: releasev1alpha1.Image{URI: "hegel:latest"},
			},
			Hook: releasev1alpha1.HookBundle{
				Initramfs: releasev1alpha1.HookArch{
					Amd: releasev1alpha1.Archive{
						URI: "https://anywhere-assests.eks.amazonaws.com/tinkerbell/hook/initramfs-x86-64",
					},
				},
			},
			TinkebellChart: releasev1alpha1.Image{
				Name: helmChartName,
				URI:  helmChartURI,
			},
		},
		KubeVip: releasev1alpha1.Image{
			URI: "public.ecr.aws/eks-anywhere/kube-vip:latest",
		},
	}
}

func TestTinkerbellStackInstallWithAllOptionsSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)

	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, fmt.Sprintf("oci://%s", helmChartPath), helmChartVersion, cluster.KubeconfigFile, overridesFileName)

	if err := s.Install(ctx,
		getTinkBundle(),
		testIP,
		cluster.KubeconfigFile,
		"",
		stack.WithNamespaceCreate(true),
		stack.WithBootsOnKubernetes(),
		stack.WithHostPortEnabled(true),
		stack.WithLoadBalancer(),
	); err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackInstallHookOverrideSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	ctx := context.Background()

	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)

	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, fmt.Sprintf("oci://%s", helmChartPath), helmChartVersion, cluster.KubeconfigFile, overridesFileName)

	if err := s.Install(ctx,
		getTinkBundle(),
		testIP,
		cluster.KubeconfigFile,
		"https://hook-override-path",
		stack.WithNamespaceCreate(true),
		stack.WithBootsOnKubernetes(),
		stack.WithHostPortEnabled(true),
	); err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackInstallWithBootsOnDockerSuccess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	cluster := &types.Cluster{Name: "test"}
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	writer.EXPECT().Write(overridesFileName, gomock.Any()).Return(overridesFileName, nil)
	helm.EXPECT().InstallChartWithValuesFile(ctx, helmChartName, fmt.Sprintf("oci://%s", helmChartPath), helmChartVersion, cluster.KubeconfigFile, overridesFileName)
	docker.EXPECT().Run(ctx, "boots:latest",
		boots,
		[]string{"-kubeconfig", "/kubeconfig", "-dhcp-addr", "0.0.0.0:67", "-osie-path-override", "https://anywhere-assests.eks.amazonaws.com/tinkerbell/hook"},
		"-v", gomock.Any(),
		"--network", "host",
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
		"-e", gomock.Any(),
	)

	err := s.Install(ctx, getTinkBundle(), testIP, cluster.KubeconfigFile, "", stack.WithBootsOnDocker())
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackUninstallLocalSucess(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	docker.EXPECT().ForceRemove(ctx, boots)

	err := s.UninstallLocal(ctx)
	if err != nil {
		t.Fatalf("failed to install Tinkerbell stack: %v", err)
	}
}

func TestTinkerbellStackUninstallLocalFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	dockerError := "docker error"
	expectedError := fmt.Sprintf("removing local boots container: %s", dockerError)
	docker.EXPECT().ForceRemove(ctx, boots).Return(errors.New(dockerError))

	err := s.UninstallLocal(ctx)
	assert.EqualError(t, err, expectedError, "Error should be: %v, got: %v", expectedError, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDoesNotExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(true, nil)
	docker.EXPECT().ForceRemove(ctx, "boots")

	err := s.CleanupLocalBoots(ctx, true)
	assert.NoError(t, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDoesExist(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)
	expectedErrorMsg := "boots container already exists, delete the container manually or re-run the command with --force-cleanup"

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(true, nil)

	err := s.CleanupLocalBoots(ctx, false)
	assert.EqualError(t, err, expectedErrorMsg, "Error should be: %v, got: %v", expectedErrorMsg, err)
}

func TestTinkerbellStackCheckLocalBootsExistenceDockerError(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	docker := mocks.NewMockDocker(mockCtrl)
	helm := mocks.NewMockHelm(mockCtrl)
	writer := filewritermocks.NewMockFileWriter(mockCtrl)
	ctx := context.Background()
	s := stack.NewInstaller(docker, writer, helm, constants.EksaSystemNamespace)

	docker.EXPECT().CheckContainerExistence(ctx, "boots").Return(false, nil)

	err := s.CleanupLocalBoots(ctx, true)
	assert.NoError(t, err)
}
