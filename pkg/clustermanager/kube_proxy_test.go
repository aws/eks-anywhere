package clustermanager_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/internal/test/envtest"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/clustermanager"
	"github.com/aws/eks-anywhere/pkg/clustermanager/mocks"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/controller/clientutil"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type prepareKubeProxyTest struct {
	ctx              context.Context
	log              logr.Logger
	spec             *cluster.Spec
	kcp              *controlplanev1.KubeadmControlPlane
	kubeProxy        *appsv1.DaemonSet
	nodeCP           *corev1.Node
	nodeWorker       *corev1.Node
	kubeProxyCP      *corev1.Pod
	kubeProxyWorker  *corev1.Pod
	managementClient kubernetes.Client
	// managementImplClient is a controller-runtime client that serves as the
	// underlying implementation for managementClient.
	managementImplClient client.Client
	workloadClient       kubernetes.Client
	// workloadImplClient is a controller-runtime client that serves as the
	// underlying implementation for workloadClient.
	workloadImplClient          client.Client
	workloadClusterExtraObjects []client.Object
}

func newPrepareKubeProxyTest() *prepareKubeProxyTest {
	tt := &prepareKubeProxyTest{}

	tt.ctx = context.Background()
	tt.log = test.NewNullLogger()
	tt.spec = test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = "my-cluster-test"
		s.Cluster.Spec.KubernetesVersion = anywherev1.Kube123
		s.VersionsBundles["1.23"] = test.VersionBundle()
		s.VersionsBundles["1.23"].KubeDistro.EKSD.Channel = "1.23"
		s.VersionsBundles["1.23"].KubeDistro.EKSD.Number = 18
		s.VersionsBundles["1.23"].KubeDistro.KubeProxy.URI = "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-18"
	})
	tt.kcp = &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterapi.KubeadmControlPlaneName(tt.spec.Cluster),
			Namespace: constants.EksaSystemNamespace,
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			Version: "v1.23.16-eks-1-23-16",
			KubeadmConfigSpec: v1beta1.KubeadmConfigSpec{
				ClusterConfiguration: &v1beta1.ClusterConfiguration{
					ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
				},
			},
		},
	}
	tt.kubeProxy = &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy",
			Namespace: "kube-system",
		},
		Spec: appsv1.DaemonSetSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Image: "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-16",
						},
					},
				},
			},
		},
	}
	tt.nodeCP = &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cp",
		},
	}
	tt.nodeWorker = &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "worker",
		},
	}
	tt.kubeProxyCP = &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-cp",
			Namespace: "kube-system",
			Labels: map[string]string{
				"k8s-app": "kube-proxy",
			},
		},
	}
	tt.kubeProxyWorker = tt.kubeProxyCP.DeepCopy()
	tt.kubeProxyWorker.Name = "kube-proxy-worker"

	return tt
}

func (tt *prepareKubeProxyTest) initClients(tb testing.TB) {
	tt.managementImplClient = fake.NewClientBuilder().WithObjects(tt.kcp).Build()
	tt.managementClient = clientutil.NewKubeClient(tt.managementImplClient)

	objs := []client.Object{
		tt.kubeProxy,
		tt.kubeProxyCP,
		tt.kubeProxyWorker,
		tt.nodeCP,
		tt.nodeWorker,
	}
	objs = append(objs, tt.workloadClusterExtraObjects...)

	tt.workloadImplClient = fake.NewClientBuilder().WithObjects(objs...).Build()
	tt.workloadClient = clientutil.NewKubeClient(tt.workloadImplClient)
}

// startKCPControllerEmulator stars a routine that reverts the kube-proxy
// version updated for n times and then stops. This is useful to simulate the real
// KCP controller behavior when it hasn't yet seen the skip annotation and it
// keeps reverting the kube-proxy image tag.
func (tt *prepareKubeProxyTest) startKCPControllerEmulator(tb testing.TB, times int) {
	go func() {
		api := envtest.NewAPIExpecter(tb, tt.workloadImplClient)
		kubeProxy := tt.kubeProxy.DeepCopy()
		originalImage := kubeProxy.Spec.Template.Spec.Containers[0].Image
		for i := 0; i < times; i++ {
			api.ShouldEventuallyMatch(tt.ctx, kubeProxy, func(g Gomega) {
				// Wait until the image has been updated by KubeProxyUpgrader
				currentImage := kubeProxy.Spec.Template.Spec.Containers[0].Image
				g.Expect(currentImage).NotTo(Equal(originalImage))
				// Then revert the change
				kubeProxy.Spec.Template.Spec.Containers[0].Image = originalImage
				g.Expect(tt.workloadClient.Update(tt.ctx, kubeProxy))
			})
		}
	}()
}

func TestKubeProxyUpgraderPrepareForUpgradeSuccess(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.initClients(t)
	// Revert the kube-proxy image update twice
	tt.startKCPControllerEmulator(t, 2)
	u := clustermanager.NewKubeProxyUpgrader(
		clustermanager.WithUpdateKubeProxyTiming(4, 100*time.Millisecond),
	)

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(Succeed())

	managementAPI := envtest.NewAPIExpecter(t, tt.managementImplClient)
	managementAPI.ShouldEventuallyMatch(tt.ctx, tt.kcp, func(g Gomega) {
		g.Expect(tt.kcp.Annotations).To(HaveKeyWithValue(controlplanev1.SkipKubeProxyAnnotation, "true"))
	})

	workloadAPI := envtest.NewAPIExpecter(t, tt.workloadImplClient)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.kubeProxy, func(g Gomega) {
		image := tt.kubeProxy.Spec.Template.Spec.Containers[0].Image
		g.Expect(image).To(Equal("public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-18"))

		firstMatchExpression := tt.kubeProxy.Spec.Template.Spec.Affinity.
			NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
		g.Expect(firstMatchExpression.Key).To(Equal("anywhere.eks.amazonaws.com/iptableslegacy"))
		g.Expect(firstMatchExpression.Operator).To(Equal(corev1.NodeSelectorOpDoesNotExist))
	})
	workloadAPI.ShouldEventuallyNotExist(tt.ctx, tt.kubeProxyCP)
	workloadAPI.ShouldEventuallyNotExist(tt.ctx, tt.kubeProxyWorker)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeCP, func(g Gomega) {
		g.Expect(tt.nodeCP.Labels).To(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeWorker, func(g Gomega) {
		g.Expect(tt.nodeWorker.Labels).To(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})

	legacyKubeProxy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-iptables-legacy",
			Namespace: "kube-system",
		},
	}
	workloadAPI.ShouldEventuallyExist(tt.ctx, legacyKubeProxy)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, legacyKubeProxy, func(g Gomega) {
		image := legacyKubeProxy.Spec.Template.Spec.Containers[0].Image
		g.Expect(image).To(Equal("public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-16"))

		firstMatchExpression := legacyKubeProxy.Spec.Template.Spec.Affinity.
			NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[0].MatchExpressions[0]
		g.Expect(firstMatchExpression.Key).To(Equal("anywhere.eks.amazonaws.com/iptableslegacy"))
		g.Expect(firstMatchExpression.Operator).To(Equal(corev1.NodeSelectorOpExists))
	})
}

func TestKubeProxyUpgraderPrepareForUpgradeNoKCP(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kcp = &controlplanev1.KubeadmControlPlane{} // no kcp
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(MatchError(ContainSubstring("reading the kubeadm control plane for an upgrade")))
}

func TestKubeProxyUpgraderPrepareForUpgradeNoKubeProxy(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kubeProxy = &appsv1.DaemonSet{} // no kube-proxy
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(MatchError(ContainSubstring("reading kube-proxy for upgrade")))
}

func TestKubeProxyUpgraderPrepareForUpgradeinvalidEKDDTag(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kcp.Spec.Version = "1.23.16-eks-1-23"
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(MatchError(ContainSubstring("invalid eksd tag format")))
}

func TestKubeProxyUpgraderPrepareForUpgradeAlreadyUsingNewKubeProxy(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kcp.Spec.Version = "1.23.16-eks-1-23-18"
	tt.kubeProxy.Spec.Template.Spec.Containers[0].Image = "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-18"
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(Succeed())

	managementAPI := envtest.NewAPIExpecter(t, tt.managementImplClient)
	managementAPI.ShouldEventuallyMatch(tt.ctx, tt.kcp, func(g Gomega) {
		g.Expect(tt.kcp.Annotations).NotTo(HaveKeyWithValue(controlplanev1.SkipKubeProxyAnnotation, "true"))
	})

	workloadAPI := envtest.NewAPIExpecter(t, tt.workloadImplClient)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.kubeProxy, func(g Gomega) {
		image := tt.kubeProxy.Spec.Template.Spec.Containers[0].Image
		g.Expect(image).To(Equal("public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-18"))
		g.Expect(tt.kubeProxy.Spec.Template.Spec.Affinity).To(BeNil())
	})
	workloadAPI.ShouldEventuallyExist(tt.ctx, tt.kubeProxyCP)
	workloadAPI.ShouldEventuallyExist(tt.ctx, tt.kubeProxyWorker)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeCP, func(g Gomega) {
		g.Expect(tt.nodeCP.Labels).NotTo(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeWorker, func(g Gomega) {
		g.Expect(tt.nodeWorker.Labels).NotTo(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})

	legacyKubeProxy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-iptables-legacy",
			Namespace: "kube-system",
		},
	}
	workloadAPI.ShouldEventuallyNotExist(tt.ctx, legacyKubeProxy)
}

func TestKubeProxyUpgraderPrepareForUpgradeNewSpecHasOldKubeProxy(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.spec.VersionsBundles["1.23"].KubeDistro.EKSD.Channel = "1.23"
	tt.spec.VersionsBundles["1.23"].KubeDistro.EKSD.Number = 16
	tt.spec.VersionsBundles["1.23"].KubeDistro.KubeProxy.URI = "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-16"
	tt.kcp.Spec.Version = "1.23.16-eks-1-23-15"
	tt.kubeProxy.Spec.Template.Spec.Containers[0].Image = "public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-15"
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.PrepareForUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(Succeed())

	managementAPI := envtest.NewAPIExpecter(t, tt.managementImplClient)
	managementAPI.ShouldEventuallyMatch(tt.ctx, tt.kcp, func(g Gomega) {
		g.Expect(tt.kcp.Annotations).NotTo(HaveKeyWithValue(controlplanev1.SkipKubeProxyAnnotation, "true"))
	})

	workloadAPI := envtest.NewAPIExpecter(t, tt.workloadImplClient)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.kubeProxy, func(g Gomega) {
		image := tt.kubeProxy.Spec.Template.Spec.Containers[0].Image
		g.Expect(image).To(Equal("public.ecr.aws/eks-distro/kubernetes/kube-proxy:v1.23.16-eks-1-23-15"))
		g.Expect(tt.kubeProxy.Spec.Template.Spec.Affinity).To(BeNil())
	})
	workloadAPI.ShouldEventuallyExist(tt.ctx, tt.kubeProxyCP)
	workloadAPI.ShouldEventuallyExist(tt.ctx, tt.kubeProxyWorker)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeCP, func(g Gomega) {
		g.Expect(tt.nodeCP.Labels).NotTo(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.nodeWorker, func(g Gomega) {
		g.Expect(tt.nodeWorker.Labels).NotTo(HaveKeyWithValue("anywhere.eks.amazonaws.com/iptableslegacy", "true"))
	})

	legacyKubeProxy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-iptables-legacy",
			Namespace: "kube-system",
		},
	}
	workloadAPI.ShouldEventuallyNotExist(tt.ctx, legacyKubeProxy)
}

func TestKubeProxyUpgraderCleanupAfterUpgradeSuccessWithReentry(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.workloadClusterExtraObjects = append(tt.workloadClusterExtraObjects, &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-iptables-legacy",
			Namespace: "kube-system",
		},
	})
	tt.kubeProxy.Spec.Template.Spec.Affinity = &corev1.Affinity{}
	clientutil.AddAnnotation(tt.kcp, controlplanev1.SkipKubeProxyAnnotation, "true")
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(Succeed())

	managementAPI := envtest.NewAPIExpecter(t, tt.managementImplClient)
	managementAPI.ShouldEventuallyMatch(tt.ctx, tt.kcp, func(g Gomega) {
		g.Expect(tt.kcp.Annotations).NotTo(HaveKeyWithValue(controlplanev1.SkipKubeProxyAnnotation, "true"))
	})

	workloadAPI := envtest.NewAPIExpecter(t, tt.workloadImplClient)
	workloadAPI.ShouldEventuallyMatch(tt.ctx, tt.kubeProxy, func(g Gomega) {
		g.Expect(tt.kubeProxy.Spec.Template.Spec.Affinity).To(BeNil())
	})

	legacyKubeProxy := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kube-proxy-iptables-legacy",
			Namespace: "kube-system",
		},
	}
	workloadAPI.ShouldEventuallyNotExist(tt.ctx, legacyKubeProxy)

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(Succeed())
}

func TestKubeProxyCleanupAfterUpgradeNoKubeProxy(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kubeProxy = &appsv1.DaemonSet{} // no kube-proxy
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(MatchError(ContainSubstring("reading kube-proxy for upgrade")))
}

func TestKubeProxyCleanupAfterUpgradeNoKCP(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kcp = &controlplanev1.KubeadmControlPlane{} // no kcp
	tt.initClients(t)
	u := clustermanager.NewKubeProxyUpgrader()

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.log, tt.managementClient, tt.workloadClient, tt.spec),
	).To(MatchError(ContainSubstring("reading the kubeadm control plane to cleanup the skip annotations")))
}

func TestEKSDVersionAndNumberFromTag(t *testing.T) {
	tests := []struct {
		name            string
		tag             string
		wantKubeVersion anywherev1.KubernetesVersion
		wantNumber      int
		wantErr         string
	}{
		{
			name:            "valid eks-d",
			tag:             "v1.23.16-eks-1-23-16",
			wantKubeVersion: anywherev1.Kube123,
			wantNumber:      16,
		},
		{
			name:    "invalid eks-d, no number",
			tag:     "v1.23.16-eks-1-23",
			wantErr: "invalid eksd tag format",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			kuberVersion, number, err := clustermanager.EKSDVersionAndNumberFromTag(tt.tag)
			if tt.wantErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			} else {
				g.Expect(kuberVersion).To(Equal(tt.wantKubeVersion))
				g.Expect(number).To(Equal(tt.wantNumber))
			}
		})
	}
}

func TestEKSDIncludesNewKubeProxy(t *testing.T) {
	tests := []struct {
		name        string
		kubeVersion anywherev1.KubernetesVersion
		number      int
		want        bool
	}{
		{
			name:        "eksd 1.22-23",
			kubeVersion: anywherev1.Kube122,
			number:      23,
			want:        true,
		},
		{
			name:        "eksd 1.22-22",
			kubeVersion: anywherev1.Kube122,
			number:      22,
			want:        true,
		},
		{
			name:        "eksd 1.22-16",
			kubeVersion: anywherev1.Kube122,
			number:      16,
			want:        false,
		},

		{
			name:        "eksd 1.23-18",
			kubeVersion: anywherev1.Kube123,
			number:      18,
			want:        true,
		},
		{
			name:        "eksd 1.23-17",
			kubeVersion: anywherev1.Kube123,
			number:      17,
			want:        true,
		},
		{
			name:        "eksd 1.23-16",
			kubeVersion: anywherev1.Kube123,
			number:      16,
			want:        false,
		},

		{
			name:        "eksd 1.24-13",
			kubeVersion: anywherev1.Kube124,
			number:      13,
			want:        true,
		},
		{
			name:        "eksd 1.24-12",
			kubeVersion: anywherev1.Kube124,
			number:      12,
			want:        true,
		},
		{
			name:        "eksd 1.24-11",
			kubeVersion: anywherev1.Kube124,
			number:      11,
			want:        false,
		},

		{
			name:        "eksd 1.25-9",
			kubeVersion: anywherev1.Kube125,
			number:      9,
			want:        true,
		},
		{
			name:        "eksd 1.25-8",
			kubeVersion: anywherev1.Kube125,
			number:      8,
			want:        true,
		},
		{
			name:        "eksd 1.25-7",
			kubeVersion: anywherev1.Kube125,
			number:      7,
			want:        false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(clustermanager.EKSDIncludesNewKubeProxy(tt.kubeVersion, tt.number)).To(Equal(tt.want))
		})
	}
}

func TestKubeProxyCLIUpgraderPrepareUpgradeErrorManagementClusterClient(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(nil, errors.New("building management client")).Times(2)

	u := clustermanager.NewKubeProxyCLIUpgrader(
		test.NewNullLogger(),
		factory,
		clustermanager.KubeProxyCLIUpgraderRetrier(*retrier.NewWithMaxRetries(2, 0)),
	)

	g.Expect(
		u.PrepareUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(MatchError(ContainSubstring("building management client")))
}

func TestKubeProxyCLIUpgraderPrepareUpgradeErrorWorkloadClusterClient(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.initClients(t)
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(tt.managementClient, nil)
	factory.EXPECT().BuildClientFromKubeconfig(workloadKubeConfig).Return(nil, errors.New("building workload client")).Times(2)

	u := clustermanager.NewKubeProxyCLIUpgrader(
		test.NewNullLogger(),
		factory,
		clustermanager.KubeProxyCLIUpgraderRetrier(*retrier.NewWithMaxRetries(2, 0)),
	)

	g.Expect(
		u.PrepareUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(MatchError(ContainSubstring("building workload client")))
}

func TestKubeProxyCLIUpgraderPrepareUpgradeSuccess(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.initClients(t)
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(tt.managementClient, nil)
	factory.EXPECT().BuildClientFromKubeconfig(workloadKubeConfig).Return(tt.workloadClient, nil)

	u := clustermanager.NewKubeProxyCLIUpgrader(test.NewNullLogger(), factory)

	g.Expect(
		u.PrepareUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(Succeed())
}

func TestKubeProxyCLIUpgraderPrepareUpgradeErrorInPrepare(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.kcp = &controlplanev1.KubeadmControlPlane{} // no kcp
	tt.initClients(t)
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(tt.managementClient, nil)
	factory.EXPECT().BuildClientFromKubeconfig(workloadKubeConfig).Return(tt.workloadClient, nil)

	u := clustermanager.NewKubeProxyCLIUpgrader(
		test.NewNullLogger(),
		factory,
		clustermanager.KubeProxyCLIUpgraderRetrier(*retrier.NewWithMaxRetries(1, 0)),
	)

	g.Expect(
		u.PrepareUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(MatchError(ContainSubstring("reading the kubeadm control plane for an upgrade")))
}

func TestKubeProxyCLIUpgraderCleanupAfterUpgradeSuccess(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.initClients(t)
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(tt.managementClient, nil)
	factory.EXPECT().BuildClientFromKubeconfig(workloadKubeConfig).Return(tt.workloadClient, nil)

	u := clustermanager.NewKubeProxyCLIUpgrader(test.NewNullLogger(), factory)

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(Succeed())
}

func TestKubeProxyCLICleanupAfterUpgradeErrorWorkloadClusterClient(t *testing.T) {
	g := NewWithT(t)
	tt := newPrepareKubeProxyTest()
	tt.initClients(t)
	managementKubeConfig := "mngmt.yaml"
	workloadKubeConfig := "workload.yaml"
	ctrl := gomock.NewController(t)
	factory := mocks.NewMockClientFactory(ctrl)
	factory.EXPECT().BuildClientFromKubeconfig(managementKubeConfig).Return(tt.managementClient, nil)
	factory.EXPECT().BuildClientFromKubeconfig(workloadKubeConfig).Return(nil, errors.New("building workload client")).Times(2)

	u := clustermanager.NewKubeProxyCLIUpgrader(
		test.NewNullLogger(),
		factory,
		clustermanager.KubeProxyCLIUpgraderRetrier(*retrier.NewWithMaxRetries(2, 0)),
	)

	g.Expect(
		u.CleanupAfterUpgrade(tt.ctx, tt.spec, managementKubeConfig, workloadKubeConfig),
	).To(MatchError(ContainSubstring("building workload client")))
}
