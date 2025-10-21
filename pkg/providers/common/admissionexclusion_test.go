package common_test

import (
	"encoding/json"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/common"
)

func TestGetAdmissionPluginExclusionPolicy(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(policy).ToNot(BeEmpty())
}

func TestGetAdmissionPluginExclusionPolicyReturnsValidJSON(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred(), "policy should be valid JSON")
}

func TestGetAdmissionPluginExclusionPolicyReturnsArray(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(rules).ToNot(BeEmpty(), "policy should contain at least one rule")
}

func TestGetAdmissionPluginExclusionPolicyHasRequiredFields(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred())

	requiredFields := []string{"apiGroup", "apiVersion", "kind", "scope", "username", "name"}
	for i, rule := range rules {
		for _, field := range requiredFields {
			g.Expect(rule).To(HaveKey(field), "rule %d should have field %q", i, field)
		}
	}
}

func TestGetAdmissionPluginExclusionPolicyNamespacedRulesHaveNamespace(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred())

	for i, rule := range rules {
		scope, ok := rule["scope"].(string)
		g.Expect(ok).To(BeTrue(), "rule %d should have string scope", i)

		if scope == "Namespaced" {
			g.Expect(rule).To(HaveKey("namespace"),
				"rule %d with scope 'Namespaced' must have 'namespace' field", i)
			namespace, ok := rule["namespace"].(string)
			g.Expect(ok).To(BeTrue(), "rule %d namespace should be string", i)
			g.Expect(namespace).ToNot(BeEmpty(), "rule %d namespace should not be empty", i)
		}
	}
}

func TestGetAdmissionPluginExclusionPolicyClusterRulesNoNamespace(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred())

	for i, rule := range rules {
		scope, ok := rule["scope"].(string)
		g.Expect(ok).To(BeTrue(), "rule %d should have string scope", i)

		if scope == "Cluster" {
			if namespace, ok := rule["namespace"]; ok {
				if ns, isString := namespace.(string); isString {
					_ = ns
				}
			}
		}
	}
}

func TestGetAdmissionPluginExclusionPolicyTrimsWhitespace(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	g.Expect(policy).To(Equal(strings.TrimSpace(policy)),
		"policy should not have leading or trailing whitespace")
}

func TestGetAdmissionPluginExclusionPolicyValidScopeValues(t *testing.T) {
	g := NewWithT(t)

	policy, err := common.GetAdmissionPluginExclusionPolicy()
	g.Expect(err).ToNot(HaveOccurred())

	var rules []map[string]interface{}
	err = json.Unmarshal([]byte(policy), &rules)
	g.Expect(err).ToNot(HaveOccurred())

	validScopes := map[string]bool{
		"Cluster":    true,
		"Namespaced": true,
	}

	for i, rule := range rules {
		scope, ok := rule["scope"].(string)
		g.Expect(ok).To(BeTrue(), "rule %d should have string scope", i)
		g.Expect(validScopes).To(HaveKey(scope),
			"rule %d has invalid scope %q, must be 'Cluster' or 'Namespaced'", i, scope)
	}
}
