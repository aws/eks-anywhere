package docker_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/docker"
)

func TestMachineTemplateEqualDifferentNames(t *testing.T) {
	g := NewWithT(t)

	machineTemplate := dockerMachineTemplate("test-machine-1")
	otherMachineTemplate := machineTemplate.DeepCopy()
	otherMachineTemplate.Name = "test-machine-2"

	isEqual := docker.MachineTemplateEqual(machineTemplate, otherMachineTemplate)
	g.Expect(isEqual).To(BeTrue())
}

func TestMachineTemplateEqualDifferentCustomImages(t *testing.T) {
	g := NewWithT(t)

	machineTemplate := dockerMachineTemplate("test-machine-1")
	otherMachineTemplate := machineTemplate.DeepCopy()
	otherMachineTemplate.Spec.Template.Spec.CustomImage = "other-custom-image"

	g.Expect(docker.MachineTemplateEqual(machineTemplate, otherMachineTemplate)).To(BeFalse())
}

func TestGetMachineTemplateNoError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	machineTemplateName := "test-machine-1"
	machineTemplate := dockerMachineTemplate(machineTemplateName)
	client := test.NewFakeKubeClient(
		machineTemplate,
	)

	m, err := docker.GetMachineTemplate(ctx, client, machineTemplateName, constants.EksaSystemNamespace)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(m).NotTo(BeNil())
}

func TestGetMachineTemplateErrorFromClient(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	client := test.NewFakeKubeClient()

	_, err := docker.GetMachineTemplate(ctx, client, "test-machine-1", constants.EksaSystemNamespace)
	g.Expect(err).To(MatchError(ContainSubstring("reading dockerMachineTemplate")))
}
