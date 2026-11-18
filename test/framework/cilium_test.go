package framework

import (
	"context"
	"errors"
	"strings"
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/yaml"
)

func TestOSSCiliumVersion(t *testing.T) {
	tests := []struct {
		name     string
		image    string
		expected string
	}{
		{
			name:     "current release",
			image:    "public.ecr.aws/eks/cilium/cilium:v1.19.10-3",
			expected: "1.19.10",
		},
		{
			name:     "released CLI",
			image:    "public.ecr.aws/eks/cilium/cilium:v1.18.5-2",
			expected: "1.18.5",
		},
		{
			name:     "development build",
			image:    "public.ecr.aws/example/cilium/cilium:v1.19.10-3-eks-a-v0.27.0-dev-build.1",
			expected: "1.19.10",
		},
		{
			name:     "tag and digest",
			image:    "public.ecr.aws/eks/cilium/cilium:v1.19.10-3@sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			expected: "1.19.10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ds := &appsv1.DaemonSet{Spec: appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
				Containers: []corev1.Container{{Name: "cilium-agent", Image: tt.image}},
			}}}}
			version, err := ossCiliumVersion(ds)
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(version).To(Equal(tt.expected))
		})
	}
}

func TestOSSCiliumVersionMissing(t *testing.T) {
	g := NewWithT(t)
	_, err := ossCiliumVersion(nil)
	g.Expect(err).To(MatchError("Cilium DaemonSet not found"))
	_, err = ossCiliumVersion(&appsv1.DaemonSet{})
	g.Expect(err).To(MatchError("Cilium agent container not found"))
}

func TestOSSCiliumVersionInvalid(t *testing.T) {
	g := NewWithT(t)
	ds := &appsv1.DaemonSet{Spec: appsv1.DaemonSetSpec{Template: corev1.PodTemplateSpec{Spec: corev1.PodSpec{
		Containers: []corev1.Container{{Name: "cilium-agent", Image: "cilium:latest"}},
	}}}}
	_, err := ossCiliumVersion(ds)
	g.Expect(err).To(MatchError(ContainSubstring("parsing Cilium version")))
}

func TestDeleteCiliumRBACAfterOlderCLIUninstall(t *testing.T) {
	g := NewWithT(t)
	// Rendered from upstream cilium:1.19.10 with the replacement fixture's Helm values.
	roleYAML := `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cilium-operator-tlsinterception-secrets
  namespace: "cilium-secrets"
  labels:
    app.kubernetes.io/part-of: cilium
rules:
- apiGroups:
  - ""
  resources:
  - secrets
  verbs:
  - create
  - delete
  - update
  - patch
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: cilium-operator-ztunnel
  namespace: kube-system
  labels:
    app.kubernetes.io/part-of: cilium
rules:
# ZTunnel DaemonSet management permissions
# Note: These permissions must always be granted (not conditional on encryption.type)
# because the controller needs to clean up stale DaemonSets when ztunnel is disabled.
- apiGroups:
  - apps
  resources:
  - daemonsets
  verbs:
  - create
  - delete
  - get
  - list
  - watch
`
	bindingYAML := `
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cilium-operator-tlsinterception-secrets
  namespace: "cilium-secrets"
  labels:
    app.kubernetes.io/part-of: cilium
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cilium-operator-tlsinterception-secrets
subjects:
- kind: ServiceAccount
  name: "cilium-operator"
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: cilium-operator-ztunnel
  namespace: kube-system
  labels:
    app.kubernetes.io/part-of: cilium
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: cilium-operator-ztunnel
subjects:
- kind: ServiceAccount
  name: "cilium-operator"
  namespace: kube-system
`
	var chartRBAC []client.Object
	for _, document := range strings.Split(roleYAML, "---") {
		if strings.TrimSpace(document) == "" {
			continue
		}
		role := &rbacv1.Role{}
		g.Expect(yaml.UnmarshalStrict([]byte(document), role)).To(Succeed())
		g.Expect(role.Rules).NotTo(BeEmpty())
		chartRBAC = append(chartRBAC, role)
	}
	for _, document := range strings.Split(bindingYAML, "---") {
		if strings.TrimSpace(document) == "" {
			continue
		}
		binding := &rbacv1.RoleBinding{}
		g.Expect(yaml.UnmarshalStrict([]byte(document), binding)).To(Succeed())
		chartRBAC = append(chartRBAC, binding)
	}
	g.Expect(chartRBAC).To(HaveLen(4))
	unrelated := chartRBAC[0].(*rbacv1.Role).DeepCopy()
	unrelated.Name = "application-role"
	unrelated.Labels = map[string]string{"app.kubernetes.io/part-of": "application"}
	unrelatedBinding := chartRBAC[2].(*rbacv1.RoleBinding).DeepCopy()
	unrelatedBinding.Name = "application-binding"
	unrelatedBinding.Labels = nil
	scheme := runtime.NewScheme()
	g.Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
	objects := append([]client.Object{unrelated, unrelatedBinding}, chartRBAC...)
	c := fake.NewClientBuilder().WithScheme(scheme).WithObjects(objects...).Build()
	ctx := context.Background()
	for _, object := range chartRBAC {
		replacement := object.DeepCopyObject().(client.Object)
		replacement.SetResourceVersion("")
		g.Expect(apierrors.IsAlreadyExists(c.Create(ctx, replacement))).To(BeTrue())
	}
	for range 2 {
		g.Expect(deleteCiliumRBAC(ctx, c)).To(Succeed())
	}
	for _, object := range chartRBAC {
		g.Expect(apierrors.IsNotFound(c.Get(ctx, client.ObjectKeyFromObject(object), object))).To(BeTrue())
		replacement := object.DeepCopyObject().(client.Object)
		replacement.SetResourceVersion("")
		g.Expect(c.Create(ctx, replacement)).To(Succeed())
	}
	for _, object := range []client.Object{unrelated, unrelatedBinding} {
		g.Expect(c.Get(ctx, client.ObjectKeyFromObject(object), object)).To(Succeed())
	}
}

func TestDeleteCiliumRBACPreservesAPIFailure(t *testing.T) {
	want := errors.New("RBAC delete denied")
	c := fake.NewClientBuilder().WithInterceptorFuncs(interceptor.Funcs{
		List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
			return want
		},
	}).Build()
	if err := deleteCiliumRBAC(context.Background(), c); !errors.Is(err, want) {
		t.Fatalf("err=%v", err)
	}
}

func TestDeleteCiliumLeftoversAfterOlderCLIUninstall(t *testing.T) {
	g := NewWithT(t)
	// Metadata from cilium:1.19.10 rendered with the replacement values.
	peer := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hubble-peer", Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app":                   "cilium",
				"app.kubernetes.io/part-of": "cilium",
				"app.kubernetes.io/name":    "hubble-peer",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"k8s-app": "cilium"},
			Ports:    []corev1.ServicePort{{Name: "peer-service", Port: 443}},
		},
	}
	unrelated := peer.DeepCopy()
	unrelated.Name = "application-service"
	otherNamespace := peer.DeepCopy()
	otherNamespace.Namespace = "application"
	role := &rbacv1.Role{ObjectMeta: metav1.ObjectMeta{
		Name: "cilium-operator-ztunnel", Namespace: "kube-system",
		Labels: map[string]string{"app.kubernetes.io/part-of": "cilium"},
	}}
	scheme := runtime.NewScheme()
	g.Expect(corev1.AddToScheme(scheme)).To(Succeed())
	g.Expect(rbacv1.AddToScheme(scheme)).To(Succeed())
	c := fake.NewClientBuilder().WithScheme(scheme).
		WithObjects(peer, unrelated, otherNamespace, role).Build()
	ctx := context.Background()
	replacement := peer.DeepCopy()
	replacement.SetResourceVersion("")
	g.Expect(apierrors.IsAlreadyExists(c.Create(ctx, replacement))).To(BeTrue())
	for range 2 {
		g.Expect(deleteCiliumLeftovers(ctx, c)).To(Succeed())
	}
	g.Expect(apierrors.IsNotFound(c.Get(ctx, client.ObjectKeyFromObject(peer), peer))).To(BeTrue())
	g.Expect(apierrors.IsNotFound(c.Get(ctx, client.ObjectKeyFromObject(role), role))).To(BeTrue())
	g.Expect(c.Create(ctx, replacement)).To(Succeed())
	for _, service := range []*corev1.Service{unrelated, otherNamespace} {
		g.Expect(c.Get(ctx, client.ObjectKeyFromObject(service), service)).To(Succeed())
	}
}

func TestDeleteCiliumLeftoversPreservesUnrelatedPeer(t *testing.T) {
	for _, label := range []string{"k8s-app", "app.kubernetes.io/part-of", "app.kubernetes.io/name"} {
		t.Run(label, func(t *testing.T) {
			g := NewWithT(t)
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name: "hubble-peer", Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":                   "cilium",
					"app.kubernetes.io/part-of": "cilium",
					"app.kubernetes.io/name":    "hubble-peer",
				},
			}}
			delete(service.Labels, label)
			c := fake.NewClientBuilder().WithObjects(service).Build()
			g.Expect(deleteCiliumLeftovers(context.Background(), c)).To(
				MatchError("refusing to delete hubble-peer service without Cilium labels"))
			g.Expect(c.Get(context.Background(), client.ObjectKeyFromObject(service), service)).To(Succeed())
		})
	}
}

func TestDeleteCiliumLeftoversPreservesServiceAPIFailure(t *testing.T) {
	want := errors.New("peer service access denied")
	for _, operation := range []string{"get", "delete"} {
		t.Run(operation, func(t *testing.T) {
			g := NewWithT(t)
			service := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
				Name: "hubble-peer", Namespace: "kube-system",
				Labels: map[string]string{
					"k8s-app":                   "cilium",
					"app.kubernetes.io/part-of": "cilium",
					"app.kubernetes.io/name":    "hubble-peer",
				},
			}}
			c := fake.NewClientBuilder().WithObjects(service).WithInterceptorFuncs(interceptor.Funcs{
				Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, object client.Object, opts ...client.GetOption) error {
					if operation == "get" {
						return want
					}
					return c.Get(ctx, key, object, opts...)
				},
				Delete: func(context.Context, client.WithWatch, client.Object, ...client.DeleteOption) error {
					return want
				},
			}).Build()
			g.Expect(errors.Is(deleteCiliumLeftovers(context.Background(), c), want)).To(BeTrue())
		})
	}
}
