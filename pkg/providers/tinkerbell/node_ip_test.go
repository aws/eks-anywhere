package tinkerbell

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func TestIPv6NodeIPBootstrap(t *testing.T) {
	for _, osFamily := range []v1alpha1.OSFamily{v1alpha1.Ubuntu, v1alpha1.RedHat} {
		t.Run(string(osFamily), func(t *testing.T) {
			g := NewWithT(t)
			spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
			spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "2001:db8::10"
			spec.Cluster.Spec.ClusterNetwork.Pods.CidrBlocks = []string{"2001:db8:100::/56"}
			spec.Cluster.Spec.ClusterNetwork.Services.CidrBlocks = []string{"2001:db8:200::/112"}
			spec.TinkerbellDatacenter.Spec.TinkerbellIP = "2001:db8::11"
			spec.TinkerbellDatacenter.Spec.DHCPv6Config = &v1alpha1.DHCPv6Config{
				DerivedAddressPool: "2001:db8::/64",
			}
			for _, machine := range spec.TinkerbellMachineConfigs {
				machine.Spec.OSFamily = osFamily
			}

			cp, err := ControlPlaneSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.Files).To(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
			command := ipv6NodeIPScriptPath + " 2001:db8::11 2001:db8::10"
			g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.PreKubeadmCommands[0]).To(Equal(command))

			workers, err := WorkersSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
			g.Expect(err).NotTo(HaveOccurred())
			for _, worker := range workers.Groups {
				g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.Files).To(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
				g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.PreKubeadmCommands[0]).To(Equal(command))
			}
		})
	}
}

func TestIPv4BootstrapDoesNotAddIPv6NodeIPHelper(t *testing.T) {
	g := NewWithT(t)
	spec := test.NewFullClusterSpec(t, testClusterConfigFilename)
	cp, err := ControlPlaneSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(cp.KubeadmControlPlane.Spec.KubeadmConfigSpec.Files).NotTo(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
	workers, err := WorkersSpec(context.Background(), test.NewNullLogger(), test.NewFakeKubeClient(), spec)
	g.Expect(err).NotTo(HaveOccurred())
	for _, worker := range workers.Groups {
		g.Expect(worker.KubeadmConfigTemplate.Spec.Template.Spec.Files).NotTo(ContainElement(HaveField("Path", ipv6NodeIPScriptPath)))
	}
}
