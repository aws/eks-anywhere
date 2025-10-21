package kubernetes

import (
	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	cloudstackv1 "sigs.k8s.io/cluster-api-provider-cloudstack/api/v1beta3"
	vspherev1 "sigs.k8s.io/cluster-api-provider-vsphere/apis/v1beta1"
	addonsv1 "sigs.k8s.io/cluster-api/api/addons/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta1"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta1"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1 "sigs.k8s.io/cluster-api/api/core/v1beta1"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"
	dockerv1 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta1"
	dockerv1beta2 "sigs.k8s.io/cluster-api/test/infrastructure/docker/api/v1beta2"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	tinkerbellv1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/capt/v1beta1"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
	releasev1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type schemeAdder func(s *runtime.Scheme) error

var schemeAdders = []schemeAdder{
	// clientgoscheme adds all the native K8s kinds
	clientgoscheme.AddToScheme,
	clusterv1.AddToScheme,
	clusterv1beta2.AddToScheme,
	controlplanev1.AddToScheme,
	controlplanev1beta2.AddToScheme,
	bootstrapv1.AddToScheme,
	bootstrapv1beta2.AddToScheme,
	anywherev1.AddToScheme,
	snowv1.AddToScheme,
	cloudstackv1.AddToScheme,
	dockerv1.AddToScheme,
	dockerv1beta2.AddToScheme,
	releasev1.AddToScheme,
	eksdv1alpha1.AddToScheme,
	vspherev1.AddToScheme,
	etcdv1.AddToScheme,
	addonsv1.AddToScheme,
	tinkerbellv1.AddToScheme,
}

func addToScheme(scheme *runtime.Scheme, schemeAdders ...schemeAdder) error {
	for _, adder := range schemeAdders {
		if err := adder(scheme); err != nil {
			return err
		}
	}

	return nil
}
