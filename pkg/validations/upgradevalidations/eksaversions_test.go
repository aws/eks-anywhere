package upgradevalidations_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	providermocks "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

func TestValidateEKSAVersionSkew(t *testing.T) {
	tests := []struct {
		name              string
		wantErr           error
		upgradeVersion    string
		clusterVersionTag string
	}{
		{
			name:              "FailureTwoMinorVersions",
			wantErr:           errors.New("WARNING: version difference between upgrade version (0.15) and cluster version (0.13) do not meet the supported version increment of +1"),
			upgradeVersion:    "v0.15.3",
			clusterVersionTag: "v0.13.0-eks-a",
		},
		{
			name:              "FailureMinusOneMinorVersion",
			wantErr:           errors.New("WARNING: version difference between upgrade version (0.14) and cluster version (0.15) do not meet the supported version increment of +1"),
			upgradeVersion:    "v0.14.0",
			clusterVersionTag: "v0.15.0-eks-a",
		},
		{
			name:              "SuccessSameVersion",
			wantErr:           nil,
			upgradeVersion:    "v0.15.3",
			clusterVersionTag: "v0.15.3-eks-a",
		},
		{
			name:              "SuccessOneMinorVersion",
			wantErr:           nil,
			upgradeVersion:    "v0.15.3",
			clusterVersionTag: "v0.14.0-eks-a",
		},
		{
			name:              "SuccessDevVersion",
			wantErr:           nil,
			upgradeVersion:    "v0.0.0-dev",
			clusterVersionTag: "v0.15.3-eks-a-v0.0.0-dev-build.6886",
		},
		{
			name:              "FailureParseClusterVersion",
			wantErr:           errors.New("parsing cluster cli version"),
			upgradeVersion:    "v0.15.3",
			clusterVersionTag: "badvalue",
		},
		{
			name:              "FailureParseUpgradeVersion",
			wantErr:           errors.New("parsing upgrade cli version"),
			upgradeVersion:    "badvalue",
			clusterVersionTag: "v0.15.3",
		},
	}

	for _, tc := range tests {
		test := newTest(t, withKubectl())
		t.Run(tc.name, func(tt *testing.T) {
			clusterName := "upgrade-cluster"
			uCluster := getCluster(clusterName)
			upgradeCluster := anywhereCluster(clusterName)
			upgradeCluster.Spec.BundlesRef = &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: constants.EksaSystemNamespace,
			}
			uri := fmt.Sprintf("public.ecr.aws/l0g8r8j6/eks-anywhere-cluster-controller:%s", tc.clusterVersionTag)
			vb := releasev1alpha1.VersionsBundle{
				Eksa: releasev1alpha1.EksaBundle{
					ClusterController: releasev1alpha1.Image{
						URI: uri,
					},
				},
			}
			bundle := &releasev1alpha1.Bundles{
				Spec: releasev1alpha1.BundlesSpec{
					VersionsBundles: []releasev1alpha1.VersionsBundle{
						vb,
					},
				},
			}

			ctx := context.Background()
			test.kubectl.EXPECT().GetEksaCluster(ctx, uCluster, uCluster.Name).Return(upgradeCluster, nil)
			test.kubectl.EXPECT().GetBundles(ctx, uCluster.KubeconfigFile, upgradeCluster.Spec.BundlesRef.Name, upgradeCluster.Spec.BundlesRef.Namespace).Return(bundle, nil)

			err := upgradevalidations.ValidateEKSAVersionSkew(ctx, tc.upgradeVersion, test.kubectl, uCluster)
			if tc.wantErr != nil {
				test.Expect(err.Error()).To(ContainSubstring(tc.wantErr.Error()))
			} else {
				test.Expect(err).To(BeNil())
			}
		})
	}
}

func TestValidateEKSAVersionSkewFailGetCluster(t *testing.T) {
	tt := newTest(t, withKubectl())
	wantErr := "failed to reach cluster"
	clusterName := "upgrade-cluster"
	uCluster := getCluster(clusterName)

	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, uCluster, uCluster.Name).Return(nil, errors.New(wantErr))

	err := upgradevalidations.ValidateEKSAVersionSkew(ctx, "", tt.kubectl, uCluster)
	tt.Expect(err.Error()).To(Equal(wantErr))
}

func TestValidateEKSAVersionSkewFailBundleRef(t *testing.T) {
	tt := newTest(t, withKubectl())
	wantErr := "cluster bundlesRef cannot be nil"
	clusterName := "upgrade-cluster"
	uCluster := getCluster(clusterName)
	upgradeCluster := anywhereCluster(clusterName)

	upgradeCluster.Spec.BundlesRef = nil
	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, uCluster, uCluster.Name).Return(upgradeCluster, nil)

	err := upgradevalidations.ValidateEKSAVersionSkew(ctx, "", tt.kubectl, uCluster)
	tt.Expect(err.Error()).To(Equal(wantErr))
}

func TestValidateEKSAVersionSkewFailParseTags(t *testing.T) {
	tt := newTest(t, withKubectl())
	wantErr := "could not find tag in Eksa Cluster Controller Image"
	clusterName := "upgrade-cluster"
	uCluster := getCluster(clusterName)
	upgradeCluster := anywhereCluster(clusterName)
	upgradeCluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		Name:      "bundles-1",
		Namespace: constants.EksaSystemNamespace,
	}
	uri := "badvalue"
	vb := releasev1alpha1.VersionsBundle{
		Eksa: releasev1alpha1.EksaBundle{
			ClusterController: releasev1alpha1.Image{
				URI: uri,
			},
		},
	}
	bundle := &releasev1alpha1.Bundles{
		Spec: releasev1alpha1.BundlesSpec{
			VersionsBundles: []releasev1alpha1.VersionsBundle{
				vb,
			},
		},
	}

	ctx := context.Background()
	tt.kubectl.EXPECT().GetEksaCluster(ctx, uCluster, uCluster.Name).Return(upgradeCluster, nil)
	tt.kubectl.EXPECT().GetBundles(ctx, uCluster.KubeconfigFile, upgradeCluster.Spec.BundlesRef.Name, upgradeCluster.Spec.BundlesRef.Namespace).Return(bundle, nil)

	err := upgradevalidations.ValidateEKSAVersionSkew(ctx, "", tt.kubectl, uCluster)
	tt.Expect(err.Error()).To(Equal(wantErr))
}

func getCluster(name string) *types.Cluster {
	return &types.Cluster{
		Name: name,
	}
}

func anywhereCluster(name string) *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: v1.ObjectMeta{
			Name: name,
		},
		Spec: anywherev1.ClusterSpec{
			ManagementCluster: anywherev1.ManagementCluster{
				Name: name,
			},
		},
	}
}

type clusterTest struct {
	*WithT
	kubectl     *mocks.MockKubectlClient
	provider    *providermocks.MockProvider
	clusterSpec *cluster.Spec
}

type clusterTestOpt func(t *testing.T, ct *clusterTest)

func newTest(t *testing.T, opts ...clusterTestOpt) *clusterTest {
	ctrl := gomock.NewController(t)
	cTest := &clusterTest{
		WithT:       NewWithT(t),
		clusterSpec: test.NewClusterSpec(),
		provider:    providermocks.NewMockProvider(ctrl),
	}
	for _, opt := range opts {
		opt(t, cTest)
	}
	return cTest
}

func withKubectl() clusterTestOpt {
	return func(t *testing.T, ct *clusterTest) {
		ctrl := gomock.NewController(t)
		ct.kubectl = mocks.NewMockKubectlClient(ctrl)
	}
}
