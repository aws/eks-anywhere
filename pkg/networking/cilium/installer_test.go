package cilium_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/networking/cilium"
)

func TestInstallerInstallErrorGeneratingManifest(t *testing.T) {
	tt := newCiliumTest(t)
	installer := cilium.NewInstaller(tt.client, tt.installTemplater)
	tt.installTemplater.EXPECT().GenerateManifest(
		tt.ctx, tt.spec, gomock.Not(gomock.Nil()),
	).Return(nil, errors.New("generating manifest"))

	tt.Expect(
		installer.Install(tt.ctx, tt.cluster, tt.spec, nil),
	).To(
		MatchError(ContainSubstring("generating Cilium manifest for install: generating manifest")),
	)
}

func TestInstallerInstallErrorApplyingManifest(t *testing.T) {
	tt := newCiliumTest(t)
	installer := cilium.NewInstaller(tt.client, tt.installTemplater)
	tt.installTemplater.EXPECT().GenerateManifest(
		tt.ctx, tt.spec, gomock.Not(gomock.Nil()),
	).Return(tt.ciliumValues, nil)

	tt.client.EXPECT().Apply(
		tt.ctx, tt.cluster, tt.ciliumValues,
	).Return(errors.New("applying"))

	tt.Expect(
		installer.Install(tt.ctx, tt.cluster, tt.spec, nil),
	).To(
		MatchError(ContainSubstring("applying Cilium manifest for install: applying")),
	)
}

func TestInstallerInstallSuccess(t *testing.T) {
	tt := newCiliumTest(t)
	installer := cilium.NewInstaller(tt.client, tt.installTemplater)
	tt.installTemplater.EXPECT().GenerateManifest(
		tt.ctx, tt.spec, gomock.Not(gomock.Nil()),
	).Return(tt.ciliumValues, nil)

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster, tt.ciliumValues)

	tt.Expect(
		installer.Install(tt.ctx, tt.cluster, tt.spec, nil),
	).To(Succeed())
}

func TestInstallForSpecInstallSuccess(t *testing.T) {
	tt := newCiliumTest(t)
	config := cilium.Config{
		Spec:              tt.spec,
		AllowedNamespaces: []string{"my-namespace"},
	}
	installer := cilium.NewInstallerForSpec(tt.client, tt.installTemplater, config)
	tt.installTemplater.EXPECT().GenerateManifest(
		tt.ctx, tt.spec, gomock.Not(gomock.Nil()),
	).Return(tt.ciliumValues, nil)

	tt.client.EXPECT().Apply(tt.ctx, tt.cluster, tt.ciliumValues)

	tt.Expect(
		installer.Install(tt.ctx, tt.cluster),
	).To(Succeed())
}
