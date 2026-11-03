package framework

import (
	"testing"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
