package vsphere

import (
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

const (
	// IPAMProviderNamespace is the namespace where the IPAM provider is deployed.
	IPAMProviderNamespace = "capi-ipam-provider"

	// IPAMProviderImage is the image for the IPAM provider controller.
	// NOTE: This version is pinned and should be updated when upgrading CAPI or when
	// a newer version of the IPAM provider is released. The version must be compatible
	// with the CAPI version used by EKS-Anywhere.
	// - v1.0.x uses the v1alpha2 API (ipam.cluster.x-k8s.io/v1alpha2)
	// - Ensure compatibility with CAPV and core CAPI versions before upgrading
	// See: https://github.com/kubernetes-sigs/cluster-api-ipam-provider-in-cluster
	IPAMProviderImage = "registry.k8s.io/capi-ipam-ic/cluster-api-ipam-in-cluster-controller:v1.0.3"

	// IPAMServiceAccountName is the name of the IPAM provider service account.
	IPAMServiceAccountName = "caip-in-cluster-controller-manager"
	// IPAMClusterRoleName is the name of the IPAM provider cluster role.
	IPAMClusterRoleName = "caip-in-cluster-manager-role"
	// IPAMDeploymentName is the name of the IPAM provider deployment.
	IPAMDeploymentName = "caip-in-cluster-controller-manager"
)

// InClusterIPPoolObject creates an InClusterIPPool unstructured object from the IPPoolConfiguration.
func InClusterIPPoolObject(ipPool *anywherev1.IPPoolConfiguration, namespace string) (client.Object, error) {
	if ipPool == nil {
		return nil, fmt.Errorf("ipPool configuration is nil")
	}

	if namespace == "" {
		namespace = constants.EksaSystemNamespace
	}

	// Convert addresses to interface slice for unstructured
	addresses := make([]interface{}, len(ipPool.Addresses))
	for i, addr := range ipPool.Addresses {
		addresses[i] = addr
	}

	pool := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "ipam.cluster.x-k8s.io/v1alpha2",
			"kind":       "InClusterIPPool",
			"metadata": map[string]interface{}{
				"name":      ipPool.Name,
				"namespace": namespace,
			},
			"spec": map[string]interface{}{
				"addresses": addresses,
				"prefix":    int64(ipPool.Prefix),
				"gateway":   ipPool.Gateway,
			},
		},
	}

	return pool, nil
}

// IPAMProviderObjects returns all the Kubernetes objects needed to deploy the CAPI IPAM provider.
// Note: This function is kept for reference but the controller no longer uses it.
// The CLI installs the IPAM provider using the embedded ipam-provider.yaml manifest.
func IPAMProviderObjects() ([]client.Object, error) {
	objects := []client.Object{}

	// Namespace
	objects = append(objects, ipamNamespace())

	// CRDs
	objects = append(objects, inClusterIPPoolCRD())
	objects = append(objects, globalInClusterIPPoolCRD())

	// RBAC
	objects = append(objects, ipamServiceAccount())
	objects = append(objects, ipamClusterRole())
	objects = append(objects, ipamClusterRoleBinding())

	// Deployment
	objects = append(objects, ipamDeployment())

	return objects, nil
}

func ipamNamespace() *corev1.Namespace {
	return &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: IPAMProviderNamespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
			},
		},
	}
}

func inClusterIPPoolCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "inclusterippools.ipam.cluster.x-k8s.io",
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
				"cluster.x-k8s.io/v1beta1":  "v1beta1",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "ipam.cluster.x-k8s.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "InClusterIPPool",
				ListKind: "InClusterIPPoolList",
				Plural:   "inclusterippools",
				Singular: "inclusterippool",
			},
			Scope: apiextensionsv1.NamespaceScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha2",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Description: "InClusterIPPool is the Schema for the inclusterippools API.",
							Type:        "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"apiVersion": {Type: "string"},
								"kind":       {Type: "string"},
								"metadata":   {Type: "object"},
								"spec": {
									Description: "InClusterIPPoolSpec defines the desired state of InClusterIPPool.",
									Type:        "object",
									Required:    []string{"addresses", "prefix", "gateway"},
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"addresses": {
											Type:        "array",
											Description: "Addresses is a list of IP addresses or CIDR blocks.",
											Items:       &apiextensionsv1.JSONSchemaPropsOrArray{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
										},
										"prefix": {
											Type:        "integer",
											Description: "Prefix is the subnet prefix.",
										},
										"gateway": {
											Type:        "string",
											Description: "Gateway is the gateway IP address.",
										},
									},
								},
								"status": {
									Description: "InClusterIPPoolStatus defines the observed state of InClusterIPPool.",
									Type:        "object",
								},
							},
						},
					},
					Served:  true,
					Storage: true,
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
					AdditionalPrinterColumns: []apiextensionsv1.CustomResourceColumnDefinition{
						{Name: "Addresses", Type: "string", JSONPath: ".spec.addresses"},
						{Name: "Prefix", Type: "integer", JSONPath: ".spec.prefix"},
						{Name: "Gateway", Type: "string", JSONPath: ".spec.gateway"},
					},
				},
			},
		},
	}
}

func globalInClusterIPPoolCRD() *apiextensionsv1.CustomResourceDefinition {
	return &apiextensionsv1.CustomResourceDefinition{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apiextensions.k8s.io/v1",
			Kind:       "CustomResourceDefinition",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "globalinclusterippools.ipam.cluster.x-k8s.io",
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
				"cluster.x-k8s.io/v1beta1":  "v1beta1",
			},
		},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "ipam.cluster.x-k8s.io",
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "GlobalInClusterIPPool",
				ListKind: "GlobalInClusterIPPoolList",
				Plural:   "globalinclusterippools",
				Singular: "globalinclusterippool",
			},
			Scope: apiextensionsv1.ClusterScoped,
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name: "v1alpha2",
					Schema: &apiextensionsv1.CustomResourceValidation{
						OpenAPIV3Schema: &apiextensionsv1.JSONSchemaProps{
							Description: "GlobalInClusterIPPool is the Schema for the globalinclusterippools API.",
							Type:        "object",
							Properties: map[string]apiextensionsv1.JSONSchemaProps{
								"apiVersion": {Type: "string"},
								"kind":       {Type: "string"},
								"metadata":   {Type: "object"},
								"spec": {
									Description: "InClusterIPPoolSpec defines the desired state of InClusterIPPool.",
									Type:        "object",
									Required:    []string{"addresses", "prefix", "gateway"},
									Properties: map[string]apiextensionsv1.JSONSchemaProps{
										"addresses": {
											Type:  "array",
											Items: &apiextensionsv1.JSONSchemaPropsOrArray{Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"}},
										},
										"prefix":  {Type: "integer"},
										"gateway": {Type: "string"},
									},
								},
								"status": {Type: "object"},
							},
						},
					},
					Served:  true,
					Storage: true,
					Subresources: &apiextensionsv1.CustomResourceSubresources{
						Status: &apiextensionsv1.CustomResourceSubresourceStatus{},
					},
				},
			},
		},
	}
}

func ipamServiceAccount() *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ServiceAccount",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      IPAMServiceAccountName,
			Namespace: IPAMProviderNamespace,
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
			},
		},
	}
}

func ipamClusterRole() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRole",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: IPAMClusterRoleName,
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
			},
		},
		Rules: []rbacv1.PolicyRule{
			{
				APIGroups: []string{""},
				Resources: []string{"events"},
				Verbs:     []string{"create", "patch"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"globalinclusterippools"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"globalinclusterippools/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"globalinclusterippools/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"inclusterippools"},
				Verbs:     []string{"get", "list", "watch"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"inclusterippools/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"inclusterippools/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddressclaims"},
				Verbs:     []string{"get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddressclaims/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddressclaims/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddresses"},
				Verbs:     []string{"create", "delete", "get", "list", "patch", "update", "watch"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddresses/finalizers"},
				Verbs:     []string{"update"},
			},
			{
				APIGroups: []string{"ipam.cluster.x-k8s.io"},
				Resources: []string{"ipaddresses/status"},
				Verbs:     []string{"get", "patch", "update"},
			},
		},
	}
}

func ipamClusterRoleBinding() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "ClusterRoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "caip-in-cluster-manager-rolebinding",
			Labels: map[string]string{
				"cluster.x-k8s.io/provider": "ipam-in-cluster",
			},
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "ClusterRole",
			Name:     IPAMClusterRoleName,
		},
		Subjects: []rbacv1.Subject{
			{
				Kind:      "ServiceAccount",
				Name:      IPAMServiceAccountName,
				Namespace: IPAMProviderNamespace,
			},
		},
	}
}

func ipamDeployment() *appsv1.Deployment {
	labels := map[string]string{
		"cluster.x-k8s.io/provider": "ipam-in-cluster",
		"control-plane":             "controller-manager",
	}

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      IPAMDeploymentName,
			Namespace: IPAMProviderNamespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					ServiceAccountName:            IPAMServiceAccountName,
					TerminationGracePeriodSeconds: ptr.To(int64(10)),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: ptr.To(true),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "manager",
							Image: IPAMProviderImage,
							Args: []string{
								"--leader-elect",
								"--diagnostics-address=:8443",
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "metrics",
									ContainerPort: 8443,
									Protocol:      corev1.ProtocolTCP,
								},
								{
									Name:          "healthz",
									ContainerPort: 9440,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/healthz",
										Port: intstr.FromString("healthz"),
									},
								},
								InitialDelaySeconds: 15,
								PeriodSeconds:       20,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/readyz",
										Port: intstr.FromString("healthz"),
									},
								},
								InitialDelaySeconds: 5,
								PeriodSeconds:       10,
							},
							Resources: corev1.ResourceRequirements{
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("500m"),
									corev1.ResourceMemory: resource.MustParse("128Mi"),
								},
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    resource.MustParse("10m"),
									corev1.ResourceMemory: resource.MustParse("64Mi"),
								},
							},
							SecurityContext: &corev1.SecurityContext{
								AllowPrivilegeEscalation: ptr.To(false),
								Privileged:               ptr.To(false),
								RunAsUser:                ptr.To(int64(65532)),
								RunAsGroup:               ptr.To(int64(65532)),
								Capabilities: &corev1.Capabilities{
									Drop: []corev1.Capability{"ALL"},
								},
							},
						},
					},
					Tolerations: []corev1.Toleration{
						{
							Key:      "node-role.kubernetes.io/master",
							Effect:   corev1.TaintEffectNoSchedule,
							Operator: corev1.TolerationOpExists,
						},
						{
							Key:      "node-role.kubernetes.io/control-plane",
							Effect:   corev1.TaintEffectNoSchedule,
							Operator: corev1.TolerationOpExists,
						},
					},
				},
			},
		},
	}
}
