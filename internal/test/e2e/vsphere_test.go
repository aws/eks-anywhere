package e2e

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestE2ESessionsetupVSphereEnv(t *testing.T) {
	g := NewWithT(t)
	session := E2ESession{
		testEnvVars: map[string]string{},
	}

	t.Setenv("T_VSPHERE_DATACENTER", "my-datacenter")
	t.Setenv("T_VSPHERE_TEMPLATE_UBUNTU_KUBERNETES_1_20_EKS_19", "template-1-20")
	t.Setenv("T_VSPHERE_TEMPLATE_UBUNTU_KUBERNETES_1_22_EKS_9", "template-1-22")
	t.Setenv("T_VSPHERE_1_22_EKS_9", "template-1-22") // shouldn't be added

	g.Expect(session.setupVSphereEnv("TestVSphere")).To(Succeed())

	g.Expect(session.testEnvVars).To(HaveKeyWithValue("T_VSPHERE_DATACENTER", "my-datacenter"))
	g.Expect(session.testEnvVars).To(HaveKeyWithValue("T_VSPHERE_TEMPLATE_UBUNTU_KUBERNETES_1_20_EKS_19", "template-1-20"))
	g.Expect(session.testEnvVars).To(HaveKeyWithValue("T_VSPHERE_TEMPLATE_UBUNTU_KUBERNETES_1_22_EKS_9", "template-1-22"))
	g.Expect(session.testEnvVars).NotTo(HaveKey("T_VSPHERE_1_22_EKS_9"))
}
