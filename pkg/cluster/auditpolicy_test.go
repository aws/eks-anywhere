package cluster_test

import (
	"strings"
	"testing"

	gomega "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestGetDefaultAuditPolicy(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(policy).NotTo(gomega.BeEmpty())

	// Verify it's valid YAML with expected structure
	g.Expect(policy).To(gomega.ContainSubstring("apiVersion: audit.k8s.io/v1"))
	g.Expect(policy).To(gomega.ContainSubstring("kind: Policy"))
	g.Expect(policy).To(gomega.ContainSubstring("rules:"))

	// Verify some key audit rules are present
	g.Expect(policy).To(gomega.ContainSubstring("aws-auth"))
	g.Expect(policy).To(gomega.ContainSubstring("system:kube-proxy"))
	g.Expect(policy).To(gomega.ContainSubstring("RequestResponse"))
	g.Expect(policy).To(gomega.ContainSubstring("Metadata"))

	// Verify the policy doesn't have leading/trailing whitespace
	g.Expect(strings.TrimSpace(policy)).To(gomega.Equal(policy))
}

func TestGetDefaultAuditPolicyStructure(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Verify specific audit rules are present
	expectedRules := []string{
		"configmaps",
		"aws-auth",
		"system:kube-proxy",
		"kubelet",
		"system:kube-controller-manager",
		"system:kube-scheduler",
		"system:apiserver",
		"RequestResponse",
		"None",
		"Request",
		"Metadata",
	}

	for _, rule := range expectedRules {
		g.Expect(policy).To(gomega.ContainSubstring(rule), "Policy should contain rule: %s", rule)
	}
}

func TestGetDefaultAuditPolicyConsistency(t *testing.T) {
	g := gomega.NewWithT(t)

	// Call the function multiple times to ensure consistency
	policy1, err1 := cluster.GetDefaultAuditPolicy()
	g.Expect(err1).NotTo(gomega.HaveOccurred())

	policy2, err2 := cluster.GetDefaultAuditPolicy()
	g.Expect(err2).NotTo(gomega.HaveOccurred())

	// Both calls should return identical policies
	g.Expect(policy1).To(gomega.Equal(policy2))
}

func TestGetDefaultAuditPolicyYAMLValidity(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Basic YAML structure validation
	lines := strings.Split(policy, "\n")
	g.Expect(len(lines)).To(gomega.BeNumerically(">", 10), "Policy should have multiple lines")

	// Should start with apiVersion
	g.Expect(lines[0]).To(gomega.Equal("apiVersion: audit.k8s.io/v1"))

	// Should have kind on second line
	g.Expect(lines[1]).To(gomega.Equal("kind: Policy"))

	// Should contain rules section
	foundRules := false
	for _, line := range lines {
		if strings.TrimSpace(line) == "rules:" {
			foundRules = true
			break
		}
	}
	g.Expect(foundRules).To(gomega.BeTrue(), "Policy should contain 'rules:' section")
}

func TestGetDefaultAuditPolicySpecificRules(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Test for specific audit policy rules that should be present
	tests := []struct {
		name        string
		shouldExist bool
		content     string
	}{
		{
			name:        "aws-auth configmap rule",
			shouldExist: true,
			content:     "aws-auth",
		},
		{
			name:        "kube-proxy rule",
			shouldExist: true,
			content:     "system:kube-proxy",
		},
		{
			name:        "kubelet rule",
			shouldExist: true,
			content:     "kubelet",
		},
		{
			name:        "controller manager rule",
			shouldExist: true,
			content:     "system:kube-controller-manager",
		},
		{
			name:        "scheduler rule",
			shouldExist: true,
			content:     "system:kube-scheduler",
		},
		{
			name:        "apiserver rule",
			shouldExist: true,
			content:     "system:apiserver",
		},
		{
			name:        "healthz endpoints",
			shouldExist: true,
			content:     "/healthz",
		},
		{
			name:        "events resource",
			shouldExist: true,
			content:     "events",
		},
		{
			name:        "secrets resource",
			shouldExist: true,
			content:     "secrets",
		},
		{
			name:        "RequestResponse level",
			shouldExist: true,
			content:     "RequestResponse",
		},
		{
			name:        "Metadata level",
			shouldExist: true,
			content:     "Metadata",
		},
		{
			name:        "Request level",
			shouldExist: true,
			content:     "Request",
		},
		{
			name:        "None level",
			shouldExist: true,
			content:     "None",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			if tt.shouldExist {
				g.Expect(policy).To(gomega.ContainSubstring(tt.content), "Policy should contain: %s", tt.content)
			} else {
				g.Expect(policy).NotTo(gomega.ContainSubstring(tt.content), "Policy should not contain: %s", tt.content)
			}
		})
	}
}

func TestGetDefaultAuditPolicyRuleCount(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Count the number of "level:" occurrences as a proxy for rule count
	levelCount := strings.Count(policy, "level:")
	g.Expect(levelCount).To(gomega.BeNumerically(">", 10), "Policy should have multiple audit rules")
	g.Expect(levelCount).To(gomega.BeNumerically("<", 50), "Policy should have reasonable number of rules")
}

func TestGetDefaultAuditPolicyNamespaces(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Verify that kube-system namespace is referenced
	g.Expect(policy).To(gomega.ContainSubstring("kube-system"))

	// Verify namespace-related rules
	g.Expect(policy).To(gomega.ContainSubstring("namespaces"))
}

func TestGetDefaultAuditPolicyVerbs(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Verify common Kubernetes verbs that are actually present in the audit policy
	expectedVerbs := []string{
		"get",
		"list",
		"watch",
		"update",
		"patch",
		"delete",
		"deletecollection",
	}

	for _, verb := range expectedVerbs {
		g.Expect(policy).To(gomega.ContainSubstring(verb), "Policy should contain verb: %s", verb)
	}
}

func TestGetDefaultAuditPolicyOmitStages(t *testing.T) {
	g := gomega.NewWithT(t)

	policy, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	// Verify that omitStages is used appropriately
	g.Expect(policy).To(gomega.ContainSubstring("omitStages"))
	g.Expect(policy).To(gomega.ContainSubstring("RequestReceived"))
}
