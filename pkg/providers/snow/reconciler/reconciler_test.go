package reconciler_test

import (
	"context"
	"testing"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log"

	_ "github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler/mocks"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var (
	clusterName      = "snow-test"
	clusterNamespace = "test-namespace"
)

func TestSnowClusterReconcileSuccess(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	cniReconciler := mocks.NewMockCNIReconciler(ctrl)

	managementCluster := createSnowCluster()
	managementCluster.Name = "management-cluster"
	cluster := createSnowCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

	datacenterConfig := createSnowDataCenter(cluster)

	bundle := createBundle(cluster)
	cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		APIVersion: bundle.APIVersion,
		Name:       bundle.Name,
		Namespace:  bundle.Namespace,
	}

	eksd := createEksdRelease()
	machineConfigCP := createSnowCPMachineConfig(true)
	machineConfigWN := createSnowWNMachineConfig(true)

	client := fake.NewClientBuilder().WithObjects(cluster, managementCluster, datacenterConfig, bundle, machineConfigCP, machineConfigWN, eksd).Build()

	_, e := reconciler.New(client, cniReconciler).Reconcile(context.Background(), log.Log, cluster)

	g.Expect(e).To(BeNil())
	g.Expect(cluster.Status.FailureMessage).To(BeZero())
}

func TestSnowClusterReconcileFailDueToInvalidWNMachineConfig(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	cniReconciler := mocks.NewMockCNIReconciler(ctrl)

	managementCluster := createSnowCluster()
	managementCluster.Name = "management-cluster"
	cluster := createSnowCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

	datacenterConfig := createSnowDataCenter(cluster)

	bundle := createBundle(cluster)
	cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		APIVersion: bundle.APIVersion,
		Name:       bundle.Name,
		Namespace:  bundle.Namespace,
	}

	eksd := createEksdRelease()
	machineConfigCP := createSnowCPMachineConfig(true)
	machineConfigWN := createSnowWNMachineConfig(false)
	mcFailureMessage := "Something wrong"
	machineConfigWN.Status.FailureMessage = &mcFailureMessage

	client := fake.NewClientBuilder().WithObjects(cluster, managementCluster, datacenterConfig, bundle, machineConfigCP, machineConfigWN, eksd).Build()

	_, e := reconciler.New(client, cniReconciler).Reconcile(context.Background(), log.Log, cluster)
	g.Expect(e).To(BeNil(), "error should be nil to prevent requeue")

	g.Expect(cluster.Status.FailureMessage).ToNot(BeZero())
	g.Expect(*cluster.Status.FailureMessage).To(ContainSubstring("SnowMachineConfig snow-test-wn is invalid"))
	g.Expect(*cluster.Status.FailureMessage).To(ContainSubstring("Something wrong"))
}

func TestSnowClusterReconcileFailDueToInvalidCPMachineConfig(t *testing.T) {
	g := NewWithT(t)
	ctrl := gomock.NewController(t)
	cniReconciler := mocks.NewMockCNIReconciler(ctrl)

	managementCluster := createSnowCluster()
	managementCluster.Name = "management-cluster"
	cluster := createSnowCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

	datacenterConfig := createSnowDataCenter(cluster)

	bundle := createBundle(cluster)
	cluster.Spec.BundlesRef = &anywherev1.BundlesRef{
		APIVersion: bundle.APIVersion,
		Name:       bundle.Name,
		Namespace:  bundle.Namespace,
	}

	eksd := createEksdRelease()
	machineConfigCP := createSnowCPMachineConfig(false)
	machineConfigWN := createSnowWNMachineConfig(true)

	client := fake.NewClientBuilder().WithObjects(cluster, managementCluster, datacenterConfig, bundle, machineConfigCP, machineConfigWN, eksd).Build()

	_, e := reconciler.New(client, cniReconciler).Reconcile(context.Background(), log.Log, cluster)
	g.Expect(e).To(BeNil(), "error should be nil to prevent requeue")

	g.Expect(cluster.Status.FailureMessage).ToNot(BeZero())
	g.Expect(*cluster.Status.FailureMessage).To(ContainSubstring("SnowMachineConfig snow-test-cp is invalid"))
}

func createSnowCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName,
			Namespace: clusterNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			DatacenterRef: anywherev1.Ref{
				Kind: "SnowDatacenterConfig",
				Name: "datacenter",
			},
			KubernetesVersion: "1.20",
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: "SnowMachineConfig",
					Name: clusterName + "-cp",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				{
					Count: 1,
					MachineGroupRef: &anywherev1.Ref{
						Kind: "SnowMachineConfig",
						Name: clusterName + "-wn",
					},
					Name:   "md-0",
					Labels: nil,
				},
			},
		},
	}
}

func createSnowDataCenter(cluster *anywherev1.Cluster) *anywherev1.SnowDatacenterConfig {
	return &anywherev1.SnowDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SnowDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: cluster.Namespace,
		},
	}
}

func createBundle(cluster *anywherev1.Cluster) *v1alpha1.Bundles {
	return &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: "default",
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				{
					KubeVersion: "1.20",
					EksD: v1alpha1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
						KubeVersion:    "1.20",
					},
					CertManager:            v1alpha1.CertManagerBundle{},
					ClusterAPI:             v1alpha1.CoreClusterAPI{},
					Bootstrap:              v1alpha1.KubeadmBootstrapBundle{},
					ControlPlane:           v1alpha1.KubeadmControlPlaneBundle{},
					Aws:                    v1alpha1.AwsBundle{},
					VSphere:                v1alpha1.VSphereBundle{},
					Docker:                 v1alpha1.DockerBundle{},
					Eksa:                   v1alpha1.EksaBundle{},
					Cilium:                 v1alpha1.CiliumBundle{},
					Kindnetd:               v1alpha1.KindnetdBundle{},
					Flux:                   v1alpha1.FluxBundle{},
					BottleRocketBootstrap:  v1alpha1.BottlerocketBootstrapBundle{},
					BottleRocketAdmin:      v1alpha1.BottlerocketAdminBundle{},
					ExternalEtcdBootstrap:  v1alpha1.EtcdadmBootstrapBundle{},
					ExternalEtcdController: v1alpha1.EtcdadmControllerBundle{},
					Tinkerbell:             v1alpha1.TinkerbellBundle{},
				},
			},
		},
	}
}

func createEksdRelease() *eksdv1alpha1.Release {
	return &eksdv1alpha1.Release{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test",
			Namespace: "eksa-system",
		},
		Status: eksdv1alpha1.ReleaseStatus{
			Components: []eksdv1alpha1.Component{
				{
					Assets: []eksdv1alpha1.Asset{
						{
							Name:  "etcd-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "node-driver-registrar-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "livenessprobe-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "external-attacher-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "external-provisioner-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "pause-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "aws-iam-authenticator-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "coredns-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
						{
							Name:  "kube-apiserver-image",
							Image: &eksdv1alpha1.AssetImage{},
						},
					},
				},
			},
		},
	}
}

func createSnowCPMachineConfig(isValid bool) *anywherev1.SnowMachineConfig {
	return &anywherev1.SnowMachineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-cp",
			Namespace: clusterNamespace,
		},
		Status: anywherev1.SnowMachineConfigStatus{
			SpecValid: isValid,
		},
	}
}

func createSnowWNMachineConfig(isValid bool) *anywherev1.SnowMachineConfig {
	return &anywherev1.SnowMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "SnowMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterName + "-wn",
			Namespace: clusterNamespace,
		},
		Status: anywherev1.SnowMachineConfigStatus{
			SpecValid: isValid,
		},
	}
}
