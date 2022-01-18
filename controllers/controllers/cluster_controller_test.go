package controllers

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere"
	"github.com/aws/eks-anywhere/pkg/providers/vsphere/mocks"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

var (
	name      = "test-cluster"
	namespace = "default"
)

func TestClusterReconcilerSkipManagement(t *testing.T) {
	ctrl := gomock.NewController(t)
	govcClient := mocks.NewMockProviderGovcClient(ctrl)

	secret := createSecret()
	cluster := createCluster()
	datacenterConfig := createDataCenter(cluster)
	bundle := createBundle(cluster)
	machineConfigCP := createCPMachineConfig()
	machineConfigWN := createWNMachineConfig()

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN}

	s := scheme.Scheme
	s.AddKnownTypes(anywherev1.GroupVersion, cluster)
	s.AddKnownTypes(anywherev1.GroupVersion, datacenterConfig)
	s.AddKnownTypes(anywherev1.GroupVersion, bundle)
	s.AddKnownTypes(anywherev1.GroupVersion, machineConfigCP)

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := vsphere.NewValidator(govcClient, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govcClient)

	r := &ClusterReconciler{
		client:    cl,
		log:       logf.Log,
		validator: validator,
		defaulter: defaulter,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	ctx := context.Background()
	govcClient.EXPECT().ValidateVCenterSetupMachineConfig(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(0)
	govcClient.EXPECT().SearchTemplate(ctx, datacenterConfig.Spec.Datacenter, gomock.Any()).Return("test", nil).Times(0)
	govcClient.EXPECT().GetTags(ctx, machineConfigCP.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", bundle.Spec.VersionsBundles[0].EksD.Name)}, nil).Times(0)
	govcClient.EXPECT().GetWorkloadAvailableSpace(ctx, machineConfigCP.Spec.Datastore).Return(100.0, nil).Times(2).Times(0)

	_, err := r.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	apiCluster := &anywherev1.Cluster{}

	err = r.client.Get(context.TODO(), req.NamespacedName, apiCluster)
	if err != nil {
		t.Fatalf("get cluster: (%v)", err)
	}
	if apiCluster.Status.FailureMessage != nil {
		t.Errorf("Expected failure message to be nil. FailureMessage:%s", *apiCluster.Status.FailureMessage)
	}
}

func TestClusterReconcilerSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	govcClient := mocks.NewMockProviderGovcClient(ctrl)

	secret := createSecret()
	cluster := createCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

	datacenterConfig := createDataCenter(cluster)
	bundle := createBundle(cluster)
	machineConfigCP := createCPMachineConfig()
	machineConfigWN := createWNMachineConfig()

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN}

	s := scheme.Scheme
	s.AddKnownTypes(anywherev1.GroupVersion, cluster)
	s.AddKnownTypes(anywherev1.GroupVersion, datacenterConfig)
	s.AddKnownTypes(anywherev1.GroupVersion, bundle)
	s.AddKnownTypes(anywherev1.GroupVersion, machineConfigCP)

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := vsphere.NewValidator(govcClient, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govcClient)

	r := &ClusterReconciler{
		client:    cl,
		log:       logf.Log,
		validator: validator,
		defaulter: defaulter,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	ctx := context.Background()
	govcClient.EXPECT().ValidateVCenterSetupMachineConfig(ctx, datacenterConfig, machineConfigCP, gomock.Any()).Return(nil)
	govcClient.EXPECT().ValidateVCenterSetupMachineConfig(ctx, datacenterConfig, machineConfigWN, gomock.Any()).Return(nil)
	govcClient.EXPECT().SearchTemplate(ctx, datacenterConfig.Spec.Datacenter, machineConfigCP).Return("test", nil)
	govcClient.EXPECT().GetTags(ctx, machineConfigCP.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", bundle.Spec.VersionsBundles[0].EksD.Name)}, nil)
	govcClient.EXPECT().GetWorkloadAvailableSpace(ctx, machineConfigCP.Spec.Datastore).Return(100.0, nil).Times(2)

	_, err := r.Reconcile(ctx, req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}

	apiCluster := &anywherev1.Cluster{}

	err = r.client.Get(context.TODO(), req.NamespacedName, apiCluster)
	if err != nil {
		t.Fatalf("get cluster: (%v)", err)
	}
	if apiCluster.Status.FailureMessage != nil {
		t.Errorf("Expected failure message to be nil. FailureMessage:%s", *apiCluster.Status.FailureMessage)
	}
}

func TestClusterReconcilerFailToSetUpMachineConfigCP(t *testing.T) {
	ctrl := gomock.NewController(t)
	govcClient := mocks.NewMockProviderGovcClient(ctrl)

	secret := createSecret()
	cluster := createCluster()
	cluster.Spec.ManagementCluster = anywherev1.ManagementCluster{Name: "management-cluster"}

	datacenterConfig := createDataCenter(cluster)
	bundle := createBundle(cluster)
	machineConfigCP := createCPMachineConfig()
	machineConfigWN := createWNMachineConfig()

	objs := []runtime.Object{cluster, datacenterConfig, secret, bundle, machineConfigCP, machineConfigWN}

	s := scheme.Scheme
	s.AddKnownTypes(anywherev1.GroupVersion, cluster)
	s.AddKnownTypes(anywherev1.GroupVersion, datacenterConfig)
	s.AddKnownTypes(anywherev1.GroupVersion, bundle)
	s.AddKnownTypes(anywherev1.GroupVersion, machineConfigCP)

	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).Build()

	validator := vsphere.NewValidator(govcClient, &networkutils.DefaultNetClient{})
	defaulter := vsphere.NewDefaulter(govcClient)

	r := &ClusterReconciler{
		client:    cl,
		log:       logf.Log,
		validator: validator,
		defaulter: defaulter,
	}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	ctx := context.Background()
	govcClient.EXPECT().ValidateVCenterSetupMachineConfig(ctx, datacenterConfig, machineConfigCP, gomock.Any()).Return(fmt.Errorf("error"))
	govcClient.EXPECT().ValidateVCenterSetupMachineConfig(ctx, datacenterConfig, machineConfigWN, gomock.Any()).Return(nil).Times(0)
	govcClient.EXPECT().SearchTemplate(ctx, datacenterConfig.Spec.Datacenter, machineConfigCP).Return("test", nil).Times(0)
	govcClient.EXPECT().GetTags(ctx, machineConfigCP.Spec.Template).Return([]string{"os:ubuntu", fmt.Sprintf("eksdRelease:%s", bundle.Spec.VersionsBundles[0].EksD.Name)}, nil).Times(0)
	govcClient.EXPECT().GetWorkloadAvailableSpace(ctx, machineConfigCP.Spec.Datastore).Return(100.0, nil).Times(0)

	_, err := r.Reconcile(ctx, req)
	if err == nil {
		t.Fatalf("expected and error in the reconcile")
	}

	apiCluster := &anywherev1.Cluster{}

	err = r.client.Get(context.TODO(), req.NamespacedName, apiCluster)
	if err != nil {
		t.Fatalf("get cluster: (%v)", err)
	}
	if apiCluster.Status.FailureMessage == nil {
		t.Errorf("Expected failure reason to be set")
	}
}

// TODO add extra tests

func createWNMachineConfig() *anywherev1.VSphereMachineConfig {
	return &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-wn",
			Namespace: namespace,
		},
		Spec: anywherev1.VSphereMachineConfigSpec{
			DiskGiB:           40,
			Datastore:         "test",
			Folder:            "test",
			NumCPUs:           2,
			MemoryMiB:         16,
			OSFamily:          "ubuntu",
			ResourcePool:      "test",
			StoragePolicyName: "test",
			Template:          "test",
			Users:             nil,
		},
		Status: anywherev1.VSphereMachineConfigStatus{},
	}
}

func createCPMachineConfig() *anywherev1.VSphereMachineConfig {
	return &anywherev1.VSphereMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereMachineConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      name + "-cp",
			Namespace: namespace,
		},
		Spec: anywherev1.VSphereMachineConfigSpec{
			DiskGiB:           40,
			Datastore:         "test",
			Folder:            "test",
			NumCPUs:           2,
			MemoryMiB:         16,
			OSFamily:          "ubuntu",
			ResourcePool:      "test",
			StoragePolicyName: "test",
			Template:          "test",
			Users:             nil,
		},
		Status: anywherev1.VSphereMachineConfigStatus{},
	}
}

func createBundle(cluster *anywherev1.Cluster) *v1alpha1.Bundles {
	return &v1alpha1.Bundles{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
		Spec: v1alpha1.BundlesSpec{
			VersionsBundles: []v1alpha1.VersionsBundle{
				v1alpha1.VersionsBundle{
					KubeVersion: "1.21",
					EksD: v1alpha1.EksDRelease{
						Name:           "test",
						EksDReleaseUrl: "testdata/release.yaml",
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

func createDataCenter(cluster *anywherev1.Cluster) *anywherev1.VSphereDatacenterConfig {
	return &anywherev1.VSphereDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       "VSphereDatacenterConfig",
			APIVersion: "anywhere.eks.amazonaws.com/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datacenter",
			Namespace: cluster.Namespace,
		},
		Spec: anywherev1.VSphereDatacenterConfigSpec{
			Thumbprint: "aaa",
			Server:     "ssss",
			Datacenter: "daaa",
			Network:    "networkA",
		},
		Status: anywherev1.VSphereDatacenterConfigStatus{
			SpecValid: true,
		},
	}
}

func createCluster() *anywherev1.Cluster {
	return &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: anywherev1.ClusterSpec{
			DatacenterRef: anywherev1.Ref{
				Kind: "VSphereDatacenterConfig",
				Name: "datacenter",
			},
			KubernetesVersion: "1.21",
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &anywherev1.Endpoint{
					Host: "1.1.1.1",
				},
				MachineGroupRef: &anywherev1.Ref{
					Kind: "VSphereMachineConfig",
					Name: name + "-cp",
				},
			},
			WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
				anywherev1.WorkerNodeGroupConfiguration{
					Count: 1,
					MachineGroupRef: &anywherev1.Ref{
						Kind: "VSphereMachineConfig",
						Name: name + "-wn",
					},
					Labels: nil,
				},
			},
		},
	}
}

func createSecret() *apiv1.Secret {
	return &apiv1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "eksa-system",
			Name:      vsphere.CredentialsObjectName,
		},
		Data: map[string][]byte{
			"username": []byte("test"),
			"password": []byte("test"),
		},
	}
}
