package flux

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/gitops/flux/mocks"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

type fluxClientTest struct {
	*WithT
	ctx        context.Context
	c          *fluxClient
	f          *mocks.MockFluxClient
	k          *mocks.MockKubeClient
	cluster    *types.Cluster
	fluxConfig *v1alpha1.FluxConfig
}

func newFluxClientTest(t *testing.T) *fluxClientTest {
	ctrl := gomock.NewController(t)
	f := mocks.NewMockFluxClient(ctrl)
	k := mocks.NewMockKubeClient(ctrl)
	c := newFluxClient(f, k)
	c.Retrier = retrier.NewWithMaxRetries(maxRetries, 0)
	return &fluxClientTest{
		WithT:      NewWithT(t),
		ctx:        context.Background(),
		c:          c,
		f:          f,
		k:          k,
		cluster:    &types.Cluster{},
		fluxConfig: &v1alpha1.FluxConfig{},
	}
}

func TestFluxClientBootstrapGithubSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in bootstrap github")).Times(4)
	tt.f.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).Times(1)

	tt.Expect(tt.c.BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig)).To(Succeed(), "fluxClient.BootstrapGithub() should succeed with 5 tries")
}

func TestFluxClientBootstrapGithubError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in bootstrap github")).Times(5)
	tt.f.EXPECT().BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).AnyTimes()

	tt.Expect(tt.c.BootstrapGithub(tt.ctx, tt.cluster, tt.fluxConfig)).To(MatchError(ContainSubstring("error in bootstrap github")), "fluxClient.BootstrapGithub() should fail after 5 tries")
}

func TestFluxClientBootstrapGitSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil).Return(errors.New("error in bootstrap git")).Times(4)
	tt.f.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil).Return(nil).Times(1)

	tt.Expect(tt.c.BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil)).To(Succeed(), "fluxClient.BootstrapGit() should succeed with 5 tries")
}

func TestFluxClientBootstrapGitError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil).Return(errors.New("error in bootstrap git")).Times(5)
	tt.f.EXPECT().BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil).Return(nil).AnyTimes()

	tt.Expect(tt.c.BootstrapGit(tt.ctx, tt.cluster, tt.fluxConfig, nil)).To(MatchError(ContainSubstring("error in bootstrap git")), "fluxClient.BootstrapGit() should fail after 5 tries")
}

func TestFluxClientUninstallSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().Uninstall(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in uninstall")).Times(4)
	tt.f.EXPECT().Uninstall(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).Times(1)

	tt.Expect(tt.c.Uninstall(tt.ctx, tt.cluster, tt.fluxConfig)).To(Succeed(), "fluxClient.Uninstall() should succeed with 5 tries")
}

func TestFluxClientUninstallError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().Uninstall(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in uninstall")).Times(5)
	tt.f.EXPECT().Uninstall(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).AnyTimes()

	tt.Expect(tt.c.Uninstall(tt.ctx, tt.cluster, tt.fluxConfig)).To(MatchError(ContainSubstring("error in uninstall")), "fluxClient.Uninstall() should fail after 5 tries")
}

func TestFluxClientSuspendKustomizationSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in suspend kustomization")).Times(4)
	tt.f.EXPECT().SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).Times(1)

	tt.Expect(tt.c.SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig)).To(Succeed(), "fluxClient.SuspendKustomization() should succeed with 5 tries")
}

func TestFluxClientSuspendKustomizationError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in suspend kustomization")).Times(5)
	tt.f.EXPECT().SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).AnyTimes()

	tt.Expect(tt.c.SuspendKustomization(tt.ctx, tt.cluster, tt.fluxConfig)).To(MatchError(ContainSubstring("error in suspend kustomization")), "fluxClient.SuspendKustomization() should fail after 5 tries")
}

func TestFluxClientResumeKustomizationSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in resume kustomization")).Times(4)
	tt.f.EXPECT().ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).Times(1)

	tt.Expect(tt.c.ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig)).To(Succeed(), "fluxClient.ResumeKustomization() should succeed with 5 tries")
}

func TestFluxClientResumeKustomizationError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in resume kustomization")).Times(5)
	tt.f.EXPECT().ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).AnyTimes()

	tt.Expect(tt.c.ResumeKustomization(tt.ctx, tt.cluster, tt.fluxConfig)).To(MatchError(ContainSubstring("error in resume kustomization")), "fluxClient.ResumeKustomization() should fail after 5 tries")
}

func TestFluxClientReconcileSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in reconcile")).Times(4)
	tt.f.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).Times(1)

	tt.Expect(tt.c.Reconcile(tt.ctx, tt.cluster, tt.fluxConfig)).To(Succeed(), "fluxClient.Reconcile() should succeed with 5 tries")
}

func TestFluxClientReconcileError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.f.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.fluxConfig).Return(errors.New("error in reconcile")).Times(5)
	tt.f.EXPECT().Reconcile(tt.ctx, tt.cluster, tt.fluxConfig).Return(nil).AnyTimes()

	tt.Expect(tt.c.Reconcile(tt.ctx, tt.cluster, tt.fluxConfig)).To(MatchError(ContainSubstring("error in reconcile")), "fluxClient.Reconcile() should fail after 5 tries")
}

func TestFluxClientForceReconcileSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.k.EXPECT().UpdateAnnotation(tt.ctx, "gitrepositories", "flux-system", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error in force reconcile")).Times(4)
	tt.k.EXPECT().UpdateAnnotation(tt.ctx, "gitrepositories", "flux-system", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)

	tt.Expect(tt.c.ForceReconcile(tt.ctx, tt.cluster, "flux-system")).To(Succeed(), "fluxClient.ForceReconcile() should succeed with 5 tries")
}

func TestFluxClientForceReconcileError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.k.EXPECT().UpdateAnnotation(tt.ctx, "gitrepositories", "flux-system", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("error in force reconcile")).Times(5)
	tt.k.EXPECT().UpdateAnnotation(tt.ctx, "gitrepositories", "flux-system", gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	tt.Expect(tt.c.ForceReconcile(tt.ctx, tt.cluster, "flux-system")).To(MatchError(ContainSubstring("error in force reconcile")), "fluxClient.ForceReconcile() should fail after 5 tries")
}

func TestFluxClientDeleteSystemSecretSuccess(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.k.EXPECT().DeleteSecret(tt.ctx, tt.cluster, "flux-system", "custom-namespace").Return(errors.New("error in delete secret")).Times(4)
	tt.k.EXPECT().DeleteSecret(tt.ctx, tt.cluster, "flux-system", "custom-namespace").Return(nil).Times(1)

	tt.Expect(tt.c.DeleteSystemSecret(tt.ctx, tt.cluster, "custom-namespace")).To(Succeed(), "fluxClient.DeleteSystemSecret() should succeed with 5 tries")
}

func TestFluxClientDeleteSystemSecretError(t *testing.T) {
	tt := newFluxClientTest(t)
	tt.k.EXPECT().DeleteSecret(tt.ctx, tt.cluster, "flux-system", "custom-namespace").Return(errors.New("error in delete secret")).Times(5)
	tt.k.EXPECT().DeleteSecret(tt.ctx, tt.cluster, "flux-system", "custom-namespace").Return(nil).AnyTimes()

	tt.Expect(tt.c.DeleteSystemSecret(tt.ctx, tt.cluster, "custom-namespace")).To(MatchError(ContainSubstring("error in delete secret")), "fluxClient.DeleteSystemSecret() should fail after 5 tries")
}
