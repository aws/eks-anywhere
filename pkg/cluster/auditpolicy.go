package cluster

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func auditPolicyEntry() *ConfigManagerEntry {
	return &ConfigManagerEntry{
		APIObjectMapping: map[string]APIObjectGenerator{
			constants.ConfigMapKind: func() APIObject {
				return &corev1.ConfigMap{}
			},
		},
		Processors: []ParsedProcessor{processAuditPolicy},
		Defaulters: []Defaulter{
			setDefaultAuditPolicy,
		},
		Validations: []Validation{validateAuditPolicy},
	}
}

func processAuditPolicy(c *Config, objects ObjectLookup) {
	if c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef == nil {
		return
	}

	configMap := objects.GetFromRef("v1", *c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef)
	if configMap == nil {
		return
	}

	cm := configMap.(*corev1.ConfigMap)
	c.AuditPolicyConfigMap = cm
}

func setDefaultAuditPolicy(c *Config) error {
	if c.AuditPolicyConfigMap == nil {
		defaultPolicy, err := GetDefaultAuditPolicy(c.Cluster)
		if err != nil {
			return err
		}
		c.AuditPolicyConfigMap = defaultPolicy
	}
	return nil
}

func validateAuditPolicy(c *Config) error {
	if c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef == nil {
		return nil
	}

	configMapRef := c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef

	if configMapRef.Name == "" {
		return errors.New("AuditPolicyConfigMapRef.Name is required when AuditPolicyConfigMapRef is provided")
	}

	if configMapRef.Kind != constants.ConfigMapKind && configMapRef.Kind != "" {
		return fmt.Errorf("AuditPolicyConfigMapRef.Kind must be %s, got %s",
			constants.ConfigMapKind, configMapRef.Kind)
	}

	return nil
}

func getAuditPolicy(ctx context.Context, client Client, c *Config) error {
	if c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef == nil {
		defaultPolicy, err := GetDefaultAuditPolicy(c.Cluster)
		if err != nil {
			return err
		}
		c.AuditPolicyConfigMap = defaultPolicy
		return nil
	}

	configMap := &corev1.ConfigMap{}
	configMapName := c.Cluster.Spec.ControlPlaneConfiguration.AuditPolicyConfigMapRef.Name
	if err := client.Get(ctx, configMapName, c.Cluster.Namespace, configMap); err != nil {
		return err
	}
	fmt.Println("--------audit policy config map found!!!!----------")
	fmt.Println(configMap)
	fmt.Printf("config map name from getauditpolicy: %s\n", configMap.Name)
	c.AuditPolicyConfigMap = configMap
	return nil
}

// GetAuditPolicyConfigMap returns configMap if AuditPolicyConfigMapRef is not nil.
func GetAuditPolicyConfigMap(spec *Spec) *corev1.ConfigMap {
	configMapName := fmt.Sprintf("%s-%s", spec.Cluster.Name, constants.AuditPolicyConfigMapName)
	auditPolicyConfigMap := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: spec.Cluster.Namespace,
			// Remove owner references - they should be set by the controller
		},
		Data: map[string]string{
			"audit-policy.yaml": spec.AuditPolicy(),
		},
	}
	fmt.Println("--------inside GetAuditPolicyConfigMap----------")
	fmt.Printf("%+v\n", auditPolicyConfigMap.Data)
	return auditPolicyConfigMap
}

// GetDefaultAuditPolicy returns the default audit policy as a ConfigMap.
func GetDefaultAuditPolicy(cluster *v1alpha1.Cluster) (*corev1.ConfigMap, error) {
	auditPolicyv1, err := auditPolicyV1Yaml()
	if err != nil {
		return nil, err
	}

	configMapName := fmt.Sprintf("%s-%s", cluster.Name, constants.AuditPolicyConfigMapName)
	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMapName,
			Namespace: cluster.Namespace,
		},
		Data: map[string]string{
			"audit-policy.yaml": strings.TrimSpace(string(auditPolicyv1)),
		},
	}, nil
}

// auditPolicyV1Yaml returns the byte array for yaml created with v1 api version for audit policy.
func auditPolicyV1Yaml() ([]byte, error) {
	auditPolicy := auditPolicyV1()
	return yaml.Marshal(auditPolicy)
}

// auditPolicyV1 returns the v1 audit policy.
func auditPolicyV1() *auditv1.Policy {
	return &auditv1.Policy{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Policy",
			APIVersion: "audit.k8s.io/v1",
		},
		Rules: []auditv1.PolicyRule{
			{
				Level: auditv1.Level("RequestResponse"),
				Verbs: []string{
					"update",
					"patch",
					"delete",
				},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"configmaps",
						},
						ResourceNames: []string{
							"aws-auth",
						},
					},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
				Namespaces: []string{"kube-system"},
			},
			{
				Level: auditv1.Level("None"),
				Users: []string{"system:kube-proxy"},
				Verbs: []string{"watch"},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"endpoints",
							"services",
							"services/status",
						},
					},
				},
			},
			{
				Level: auditv1.Level("None"),
				Users: []string{"kubelet"},
				Verbs: []string{"get"},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"nodes",
							"nodes/status",
						},
					},
				},
			},
			{
				Level: auditv1.Level("None"),
				Verbs: []string{"get"},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"nodes",
							"nodes/status",
						},
					},
				},
			},
			{
				Level: auditv1.Level("None"),
				Users: []string{
					"system:kube-controller-manager",
					"system:kube-scheduler",
					"system:serviceaccount:kube-system:endpoint-controller",
				},
				Verbs: []string{
					"get",
					"update",
				},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{"endpoints"},
					},
				},
				Namespaces: []string{"kube-system"},
			},
			{
				Level: auditv1.Level("None"),
				Users: []string{"system:apiserver"},
				Verbs: []string{"get"},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"namespaces",
							"namespaces/status",
							"namespaces/finalize",
						},
					},
				},
			},
			{
				Level: auditv1.Level("None"),
				Users: []string{"system:kube-controller-manager"},
				Verbs: []string{
					"get",
					"list",
				},
				Resources: []auditv1.GroupResources{
					{
						Group: "metrics.k8s.io",
					},
				},
			},
			{
				Level: auditv1.Level("None"),
				NonResourceURLs: []string{
					"/healthz*",
					"/version",
					"/swagger*",
				},
			},
			{
				Level: auditv1.Level("None"),
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{"events"},
					},
				},
			},
			{
				Level: auditv1.Level("Request"),
				Users: []string{
					"kubelet",
					"system:node-problem-detector",
					"system:serviceaccount:kube-system:node-problem-detector",
				},
				Verbs: []string{
					"update",
					"patch",
				},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"nodes/status",
							"pods/status",
						},
					},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
			{
				Level: auditv1.Level("Request"),
				Verbs: []string{
					"update",
					"patch",
				},
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{
							"nodes/status",
							"pods/status",
						},
					},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
				UserGroups: []string{
					"system:nodes",
				},
			},
			{
				Level: auditv1.Level("Request"),
				Users: []string{"system:serviceaccount:kube-system:namespace-controller"},
				Verbs: []string{"deletecollection"},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
			{
				Level: auditv1.Level("Metadata"),
				Resources: []auditv1.GroupResources{
					{Resources: []string{
						"secrets",
						"configmaps",
					}},
					{
						Group:     "authentication.k8s.io",
						Resources: []string{"tokenreviews"},
					},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
			{
				Level: auditv1.Level("Request"),
				Resources: []auditv1.GroupResources{
					{
						Resources: []string{"serviceaccounts/token"},
					},
				},
			},
			{
				Level: auditv1.Level("Request"),
				Verbs: []string{
					"get",
					"list",
					"watch",
				},
				Resources: []auditv1.GroupResources{
					{Group: ""},
					{Group: "admissionregistration.k8s.io"},
					{Group: "apiextensions.k8s.io"},
					{Group: "apiregistration.k8s.io"},
					{Group: "apps"},
					{Group: "authentication.k8s.io"},
					{Group: "authorization.k8s.io"},
					{Group: "autoscaling"},
					{Group: "batch"},
					{Group: "certificates.k8s.io"},
					{Group: "extensions"},
					{Group: "metrics.k8s.io"},
					{Group: "networking.k8s.io"},
					{Group: "policy"},
					{Group: "rbac.authorization.k8s.io"},
					{Group: "scheduling.k8s.io"},
					{Group: "settings.k8s.io"},
					{Group: "storage.k8s.io"},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
			{
				Level: auditv1.Level("RequestResponse"),
				Resources: []auditv1.GroupResources{
					{Group: ""},
					{Group: "admissionregistration.k8s.io"},
					{Group: "apiextensions.k8s.io"},
					{Group: "apiregistration.k8s.io"},
					{Group: "apps"},
					{Group: "authentication.k8s.io"},
					{Group: "authorization.k8s.io"},
					{Group: "autoscaling"},
					{Group: "batch"},
					{Group: "certificates.k8s.io"},
					{Group: "extensions"},
					{Group: "metrics.k8s.io"},
					{Group: "networking.k8s.io"},
					{Group: "policy"},
					{Group: "rbac.authorization.k8s.io"},
					{Group: "scheduling.k8s.io"},
					{Group: "settings.k8s.io"},
					{Group: "storage.k8s.io"},
				},
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
			{
				Level: auditv1.Level("Metadata"),
				OmitStages: []auditv1.Stage{
					"RequestReceived",
				},
			},
		},
	}
}
