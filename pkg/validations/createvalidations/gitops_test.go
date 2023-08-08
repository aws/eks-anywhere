package createvalidations_test

import (
	"context"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
)

func TestValidateGitOpsConfigFluxConfigBothNil(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(Succeed())
}

func TestValidateGitOpsConfigGetGitOpsConfigError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(errors.New("error get gitOpsConfig"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get gitOpsConfig")))
}

func TestValidateGitOpsConfigNameExists(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(nil)
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("gitOpsConfig gitops already exists")))
}

func TestValidateGitOpsConfigGetClusterError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(nil, errors.New("error get cluster"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get cluster")))
}

func TestValidateGitOpsConfigGetMgmtGitOpsConfigError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(errors.New("error get gitops"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get gitops")))
}

func TestValidateGitOpsConfigNotEqual(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().
		GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).
		DoAndReturn(func(_ context.Context, _, _, _, _ string, obj *v1alpha1.GitOpsConfig) error {
			obj.Spec = v1alpha1.GitOpsConfigSpec{
				Flux: v1alpha1.Flux{
					Github: v1alpha1.Github{
						FluxSystemNamespace: "custom",
					},
				},
			}
			return nil
		})
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("expected gitOpsConfig.spec to be the same")))
}

func TestValidateGitOpsConfigSuccess(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.c.Opts.Spec.GitOpsConfig = &v1alpha1.GitOpsConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().GetObject(tt.ctx, "gitopsconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.GitOpsConfig{}).Return(nil)
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(Succeed())
}

func TestValidateFluxConfigGetFluxConfigError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(errors.New("error get fluxConfig"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get fluxConfig")))
}

func TestValidateFluxConfigNameExists(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(nil)
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("fluxConfig gitops already exists")))
}

func TestValidateFluxConfigGetClusterError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(nil, errors.New("error get cluster"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get cluster")))
}

func TestValidateFluxConfigGetMgmtFluxConfigError(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(errors.New("error get gitops"))
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("error get gitops")))
}

func TestValidateFluxConfigNotEqual(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().
		GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.FluxConfig{}).
		DoAndReturn(func(_ context.Context, _, _, _, _ string, obj *v1alpha1.FluxConfig) error {
			obj.Spec = v1alpha1.FluxConfigSpec{
				SystemNamespace: "custom",
			}
			return nil
		})
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(MatchError(ContainSubstring("expected fluxConfig.spec to be the same")))
}

func TestValidateFluxConfigSuccess(t *testing.T) {
	tt := newPreflightValidationsTest(t)
	tt.c.Opts.Spec.FluxConfig = &v1alpha1.FluxConfig{}
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(apierrors.NewNotFound(schema.GroupResource{Group: "", Resource: ""}, ""))
	tt.k.EXPECT().GetEksaCluster(tt.ctx, tt.c.Opts.Spec.ManagementCluster, tt.c.Opts.Spec.Cluster.ManagedBy()).Return(&v1alpha1.Cluster{Spec: v1alpha1.ClusterSpec{GitOpsRef: &v1alpha1.Ref{Name: "gitops"}}}, nil)
	tt.k.EXPECT().GetObject(tt.ctx, "fluxconfigs.anywhere.eks.amazonaws.com", "gitops", "", "kubeconfig", &v1alpha1.FluxConfig{}).Return(nil)
	tt.Expect(createvalidations.ValidateGitOps(tt.ctx, tt.k, tt.c.Opts.ManagementCluster, tt.c.Opts.Spec)).To(Succeed())
}
