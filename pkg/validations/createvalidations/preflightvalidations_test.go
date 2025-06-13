package createvalidations_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
	internalmocks "github.com/aws/eks-anywhere/internal/test/mocks"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/manifests"
	"github.com/aws/eks-anywhere/pkg/manifests/releases"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
)

type preflightValidationsTest struct {
	*WithT
	ctx context.Context
	k   *mocks.MockKubectlClient
	c   *createvalidations.CreateValidations
}

func newPreflightValidationsTest(t *testing.T) *preflightValidationsTest {
	ctrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(ctrl)

	version := anywherev1.EksaVersion("v0.22.0")

	c := &types.Cluster{
		KubeconfigFile: "kubeconfig",
	}

	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &anywherev1.Cluster{
			Spec: anywherev1.ClusterSpec{
				KubernetesVersion: "1.31",
				GitOpsRef: &anywherev1.Ref{
					Name: "gitops",
				},
				EksaVersion: &version,
			},
		}
	})

	eksaReleaseV022 := test.EKSARelease()
	eksaReleaseV022.Name = "eksa-v0-22-0"
	eksaReleaseV022.Spec.Version = "eksa-v0-22-0"

	eksdRelease := test.EksdRelease("1-31")

	objects := []client.Object{eksaReleaseV022, eksdRelease}
	opts := &validations.Opts{
		Kubectl:           k,
		Spec:              clusterSpec,
		WorkloadCluster:   c,
		ManagementCluster: c,
		CliVersion:        string(version),
		KubeClient:        test.NewFakeKubeClient(objects...),
	}
	return &preflightValidationsTest{
		WithT: NewWithT(t),
		ctx:   context.Background(),
		k:     k,
		c:     createvalidations.New(opts),
	}
}

func addManifestReaderMock(t *testing.T, version anywherev1.EksaVersion) *manifests.Reader {
	ctrl := gomock.NewController(t)
	reader := internalmocks.NewMockReader(ctrl)
	releasesURL := releases.ManifestURL()

	releasesManifest := fmt.Sprintf(`apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Release
metadata:
  name: release-1
spec:
  releases:
  - bundleManifestUrl: "https://bundles/bundles.yaml"
    version: %s`, string(version))
	reader.EXPECT().ReadFile(releasesURL).Return([]byte(releasesManifest), nil)

	bundlesManifest := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Bundles
metadata:
  annotations:
    anywhere.eks.amazonaws.com/signature: MEQCICjq1rZmhH0FYOlruZmh6QADCrr5ccrN6hE7Lu0vaXGrAiBhV+kfh64sqLblBt98DvIfHMerEqJVhHzpGy1YJthZQw==
  name: bundles-1
spec:
  number: 1
  versionsBundles:
  - kubeVersion: "1.31"
    endOfStandardSupport: "2026-12-31"
    eksD:
      name: "test"
      channel: "1-31"
      manifestUrl: "https://distro.eks.amazonaws.com/kubernetes-1-31/kubernetes-1-31-eks-1.yaml"`

	reader.EXPECT().ReadFile("https://bundles/bundles.yaml").Return([]byte(bundlesManifest), nil)

	return manifests.NewReader(reader)
}

func TestPreFlightValidationsGitProvider(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.ManifestReader = addManifestReaderMock(t, anywherev1.EksaVersion(tt.c.Opts.CliVersion))
	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}

func TestPreFlightValidationsWorkloadCluster(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	mgmtClusterName := "mgmt-cluster"
	tt.c.Opts.Spec.Cluster.SetManagedBy(mgmtClusterName)
	tt.c.Opts.Spec.Cluster.Spec.ManagementCluster.Name = mgmtClusterName
	tt.c.Opts.ManagementCluster.Name = mgmtClusterName
	version := anywherev1.EksaVersion(tt.c.Opts.CliVersion)
	tt.c.Opts.ManifestReader = addManifestReaderMock(t, version)

	mgmt := &anywherev1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: "mgmt-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: "1.30",
			ManagementCluster: anywherev1.ManagementCluster{
				Name: "mgmt-cluster",
			},
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-29",
				Namespace: constants.EksaSystemNamespace,
			},
			EksaVersion: &version,
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				KubeletConfiguration: &unstructured.Unstructured{
					Object: map[string]interface{}{
						"staticPodPath": "path",
					},
				},
			},
		},
	}

	tt.c.Opts.Spec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef = &anywherev1.Ref{
		Name: "cpRef",
	}
	tt.c.Opts.Spec.VSphereMachineConfigs = map[string]*anywherev1.VSphereMachineConfig{
		"cpRef": {
			Spec: anywherev1.VSphereMachineConfigSpec{
				OSFamily: anywherev1.Bottlerocket,
			},
		},
	}

	tt.c.Opts.Spec.Cluster.Spec.WorkerNodeGroupConfigurations = []anywherev1.WorkerNodeGroupConfiguration{
		{
			MachineGroupRef: &anywherev1.Ref{
				Name: "wnRef",
			},
		},
	}

	tt.c.Opts.Spec.VSphereMachineConfigs["wnRef"] = &anywherev1.VSphereMachineConfig{
		Spec: anywherev1.VSphereMachineConfigSpec{
			OSFamily: anywherev1.Bottlerocket,
		},
	}

	tt.k.EXPECT().GetClusters(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil, nil)
	tt.k.EXPECT().ValidateClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().ValidateEKSAClustersCRD(tt.ctx, tt.c.Opts.WorkloadCluster).Return(nil)
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.ManagementCluster, mgmtClusterName).Return(mgmt, nil).MaxTimes(3)

	tt.Expect(validations.ProcessValidationResults(tt.c.PreflightValidations(tt.ctx))).To(Succeed())
}
