package executables_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/files"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

func TestLocalExecutablesBuilderAllExecutables(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	b := executables.NewLocalExecutablesBuilder()
	closer, err := b.Init(ctx)
	g.Expect(err).NotTo(HaveOccurred())
	_, writer := test.NewWriter(t)
	reader := files.NewReader()

	kind := b.BuildKindExecutable(writer)
	g.Expect(kind).NotTo(BeNil())
	awsAdm := b.BuildClusterAwsAdmExecutable()
	g.Expect(awsAdm).NotTo(BeNil())
	clusterctl := b.BuildClusterCtlExecutable(writer, reader)
	g.Expect(clusterctl).NotTo(BeNil())
	kubectl := b.BuildKubectlExecutable()
	g.Expect(kubectl).NotTo(BeNil())
	govc := b.BuildGovcExecutable(writer)
	g.Expect(govc).NotTo(BeNil())
	cmk, err := b.BuildCmkExecutable(writer, &decoder.CloudStackExecConfig{
		Profiles: make([]decoder.CloudStackProfileConfig, 0),
	})
	g.Expect(cmk).NotTo(BeNil())
	aws := b.BuildAwsCli()
	g.Expect(aws).NotTo(BeNil())
	flux := b.BuildFluxExecutable()
	g.Expect(flux).NotTo(BeNil())
	trouble := b.BuildTroubleshootExecutable()
	g.Expect(trouble).NotTo(BeNil())
	helm := b.BuildHelmExecutable()
	g.Expect(helm).NotTo(BeNil())
	docker := b.BuildDockerExecutable()
	g.Expect(docker).NotTo(BeNil())
	ssh := b.BuildSSHExecutable()
	g.Expect(ssh).NotTo(BeNil())

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(closer(ctx)).To(Succeed())
}

func TestExecutablesInDocker(t *testing.T) {
	tests := []struct {
		name        string
		envVarValue string
		want        bool
	}{
		{
			name:        "true",
			envVarValue: "true",
			want:        false,
		},
		{
			name:        "false",
			envVarValue: "false",
			want:        true,
		},
		{
			name:        "not set",
			envVarValue: "",
			want:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envVarValue != "" {
				t.Setenv("MR_TOOLS_DISABLE", tt.envVarValue)
			}
			g := NewWithT(t)
			g.Expect(executables.ExecutablesInDocker()).To(Equal(tt.want))
		})
	}
}

func TestInDockerExecutablesBuilder(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	image := "image"
	ctrl := gomock.NewController(t)
	c := mocks.NewMockDockerClient(ctrl)
	c.EXPECT().PullImage(ctx, image)
	c.EXPECT().Execute(ctx, gomock.Any()) // Init container
	c.EXPECT().Execute(ctx, gomock.Any()) // Remove container

	b, err := executables.NewInDockerExecutablesBuilder(c, image)
	g.Expect(err).NotTo(HaveOccurred())
	closer, err := b.Init(ctx)
	h := b.BuildHelmExecutable()
	g.Expect(h).NotTo(BeNil())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(closer(ctx)).To(Succeed())
}

func TestLocalExecutablesBuilder(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	b := executables.NewLocalExecutablesBuilder()
	closer, err := b.Init(ctx)
	h := b.BuildHelmExecutable()
	g.Expect(h).NotTo(BeNil())
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(closer(ctx)).To(Succeed())
}
