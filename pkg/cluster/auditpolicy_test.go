package cluster_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	gomega "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestGetDefaultAuditPolicy(t *testing.T) {
	g := gomega.NewWithT(t)

	policyConfigMap, err := cluster.GetDefaultAuditPolicy()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(policyConfigMap).NotTo(gomega.BeNil())
	g.Expect(policyConfigMap.Data).To(gomega.HaveKey("audit-policy.yaml"))

	policy := policyConfigMap.Data["audit-policy.yaml"]
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

func TestAuditPolicyProcessingWithInvalidKind(t *testing.T) {
	g := gomega.NewWithT(t)

	config, err := cluster.ParseConfigFromFile("testdata/cluster_audit_policy_invalid_kind.yaml")
	g.Expect(err).NotTo(gomega.HaveOccurred())

	manager, err := cluster.NewDefaultConfigManager()
	g.Expect(err).NotTo(gomega.HaveOccurred())
	// Now validate - this should fail because auditPolicyConfigMapRef has kind: Secret instead of ConfigMap
	err = manager.Validate(config)
	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.ContainSubstring("AuditPolicyConfigMapRef.Kind must be ConfigMap"))
}

func TestGetAuditPolicyConfigMap(t *testing.T) {
	tests := []struct {
		name                string
		setupAuditPolicyRef bool
		auditPolicyContent  string
		expectConfigMap     bool
	}{
		{
			name:                "no audit policy ref should return configmap with default policy",
			setupAuditPolicyRef: false,
			auditPolicyContent:  "some-policy",
			expectConfigMap:     true,
		},
		{
			name:                "audit policy ref configured should return configmap",
			setupAuditPolicyRef: true,
			auditPolicyContent:  "custom-audit-policy-content",
			expectConfigMap:     true,
		},
		{
			name:                "audit policy ref with empty content should return configmap",
			setupAuditPolicyRef: true,
			auditPolicyContent:  "",
			expectConfigMap:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)

			spec := &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &anywherev1.Cluster{
						ObjectMeta: metav1.ObjectMeta{
							Name: "test",
						},
						Spec: anywherev1.ClusterSpec{
							ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{},
						},
					},
					AuditPolicyConfigMap: &corev1.ConfigMap{
						Data: map[string]string{
							"audit-policy.yaml": tt.auditPolicyContent,
						},
					},
				},
			}

			if tt.setupAuditPolicyRef {
				spec.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef = &anywherev1.Ref{
					Name: "test-audit-policy",
					Kind: "ConfigMap",
				}
			}

			configMap := cluster.GetAuditPolicyConfigMap(spec)

			if tt.expectConfigMap {
				g.Expect(configMap).NotTo(gomega.BeNil())
				g.Expect(configMap.Name).To(gomega.Equal(fmt.Sprintf("%s-%s", spec.Cluster.Name, constants.AuditPolicyConfigMapName)))
				g.Expect(configMap.Namespace).To(gomega.Equal(constants.EksaSystemNamespace))
				g.Expect(configMap.Kind).To(gomega.Equal("ConfigMap"))
				g.Expect(configMap.APIVersion).To(gomega.Equal("v1"))
				g.Expect(configMap.Data).To(gomega.HaveKey("audit-policy.yaml"))
				g.Expect(configMap.Data["audit-policy.yaml"]).To(gomega.Equal(tt.auditPolicyContent))
			} else {
				g.Expect(configMap).To(gomega.BeNil())
			}
		})
	}
}

func TestAuditPolicyProcessing(t *testing.T) {
	tests := []struct {
		name                   string
		configPath             string
		wantAuditPolicyRef     *anywherev1.Ref
		wantAuditPolicyContent string
	}{
		{
			name:       "cluster with audit policy ConfigMap",
			configPath: "testdata/cluster_audit_policy.yaml",
			wantAuditPolicyRef: &anywherev1.Ref{
				Kind: "ConfigMap",
				Name: "custom-audit-policy",
			},
			wantAuditPolicyContent: `apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: RequestResponse
  verbs: ["update", "patch", "delete"]
  resources:
  - group: ""
    resources: ["configmaps"]
    resourceNames: ["aws-auth"]
  namespaces: ["kube-system"]
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := gomega.NewWithT(t)
			config, err := cluster.ParseConfigFromFile(tt.configPath)
			g.Expect(err).NotTo(gomega.HaveOccurred())

			// Verify the audit policy reference is set correctly
			g.Expect(config.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef).To(gomega.Equal(tt.wantAuditPolicyRef))

			// Verify the audit policy content was processed correctly
			g.Expect(config.AuditPolicy()).To(gomega.Equal(tt.wantAuditPolicyContent))
		})
	}
}

func TestDefaultConfigClientBuilderAuditPolicyConfig(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				AuditPolicyConfigMapRef: &anywherev1.Ref{
					Kind: "ConfigMap",
					Name: "my-audit-policy",
				},
			},
		},
	}
	auditPolicyCmName := cluster.Name + "-" + constants.AuditPolicyConfigMapName
	auditPolicyConfigMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      auditPolicyCmName,
			Namespace: constants.EksaSystemNamespace,
		},
		Data: map[string]string{
			"audit-policy.yaml": `apiVersion: audit.k8s.io/v1
kind: Policy
rules:
- level: Metadata
  resources:
  - group: ""
    resources: ["secrets", "configmaps"]
`,
		},
	}

	client.EXPECT().Get(ctx, auditPolicyCmName, constants.EksaSystemNamespace, &corev1.ConfigMap{}).DoAndReturn(
		func(_ context.Context, _, _ string, obj runtime.Object) error {
			c := obj.(*corev1.ConfigMap)
			c.ObjectMeta = auditPolicyConfigMap.ObjectMeta
			c.Data = auditPolicyConfigMap.Data
			return nil
		},
	)

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(config).NotTo(gomega.BeNil())
	g.Expect(config.Cluster).To(gomega.Equal(cluster))
	g.Expect(config.AuditPolicy()).To(gomega.Equal(auditPolicyConfigMap.Data["audit-policy.yaml"]))
}

func TestDefaultConfigClientBuilderNoAuditPolicyConfig(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	b := cluster.NewDefaultConfigClientBuilder()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				// No audit policy config map ref
			},
		},
	}

	config, err := b.Build(ctx, client, cluster)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(config).NotTo(gomega.BeNil())
	g.Expect(config.Cluster).To(gomega.Equal(cluster))

	// Should have default audit policy when no custom one is provided
	g.Expect(config.AuditPolicy()).NotTo(gomega.BeEmpty())
	g.Expect(config.AuditPolicy()).To(gomega.ContainSubstring("apiVersion: audit.k8s.io/v1"))
	g.Expect(config.AuditPolicy()).To(gomega.ContainSubstring("kind: Policy"))
}
