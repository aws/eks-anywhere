package controllers_test

import (
	"context"
	"errors"
	"testing"

	"github.com/go-logr/logr"
	"github.com/go-logr/logr/testr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/controllers/mocks"
	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	c "github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func TestClusterReconcilerEnsureOwnerReferences(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			EksaVersion:       &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.Spec.IdentityProviderRefs = []anywherev1.Ref{
		{
			Kind: anywherev1.OIDCConfigKind,
			Name: "my-oidc",
		},
		{
			Kind: anywherev1.AWSIamConfigKind,
			Name: "my-iam",
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	oidc := &anywherev1.OIDCConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-oidc",
			Namespace: cluster.Namespace,
		},
	}
	awsIAM := &anywherev1.AWSIamConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-iam",
			Namespace: cluster.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				{
					APIVersion: anywherev1.GroupVersion.String(),
					Kind:       anywherev1.ClusterKind,
					Name:       cluster.Name,
				},
			},
		},
	}
	bundles := createBundle()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	eksaRelease := test.EKSARelease()

	objs := []runtime.Object{cluster, managementCluster, oidc, awsIAM, bundles, secret, eksaRelease}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	iam := newMockAWSIamConfigReconciler(t)
	iam.EXPECT().EnsureCASecret(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)
	iam.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(controller.Result{}, nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), iam, validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))

	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: bundles.Namespace, Name: bundles.Name}, bundles)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: constants.EksaSystemNamespace, Name: cluster.Name + "-kubeconfig"}, secret)).To(Succeed())

	g.Expect(err).NotTo(HaveOccurred())

	newOidc := &anywherev1.OIDCConfig{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-oidc"}, newOidc)).To(Succeed())
	g.Expect(newOidc.OwnerReferences).To(HaveLen(1))
	g.Expect(newOidc.OwnerReferences[0].Name).To(Equal(cluster.Name))

	newAWSIam := &anywherev1.AWSIamConfig{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-iam"}, newAWSIam)).To(Succeed())
	g.Expect(newAWSIam.OwnerReferences).To(HaveLen(1))
	g.Expect(newAWSIam.OwnerReferences[0]).To(Equal(awsIAM.OwnerReferences[0]))
}

func TestClusterReconcilerReconcileChildObjectNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
	}
	cluster.Spec.IdentityProviderRefs = []anywherev1.Ref{
		{
			Kind: anywherev1.OIDCConfigKind,
			Name: "my-oidc",
		},
		{
			Kind: anywherev1.AWSIamConfigKind,
			Name: "my-iam",
		},
	}
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, newMockMachineHealthCheckReconciler(t), nil)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(MatchError(ContainSubstring("not found")))
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal(
			"Dependent cluster objects don't exist: oidcconfigs.anywhere.eks.amazonaws.com \"my-oidc\" not found",
		)))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.MissingDependentObjectsReason)))
	})
}

func TestClusterReconcilerSetupWithManager(t *testing.T) {
	client := env.Client()
	r := controllers.NewClusterReconciler(client, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil, nil)

	g := NewWithT(t)
	g.Expect(r.SetupWithManager(env.Manager(), env.Manager().GetLogger())).To(Succeed())
}

func TestClusterReconcilerManagementClusterNotFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
	}
	cluster.SetManagedBy("my-management-cluster-2")

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cb.WithIndex(&anywherev1.Cluster{}, "metadata.name", clientutil.ClusterNameIndexer)
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()
	api := envtest.NewAPIExpecter(t, cl)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), newMockClusterValidator(t), nil, nil, nil)
	g.Expect(r.Reconcile(ctx, clusterRequest(cluster))).Error().To(BeNil())

	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("Management cluster my-management-cluster-2 does not exist")))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.ManagementClusterRefInvalidReason)))
	})
}

func TestClusterReconcilerSetBundlesRef(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-management-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "default",
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			BundlesRef: &anywherev1.BundlesRef{
				Name:      "bundles-1",
				Namespace: "default",
			},
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster-kubeconfig",
			Namespace: constants.EksaSystemNamespace,
		},
	}
	bundles := createBundle()

	objs := []runtime.Object{cluster, managementCluster, secret, bundles}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Spec.BundlesRef.Namespace, Name: cluster.Spec.BundlesRef.Name}, bundles)).To(Succeed())
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: constants.EksaSystemNamespace, Name: cluster.Name + "-kubeconfig"}, secret)).To(Succeed())
	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	newCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-cluster"}, newCluster)).To(Succeed())
	g.Expect(newCluster.Spec.BundlesRef).To(Equal(mgmtCluster.Spec.BundlesRef))
}

func TestClusterReconcilerSetDefaultEksaVersion(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "default",
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	bundles := createBundle()
	cluster.SetManagedBy("my-management-cluster")

	objs := []runtime.Object{cluster, managementCluster, test.EKSARelease(), bundles}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	mgmtCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: managementCluster.Name}, mgmtCluster)).To(Succeed())
	pcc := newMockPackagesClient(t)
	pcc.EXPECT().Reconcile(ctx, gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	mhc := newMockMachineHealthCheckReconciler(t)
	mhc.EXPECT().Reconcile(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, pcc, mhc, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())

	newCluster := &anywherev1.Cluster{}
	g.Expect(cl.Get(ctx, client.ObjectKey{Namespace: cluster.Namespace, Name: "my-cluster"}, newCluster)).To(Succeed())
	g.Expect(newCluster.Spec.EksaVersion).To(Equal(mgmtCluster.Spec.EksaVersion))
}

func TestClusterReconcilerWorkloadClusterMgmtClusterNameFail(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := test.DevEksaVersion()

	managementCluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-management-cluster",
			Namespace: "my-namespace",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-cluster",
			Namespace: "my-namespace",
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	cluster.SetManagedBy("my-management-cluster")
	// clusterSpec := &c.Spec{
	// 	Config: &c.Config{
	// 		Cluster: cluster,
	// 	},
	// }

	objs := []runtime.Object{cluster, managementCluster}
	cb := fake.NewClientBuilder()
	cl := cb.WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	validator := newMockClusterValidator(t)
	validator.EXPECT().ValidateManagementClusterName(ctx, gomock.AssignableToTypeOf(logr.Logger{}), gomock.AssignableToTypeOf(cluster)).
		Return(errors.New("test error"))

	r := controllers.NewClusterReconciler(cl, newRegistryForDummyProviderReconciler(), newMockAWSIamConfigReconciler(t), validator, nil, nil, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(HaveOccurred())

	api := envtest.NewAPIExpecter(t, cl)
	c := envtest.CloneNameNamespace(cluster)
	api.ShouldEventuallyMatch(ctx, c, func(g Gomega) {
		g.Expect(c.Status.FailureMessage).To(HaveValue(Equal("test error")))
		g.Expect(c.Status.FailureReason).To(HaveValue(Equal(anywherev1.ManagementClusterRefInvalidReason)))
	})
}

func TestClusterReconcilerNoBundleFound(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := anywherev1.EksaVersion("v0.22.0")

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}

	kcp := testKubeadmControlPlaneFromCluster(cluster)

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	eksaReleaseV022 := test.EKSARelease()
	eksaReleaseV022.Name = "eksa-v0-22-0"
	eksaReleaseV022.Spec.Version = "eksa-v0-22-0"
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaReleaseV022).
		WithStatusSubresource(cluster).
		Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(MatchError(ContainSubstring("getting bundle for cluster")))
}

func TestClusterReconcilerFailSignatureValidation(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()
	version := anywherev1.EksaVersion("v0.22.0")

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-cluster",
		},
		Spec: anywherev1.ClusterSpec{
			EksaVersion: &version,
		},
		Status: anywherev1.ClusterStatus{
			ReconciledGeneration: 1,
		},
	}
	eksaReleaseV022 := test.EKSARelease()
	eksaReleaseV022.Name = "eksa-v0-22-0"
	eksaReleaseV022.Spec.Version = "eksa-v0-22-0"
	bundles := createBundle()
	bundles.Spec.VersionsBundles[0].KubeVersion = "1.25"

	kcp := testKubeadmControlPlaneFromCluster(cluster)

	controller := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(controller)
	iam := mocks.NewMockAWSIamConfigReconciler(controller)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(controller)

	clusterValidator := mocks.NewMockClusterValidator(controller)
	registry := newRegistryMock(providerReconciler)
	c := fake.NewClientBuilder().WithRuntimeObjects(cluster, kcp, eksaReleaseV022, bundles).
		WithStatusSubresource(cluster).
		Build()
	mockPkgs := mocks.NewMockPackagesClient(controller)

	r := controllers.NewClusterReconciler(c, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)
	_, err := r.Reconcile(ctx, clusterRequest(cluster))
	g.Expect(err).To(MatchError(ContainSubstring("validating bundle signature")))
}

func TestClusterReconcilerCertificateStatusStatus(t *testing.T) {
	tests := []struct {
		name                      string
		cluster                   *anywherev1.Cluster
		kcp                       *controlplanev1.KubeadmControlPlane
		machineDeployments        []clusterv1.Machine
		wantCertificateStatusCall bool
		clusterReady              bool
	}{
		{
			name: "cluster ready - should update certificate status",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: anywherev1.Kube132,
					ClusterNetwork: anywherev1.ClusterNetwork{
						CNIConfig: &anywherev1.CNIConfig{
							Cilium: &anywherev1.CiliumConfig{},
						},
					},
					ManagementCluster: anywherev1.ManagementCluster{Name: "management-cluster"},
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
						Endpoint: &anywherev1.Endpoint{
							Host: "127.0.0.1",
						},
					},
					MachineHealthCheck: &anywherev1.MachineHealthCheck{
						UnhealthyMachineTimeout: &metav1.Duration{
							Duration: constants.DefaultUnhealthyMachineTimeout,
						},
						NodeStartupTimeout: &metav1.Duration{
							Duration: constants.DefaultNodeStartupTimeout,
						},
					},
				},
			},
			kcp: testKubeadmControlPlaneFromCluster(&anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
			}),
			machineDeployments: []clusterv1.Machine{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-control-plane-1",
						Namespace: constants.EksaSystemNamespace,
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name":  "test-cluster",
							"cluster.x-k8s.io/control-plane": "",
						},
					},
					Status: clusterv1.MachineStatus{
						Addresses: []clusterv1.MachineAddress{
							{
								Type:    clusterv1.MachineExternalIP,
								Address: "192.168.1.100",
							},
						},
					},
				},
			},
			wantCertificateStatusCall: true,
			clusterReady:              true,
		},
		{
			name: "cluster not ready - should skip certificate status update",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Spec: anywherev1.ClusterSpec{
					KubernetesVersion: anywherev1.Kube132,
					ClusterNetwork: anywherev1.ClusterNetwork{
						CNIConfig: &anywherev1.CNIConfig{
							Cilium: &anywherev1.CiliumConfig{},
						},
					},
					ManagementCluster: anywherev1.ManagementCluster{Name: "management-cluster"},
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						Count: 1,
					},
					MachineHealthCheck: &anywherev1.MachineHealthCheck{
						UnhealthyMachineTimeout: &metav1.Duration{
							Duration: constants.DefaultUnhealthyMachineTimeout,
						},
						NodeStartupTimeout: &metav1.Duration{
							Duration: constants.DefaultNodeStartupTimeout,
						},
					},
				},
			},
			kcp: &controlplanev1.KubeadmControlPlane{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: constants.EksaSystemNamespace,
				},
				Status: controlplanev1.KubeadmControlPlaneStatus{
					Conditions: clusterv1.Conditions{
						{
							Type:   controlplanev1.AvailableCondition,
							Status: "False",
						},
					},
				},
			},
			machineDeployments:        []clusterv1.Machine{},
			wantCertificateStatusCall: false,
			clusterReady:              false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()
			version := test.DevEksaVersion()
			tt.cluster.Spec.EksaVersion = &version

			mgmt := tt.cluster.DeepCopy()
			mgmt.Name = "management-cluster"

			objs := make([]runtime.Object, 0, 4+len(tt.machineDeployments))
			objs = append(objs, tt.cluster, mgmt, test.EKSARelease(), createBundle())

			if tt.kcp != nil {
				objs = append(objs, tt.kcp)
			}

			for _, machine := range tt.machineDeployments {
				objs = append(objs, machine.DeepCopy())
			}

			client := fake.NewClientBuilder().WithRuntimeObjects(objs...).
				WithStatusSubresource(tt.cluster).
				Build()

			mockCtrl := gomock.NewController(t)
			providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
			iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
			clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
			registry := newRegistryMock(providerReconciler)
			mockPkgs := mocks.NewMockPackagesClient(mockCtrl)
			mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(mockCtrl)

			log := testr.New(t)
			logCtx := ctrl.LoggerInto(ctx, log)

			iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Return(controller.Result{}, nil)
			iam.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Return(controller.Result{}, nil)

			if tt.clusterReady {
				// Set up conditions to make cluster appear ready
				providerReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Times(1).Do(
					func(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) {
						conditions.MarkTrue(cluster, anywherev1.DefaultCNIConfiguredCondition)
					},
				)
			} else {
				providerReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Times(1).Do(
					func(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) {
						conditions.MarkFalse(cluster, anywherev1.DefaultCNIConfiguredCondition, anywherev1.ControlPlaneNotReadyReason, clusterv1.ConditionSeverityInfo, "")
					},
				)
			}

			clusterValidator.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Return(nil)
			mockPkgs.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
			mhcReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(tt.cluster)).Return(nil)

			r := controllers.NewClusterReconciler(client, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)

			result, err := r.Reconcile(logCtx, clusterRequest(tt.cluster))

			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(result).To(Equal(ctrl.Result{}))

			// Verify certificate status was updated if cluster is ready
			api := envtest.NewAPIExpecter(t, client)
			c := envtest.CloneNameNamespace(tt.cluster)
			api.ShouldEventuallyMatch(logCtx, c, func(g Gomega) {
				if tt.wantCertificateStatusCall && tt.clusterReady {
					// Certificate status should be populated (even if empty due to connection failures in test)
					g.Expect(c.Status.ClusterCertificateInfo).ToNot(BeNil())
				}
			})
		})
	}
}

func TestClusterReconcilerCertificateStatusError(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cluster := &anywherev1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: anywherev1.ClusterSpec{
			KubernetesVersion: anywherev1.Kube132,
			ClusterNetwork: anywherev1.ClusterNetwork{
				CNIConfig: &anywherev1.CNIConfig{
					Cilium: &anywherev1.CiliumConfig{},
				},
			},
			ManagementCluster: anywherev1.ManagementCluster{Name: "management-cluster"},
			ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
				Count: 1,
				Endpoint: &anywherev1.Endpoint{
					Host: "127.0.0.1",
				},
			},
			MachineHealthCheck: &anywherev1.MachineHealthCheck{
				UnhealthyMachineTimeout: &metav1.Duration{
					Duration: constants.DefaultUnhealthyMachineTimeout,
				},
				NodeStartupTimeout: &metav1.Duration{
					Duration: constants.DefaultNodeStartupTimeout,
				},
			},
		},
	}

	version := test.DevEksaVersion()
	cluster.Spec.EksaVersion = &version

	mgmt := cluster.DeepCopy()
	mgmt.Name = "management-cluster"

	kcp := testKubeadmControlPlaneFromCluster(cluster)

	objs := []runtime.Object{cluster, mgmt, kcp, test.EKSARelease(), createBundle()}

	client := fake.NewClientBuilder().WithRuntimeObjects(objs...).
		WithStatusSubresource(cluster).
		Build()

	mockCtrl := gomock.NewController(t)
	providerReconciler := mocks.NewMockProviderClusterReconciler(mockCtrl)
	iam := mocks.NewMockAWSIamConfigReconciler(mockCtrl)
	clusterValidator := mocks.NewMockClusterValidator(mockCtrl)
	registry := newRegistryMock(providerReconciler)
	mockPkgs := mocks.NewMockPackagesClient(mockCtrl)
	mhcReconciler := mocks.NewMockMachineHealthCheckReconciler(mockCtrl)

	log := testr.New(t)
	logCtx := ctrl.LoggerInto(ctx, log)

	iam.EXPECT().EnsureCASecret(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Return(controller.Result{}, nil)
	iam.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Return(controller.Result{}, nil)

	// Make cluster appear ready so certificate status update is attempted
	providerReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Times(1).Do(
		func(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) {
			conditions.MarkTrue(cluster, anywherev1.DefaultCNIConfiguredCondition)
		},
	)

	clusterValidator.EXPECT().ValidateManagementClusterName(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Return(nil)
	mockPkgs.EXPECT().Reconcile(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any())
	mhcReconciler.EXPECT().Reconcile(logCtx, gomock.AssignableToTypeOf(logr.Logger{}), sameName(cluster)).Return(nil)

	r := controllers.NewClusterReconciler(client, registry, iam, clusterValidator, mockPkgs, mhcReconciler, nil)

	// This should not fail even if certificate status update encounters errors
	result, err := r.Reconcile(logCtx, clusterRequest(cluster))
	g.Expect(err).ToNot(HaveOccurred())
	g.Expect(result).To(Equal(ctrl.Result{}))
}

func newRegistryForDummyProviderReconciler() controllers.ProviderClusterReconcilerRegistry {
	return newRegistryMock(dummyProviderReconciler{})
}

func newRegistryMock(reconciler clusters.ProviderClusterReconciler) dummyProviderReconcilerRegistry {
	return dummyProviderReconcilerRegistry{
		reconciler: reconciler,
	}
}

type dummyProviderReconcilerRegistry struct {
	reconciler clusters.ProviderClusterReconciler
}

func (d dummyProviderReconcilerRegistry) Get(_ string) clusters.ProviderClusterReconciler {
	return d.reconciler
}

type dummyProviderReconciler struct{}

func (dummyProviderReconciler) Reconcile(ctx context.Context, log logr.Logger, cluster *anywherev1.Cluster) (controller.Result, error) {
	return controller.Result{}, nil
}

func (dummyProviderReconciler) ReconcileCNI(ctx context.Context, log logr.Logger, clusterSpec *c.Spec) (controller.Result, error) {
	return controller.Result{}, nil
}

func clusterRequest(cluster *anywherev1.Cluster) reconcile.Request {
	return reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      cluster.Name,
			Namespace: cluster.Namespace,
		},
	}
}

func nullLog() logr.Logger {
	return logr.New(logf.NullLogSink{})
}

func newMockAWSIamConfigReconciler(t *testing.T) *mocks.MockAWSIamConfigReconciler {
	ctrl := gomock.NewController(t)
	return mocks.NewMockAWSIamConfigReconciler(ctrl)
}

func newMockClusterValidator(t *testing.T) *mocks.MockClusterValidator {
	ctrl := gomock.NewController(t)
	return mocks.NewMockClusterValidator(ctrl)
}

func newMockPackagesClient(t *testing.T) *mocks.MockPackagesClient {
	ctrl := gomock.NewController(t)
	return mocks.NewMockPackagesClient(ctrl)
}

func newMockMachineHealthCheckReconciler(t *testing.T) *mocks.MockMachineHealthCheckReconciler {
	ctrl := gomock.NewController(t)
	return mocks.NewMockMachineHealthCheckReconciler(ctrl)
}
