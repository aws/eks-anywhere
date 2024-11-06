package kubernetes

import (
	"fmt"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	tinkerbellv1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/api/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/exp/addons/api/v1beta1"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	tinkv1 "github.com/aws/eks-anywhere/internal/thirdparty/tink/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type schemeAdder func(s *runtime.Scheme) error

var schemeAdders = []schemeAdder{
	// clientgoscheme adds all the native K8s kinds
	clientgoscheme.AddToScheme,
	clusterv1.AddToScheme,
	controlplanev1.AddToScheme,
	anywherev1.AddToScheme,
	snowv1.AddToScheme,
	cloudstackv1.AddToScheme,
	bootstrapv1.AddToScheme,
	dockerv1.AddToScheme,
	releasev1.AddToScheme,
	eksdv1alpha1.AddToScheme,
	vspherev1.AddToScheme,
	etcdv1.AddToScheme,
	addonsv1.AddToScheme,
	tinkerbellv1.AddToScheme,
	tinkv1.AddToScheme,
}

func addToScheme(scheme *runtime.Scheme, schemeAdders ...schemeAdder) error {
	for _, adder := range schemeAdders {
		if err := adder(scheme); err != nil {
			return err
		}
	}

	return nil
}

// Scheme is a wrapper around runtime.Scheme that provides additional functionality.
type Scheme struct {
	*runtime.Scheme
}

// KubectlResourceTypeForObj returns the resource type for the provided object.
// If the object is not registered in the scheme, an error is returned.
func (s *Scheme) KubectlResourceTypeForObj(obj runtime.Object) (string, error) {
	groupVersionKind, err := apiutil.GVKForObject(obj, s.Scheme)
	if err != nil {
		return "", err
	}

	if meta.IsListType(obj) && strings.HasSuffix(groupVersionKind.Kind, "List") {
		// if obj is a list, treat it as a request for the "individual" item's resource
		groupVersionKind.Kind = TrimListFromKind(groupVersionKind.Kind)
	}

	return groupVersionToKubectlResourceType(groupVersionKind), nil
}

func TrimListFromKind(kind string) string {
	if strings.HasSuffix(kind, "List") {
		return kind[:len(kind)-4]
	}
	return kind
}

func groupVersionToKubectlResourceType(g schema.GroupVersionKind) string {
	if g.Group == "" {
		// if Group is not set, this probably an obj from "core", which api group is just v1
		return g.Kind
	}

	return fmt.Sprintf("%s.%s.%s", g.Kind, g.Version, g.Group)
}

// InitScheme adds the common EKS-A types to the provided scheme.
// It is not thread safe.
func InitScheme(s *runtime.Scheme) error {
	return addToScheme(s, schemeAdders...)
}
