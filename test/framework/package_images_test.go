package framework

import (
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
)

func TestImageRepository(t *testing.T) {
	tests := map[string]string{
		"quay.io/frrouting/frr:9.1.0":                         "quay.io/frrouting/frr",
		"quay.io/metallb/frr-k8s@sha256:abc":                  "quay.io/metallb/frr-k8s",
		"1.2.3.4:5000/curated/harbor/harbor-core:v1@sha256:a": "1.2.3.4:5000/curated/harbor/harbor-core",
		"1.2.3.4:5000/curated/harbor/harbor-core":             "1.2.3.4:5000/curated/harbor/harbor-core",
		"nginx": "nginx",
	}
	for image, want := range tests {
		t.Run(image, func(t *testing.T) {
			NewWithT(t).Expect(imageRepository(image)).To(Equal(want))
		})
	}
}

func TestImageFromRegistry(t *testing.T) {
	tests := []struct {
		image, registry string
		want            bool
	}{
		{"067575901363.dkr.ecr.us-west-2.amazonaws.com/harbor/harbor-core@sha256:a", "067575901363.dkr.ecr.us-west-2.amazonaws.com", true},
		{"067575901363.dkr.ecr.us-west-2.amazonaws.com/harbor/harbor-core@sha256:a", "067575901363.dkr.ecr.us-west-2.amazonaws.com/", true},
		{"067575901363.dkr.ecr.us-west-2.amazonaws.com.example.com/harbor/harbor-core", "067575901363.dkr.ecr.us-west-2.amazonaws.com", false},
		{"1.2.3.4:443/curated-packages/harbor/harbor-core", "1.2.3.4:443/curated-packages", true},
		{"1.2.3.4:443/eks-anywhere/harbor/harbor-core", "1.2.3.4:443/curated-packages", false},
		{"quay.io/frrouting/frr:9.1.0", "067575901363.dkr.ecr.us-west-2.amazonaws.com", false},
		{"nginx", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.image+" from "+tt.registry, func(t *testing.T) {
			NewWithT(t).Expect(imageFromRegistry(tt.image, tt.registry)).To(Equal(tt.want))
		})
	}
}

func TestExpectedPackageImageRegistry(t *testing.T) {
	g := NewWithT(t)
	pbc := &packagesv1.PackageBundleController{Spec: packagesv1.PackageBundleControllerSpec{DefaultImageRegistry: "1.2.3.4:443/curated-packages/"}}

	got, err := expectedPackageImageRegistry(pbc, &packagesv1.Package{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(got).To(Equal("1.2.3.4:443/curated-packages"))

	got, err = expectedPackageImageRegistry(pbc, &packagesv1.Package{Spec: packagesv1.PackageSpec{Config: "sourceRegistry: public.ecr.aws/x3k6m8v0\n"}})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(got).To(Equal("public.ecr.aws/x3k6m8v0"))

	got, err = expectedPackageImageRegistry(&packagesv1.PackageBundleController{}, &packagesv1.Package{})
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(got).To(Equal("783794618700.dkr.ecr.us-west-2.amazonaws.com"))

	_, err = expectedPackageImageRegistry(pbc, &packagesv1.Package{Spec: packagesv1.PackageSpec{Config: "not: [valid"}})
	g.Expect(err).To(HaveOccurred())
}

func testPod(name string, labels map[string]string, images ...string) corev1.Pod {
	pod := corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: name, Labels: labels}}
	pod.Spec.InitContainers = []corev1.Container{{Name: "init", Image: images[0]}}
	for _, image := range images[1:] {
		pod.Spec.Containers = append(pod.Spec.Containers, corev1.Container{Name: "main", Image: image})
	}
	return pod
}

func testWorkload(kind, name, release, namespace string, selector map[string]string, image string) helmWorkload {
	w := helmWorkload{Kind: kind, Metadata: metav1.ObjectMeta{Name: name, Annotations: map[string]string{
		helmReleaseNameAnnotation:      release,
		helmReleaseNamespaceAnnotation: namespace,
	}}}
	if selector != nil {
		w.Spec.Selector = &metav1.LabelSelector{MatchLabels: selector}
	}
	spec := corev1.PodSpec{Containers: []corev1.Container{{Name: "main", Image: image}}}
	if kind == "CronJob" {
		w.Spec.JobTemplate.Spec.Template.Spec = spec
	} else {
		w.Spec.Template.Spec = spec
	}
	return w
}

func TestReleaseImages(t *testing.T) {
	g := NewWithT(t)
	pods := []corev1.Pod{
		testPod("instance-label", map[string]string{"app.kubernetes.io/instance": "test"}, "reg/init:1", "reg/a:1"),
		testPod("release-label", map[string]string{"release": "test"}, "reg/init:1", "reg/b:1"),
		// Charts like hello-eks-anywhere only set app labels; the Helm-managed Deployment selects the pod.
		testPod("selected-by-workload", map[string]string{"app": "hello"}, "reg/init:1", "reg/c:1"),
		testPod("other-release", map[string]string{"app.kubernetes.io/instance": "other", "app": "other"}, "quay.io/x/init:1", "quay.io/x/y:1"),
		testPod("unlabeled", nil, "busybox", "busybox"),
	}
	workloads := []helmWorkload{
		testWorkload("Deployment", "hello", "test", "ns", map[string]string{"app": "hello"}, "reg/c:1"),
		testWorkload("CronJob", "cleanup", "test", "ns", nil, "reg/d:1"),
		testWorkload("Deployment", "other", "other", "ns", map[string]string{"app": "other"}, "quay.io/x/y:1"),
		testWorkload("Deployment", "other-namespace", "test", "elsewhere", map[string]string{"app": "other"}, "quay.io/x/y:1"),
	}

	images, err := releaseImages("test", "ns", pods, workloads)
	g.Expect(err).NotTo(HaveOccurred())

	var got []string
	for _, i := range images {
		got = append(got, i.String())
	}
	g.Expect(got).To(ConsistOf(
		"deployment/hello container main: reg/c:1",
		"cronjob/cleanup container main: reg/d:1",
		"pod/instance-label container init: reg/init:1",
		"pod/instance-label container main: reg/a:1",
		"pod/release-label container init: reg/init:1",
		"pod/release-label container main: reg/b:1",
		"pod/selected-by-workload container init: reg/init:1",
		"pod/selected-by-workload container main: reg/c:1",
	))
}

func TestImageSourceViolations(t *testing.T) {
	g := NewWithT(t)
	registry := "067575901363.dkr.ecr.us-west-2.amazonaws.com"
	images := []packageImage{
		{Source: "daemonset/test-metallb-speaker", Container: "speaker", Image: registry + "/metallb/speaker@sha256:a"},
		{Source: "daemonset/test-metallb-speaker", Container: "frr", Image: "quay.io/frrouting/frr:9.1.0"},
		{Source: "pod/test-metallb-speaker-abc", Container: "frr", Image: "quay.io/frrouting/frr:9.1.0"},
		{Source: "daemonset/test-frr-k8s", Container: "controller", Image: "quay.io/metallb/frr-k8s:v0.0.17"},
		{Source: "pod/test-other", Container: "main", Image: "docker.io/library/nginx:latest"},
		{Source: "pod/test-other", Container: "main", Image: "docker.io/library/nginx:latest"},
		{Source: "deployment/test-other", Container: "main", Image: "public.ecr.aws/eks-anywhere/nginx:latest"},
	}

	violations, allowed := imageSourceViolations(images, registry, knownNonRegistryPackageImages)
	g.Expect(violations).To(Equal([]string{
		"deployment/test-other container main: public.ecr.aws/eks-anywhere/nginx:latest",
		"pod/test-other container main: docker.io/library/nginx:latest",
	}))
	g.Expect(allowed).To(Equal([]string{"quay.io/frrouting/frr:9.1.0", "quay.io/metallb/frr-k8s:v0.0.17"}))
}

func TestKnownNonRegistryPackageImagesOnlyMetalLBFRR(t *testing.T) {
	NewWithT(t).Expect(knownNonRegistryPackageImages).To(ConsistOf("quay.io/metallb/frr-k8s", "quay.io/frrouting/frr"))
}
