package framework

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	helmReleaseNameAnnotation      = "meta.helm.sh/release-name"
	helmReleaseNamespaceAnnotation = "meta.helm.sh/release-namespace"
	packageSourceRegistryValue     = "sourceRegistry"
)

// helmReleasePodLabels are the labels charts use to tie pods to their Helm release.
var helmReleasePodLabels = []string{"app.kubernetes.io/instance", "release"}

// knownNonRegistryPackageImages lists image repositories that may come from a registry other
// than the one the package controller is configured with.
//
// MetalLB 0.16.1 deploys frr-k8s/FRR containers whose images are pulled from quay.io instead of
// the curated packages registry. A decision on how to ship them is pending. Until then they are
// allowed here so the gap stays visible in test logs. Remove these entries once the MetalLB
// package pulls them from the curated packages registry. Do not add other images here.
var knownNonRegistryPackageImages = []string{
	"quay.io/metallb/frr-k8s",
	"quay.io/frrouting/frr",
}

// packageImage is a container image found in a pod or workload template of a package release.
type packageImage struct {
	// Source is the object the image was found in, e.g. "pod/test-harbor-core-abc".
	Source    string
	Container string
	Image     string
}

func (i packageImage) String() string {
	return fmt.Sprintf("%s container %s: %s", i.Source, i.Container, i.Image)
}

// helmWorkload holds the fields shared by Deployments, StatefulSets, DaemonSets, Jobs and CronJobs
// that are needed to find a release's pods and pod templates.
type helmWorkload struct {
	Kind     string            `json:"kind"`
	Metadata metav1.ObjectMeta `json:"metadata"`
	Spec     struct {
		Selector    *metav1.LabelSelector  `json:"selector,omitempty"`
		Template    corev1.PodTemplateSpec `json:"template"`
		JobTemplate struct {
			Spec struct {
				Template corev1.PodTemplateSpec `json:"template"`
			} `json:"spec"`
		} `json:"jobTemplate"`
	} `json:"spec"`
}

type helmWorkloadList struct {
	Items []helmWorkload `json:"items"`
}

// VerifyPackageImagesFromConfiguredRegistry fails the test if any container or init container image
// of the package's Helm release does not come from the registry the package controller installs the
// package images from (the package's sourceRegistry value if set, otherwise the
// PackageBundleController defaultImageRegistry, which is the mirror when a registry mirror is configured).
//
// packageName is the Package object name, which the controller also uses as the Helm release name.
// The Package and PackageBundleController are read from mgmtCluster. Pods and workloads are read
// from the cluster under test and selected by the release's labels and Helm annotations, because a
// namespace can hold several packages.
func (e *ClusterE2ETest) VerifyPackageImagesFromConfiguredRegistry(packageName, targetNamespace string, mgmtCluster *types.Cluster) {
	ctx := context.Background()

	registry, err := e.packageImageRegistry(ctx, packageName, mgmtCluster)
	if err != nil {
		e.T.Fatalf("Getting image registry for package %s: %v", packageName, err)
	}

	var images []packageImage
	err = retrier.New(5*time.Minute, retrier.WithRetryPolicy(retrier.BackOffPolicy(10*time.Second))).Retry(func() error {
		var err error
		images, err = e.packageReleaseImages(ctx, packageName, targetNamespace)
		if err != nil {
			return err
		}
		if len(images) == 0 {
			return fmt.Errorf("no pods or workloads found for Helm release %s in namespace %s", packageName, targetNamespace)
		}
		return nil
	})
	if err != nil {
		e.T.Fatalf("Listing images of package %s: %v", packageName, err)
	}

	violations, allowed := imageSourceViolations(images, registry, knownNonRegistryPackageImages)
	for _, img := range allowed {
		e.T.Logf("Package %s: allowing known non-registry image (decision pending): %s", packageName, img)
	}
	if len(violations) > 0 {
		e.T.Fatalf("Package %s has images that do not come from the configured registry %s:\n%s",
			packageName, registry, strings.Join(violations, "\n"))
	}
	e.T.Logf("Package %s: all %d container images come from %s", packageName, len(images), registry)
}

func (e *ClusterE2ETest) packageImageRegistry(ctx context.Context, packageName string, mgmtCluster *types.Cluster) (string, error) {
	pbc, err := e.KubectlClient.GetPackageBundleController(ctx, mgmtCluster.KubeconfigFile, e.ClusterName)
	if err != nil {
		return "", fmt.Errorf("getting PackageBundleController %s: %v", e.ClusterName, err)
	}

	pkg := &packagesv1.Package{}
	packageNamespace := fmt.Sprintf("%s-%s", constants.EksaPackagesName, e.ClusterName)
	if err := e.KubectlClient.GetObject(ctx, "packages.packages.eks.amazonaws.com", packageName, packageNamespace, mgmtCluster.KubeconfigFile, pkg); err != nil {
		return "", fmt.Errorf("getting package %s/%s: %v", packageNamespace, packageName, err)
	}

	return expectedPackageImageRegistry(&pbc, pkg)
}

// expectedPackageImageRegistry mirrors how the package controller sets the chart's sourceRegistry value:
// the package's own sourceRegistry config wins, otherwise the PackageBundleController defaultImageRegistry.
func expectedPackageImageRegistry(pbc *packagesv1.PackageBundleController, pkg *packagesv1.Package) (string, error) {
	values, err := pkg.GetValues()
	if err != nil {
		return "", fmt.Errorf("parsing config of package %s: %v", pkg.Name, err)
	}
	if registry, ok := values[packageSourceRegistryValue].(string); ok && registry != "" {
		return strings.TrimSuffix(registry, "/"), nil
	}
	return strings.TrimSuffix(pbc.GetDefaultImageRegistry(), "/"), nil
}

func (e *ClusterE2ETest) packageReleaseImages(ctx context.Context, release, namespace string) ([]packageImage, error) {
	kubeconfig := e.KubeconfigFilePath()
	pods, err := e.KubectlClient.GetPods(ctx, executables.WithKubeconfig(kubeconfig), executables.WithNamespace(namespace))
	if err != nil {
		return nil, err
	}

	out, err := e.KubectlClient.ExecuteCommand(ctx, "get", "deployments,statefulsets,daemonsets,jobs,cronjobs",
		"-o", "json", "--namespace", namespace, "--kubeconfig", kubeconfig)
	if err != nil {
		return nil, fmt.Errorf("getting workloads in namespace %s: %v", namespace, err)
	}
	workloads := &helmWorkloadList{}
	if err := json.Unmarshal(out.Bytes(), workloads); err != nil {
		return nil, fmt.Errorf("parsing workloads in namespace %s: %v", namespace, err)
	}

	return releaseImages(release, namespace, pods, workloads.Items)
}

// releaseImages returns the container and init container images of the pods and workload pod
// templates that belong to a Helm release. Pods are selected by the release labels charts set,
// plus the selectors of workloads Helm annotated as part of the release, which covers charts that
// don't label pods with the release name.
func releaseImages(release, namespace string, pods []corev1.Pod, workloads []helmWorkload) ([]packageImage, error) {
	var images []packageImage
	var selectors []metav1.LabelSelector

	for _, w := range workloads {
		if !isHelmReleaseObject(w.Metadata, release, namespace) {
			continue
		}
		source := strings.ToLower(w.Kind) + "/" + w.Metadata.Name
		template := w.Spec.Template
		if w.Kind == "CronJob" {
			template = w.Spec.JobTemplate.Spec.Template
		}
		images = append(images, podSpecImages(source, template.Spec)...)
		if w.Spec.Selector != nil {
			selectors = append(selectors, *w.Spec.Selector)
		}
	}

	for _, pod := range pods {
		matched, err := podBelongsToRelease(pod, release, selectors)
		if err != nil {
			return nil, err
		}
		if matched {
			images = append(images, podSpecImages("pod/"+pod.Name, pod.Spec)...)
		}
	}

	return images, nil
}

func isHelmReleaseObject(meta metav1.ObjectMeta, release, namespace string) bool {
	if meta.Annotations[helmReleaseNameAnnotation] != release {
		return false
	}
	ns, ok := meta.Annotations[helmReleaseNamespaceAnnotation]
	return !ok || ns == namespace
}

func podBelongsToRelease(pod corev1.Pod, release string, selectors []metav1.LabelSelector) (bool, error) {
	for _, label := range helmReleasePodLabels {
		if pod.Labels[label] == release {
			return true, nil
		}
	}
	for i := range selectors {
		selector, err := metav1.LabelSelectorAsSelector(&selectors[i])
		if err != nil {
			return false, fmt.Errorf("parsing workload selector: %v", err)
		}
		if !selector.Empty() && selector.Matches(labels.Set(pod.Labels)) {
			return true, nil
		}
	}
	return false, nil
}

func podSpecImages(source string, spec corev1.PodSpec) []packageImage {
	images := make([]packageImage, 0, len(spec.InitContainers)+len(spec.Containers))
	for _, c := range spec.InitContainers {
		images = append(images, packageImage{Source: source, Container: c.Name, Image: c.Image})
	}
	for _, c := range spec.Containers {
		images = append(images, packageImage{Source: source, Container: c.Name, Image: c.Image})
	}
	return images
}

// imageSourceViolations returns the images that don't come from registry, and the images that don't
// but are in the allowed list. Both lists are sorted and de-duplicated.
func imageSourceViolations(images []packageImage, registry string, allowedRepositories []string) (violations, allowed []string) {
	allowedSet := map[string]bool{}
	for _, r := range allowedRepositories {
		allowedSet[r] = true
	}
	seenViolations := map[string]bool{}
	seenAllowed := map[string]bool{}
	for _, img := range images {
		if imageFromRegistry(img.Image, registry) {
			continue
		}
		if allowedSet[imageRepository(img.Image)] {
			if !seenAllowed[img.Image] {
				seenAllowed[img.Image] = true
				allowed = append(allowed, img.Image)
			}
			continue
		}
		if s := img.String(); !seenViolations[s] {
			seenViolations[s] = true
			violations = append(violations, s)
		}
	}
	sort.Strings(violations)
	sort.Strings(allowed)
	return violations, allowed
}

func imageFromRegistry(image, registry string) bool {
	registry = strings.TrimSuffix(registry, "/")
	return registry != "" && strings.HasPrefix(image, registry+"/")
}

// imageRepository strips the tag and digest from an image reference.
func imageRepository(image string) string {
	if i := strings.Index(image, "@"); i >= 0 {
		image = image[:i]
	}
	if colon := strings.LastIndex(image, ":"); colon > strings.LastIndex(image, "/") {
		image = image[:colon]
	}
	return image
}
