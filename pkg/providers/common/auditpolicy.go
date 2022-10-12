package common

import (
	auditv1 "k8s.io/apiserver/pkg/apis/audit/v1"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

func GetAuditPolicyYaml(kubeVersion anywherev1.KubernetesVersion) ([]byte, error) {
	auditPolicy, err := GetAuditPolicy(kubeVersion)
	if err != nil {
		return nil, err
	}
	return yaml.Marshal(auditPolicy)
}

func GetAuditPolicy(kubeVersion anywherev1.KubernetesVersion) (*auditv1.Policy, error) {
	a := baseAuditPolicy()
	v1policy, err := anywherev1.GreaterOrEqual(kubeVersion, anywherev1.Kube124)
	if err != nil {
		return nil, err
	}

	if v1policy {
		a.APIVersion = "audit.k8s.io/v1"
	} else {
		a.APIVersion = "audit.k8s.io/v1beta1"
	}

	return a, nil
}

func baseAuditPolicy() *auditv1.Policy {
	return &auditv1.Policy{
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
