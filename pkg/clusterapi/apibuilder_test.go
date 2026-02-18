package clusterapi_test

import (
	"testing"

	etcdbootstrapv1 "github.com/aws/etcdadm-bootstrap-provider/api/v1beta1"
	etcdv1 "github.com/aws/etcdadm-controller/api/v1beta1"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"
	clusterv1beta2 "sigs.k8s.io/cluster-api/api/core/v1beta2"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type apiBuilerTest struct {
	*WithT
	clusterSpec             *cluster.Spec
	workerNodeGroupConfig   *anywherev1.WorkerNodeGroupConfiguration
	kubeadmConfigTemplate   *bootstrapv1beta2.KubeadmConfigTemplate
	providerCluster         clusterapi.APIObject
	controlPlane            clusterapi.APIObject
	providerMachineTemplate clusterapi.APIObject
	unstackedEtcdCluster    clusterapi.APIObject
}

type providerCluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (c *providerCluster) DeepCopyObject() runtime.Object {
	return nil
}

type providerMachineTemplate struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (c *providerMachineTemplate) DeepCopyObject() runtime.Object {
	return nil
}

type unstackedEtcdCluster struct {
	metav1.TypeMeta
	metav1.ObjectMeta
}

func (c *unstackedEtcdCluster) DeepCopyObject() runtime.Object {
	return nil
}

func newApiBuilerTest(t *testing.T) apiBuilerTest {
	clusterSpec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster = &anywherev1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster",
				Namespace: "my-namespace",
			},
			Spec: anywherev1.ClusterSpec{
				ClusterNetwork: anywherev1.ClusterNetwork{
					Pods: anywherev1.Pods{
						CidrBlocks: []string{
							"1.2.3.4/5",
						},
					},
					Services: anywherev1.Services{
						CidrBlocks: []string{
							"1.2.3.4/5",
						},
					},
				},
				ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
					Endpoint: &anywherev1.Endpoint{
						Host: "1.2.3.4",
					},
					Count: 3,
					Taints: []v1.Taint{
						{
							Key:       "key1",
							Value:     "val1",
							Effect:    v1.TaintEffectNoExecute,
							TimeAdded: nil,
						},
					},
					Labels: map[string]string{
						"key1": "val1",
						"key2": "val2",
					},
					CertSANs: []string{"foo.bar", "11.11.11.11"},
				},
				KubernetesVersion: "1.21",
			},
		}

		s.VersionsBundles = map[anywherev1.KubernetesVersion]*cluster.VersionsBundle{
			"1.21": {
				KubeDistro: &cluster.KubeDistro{
					Kubernetes: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/kubernetes",
						Tag:        "v1.21.5-eks-1-21-9",
					},
					CoreDNS: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/coredns",
						Tag:        "v1.8.4-eks-1-21-9",
					},
					Etcd: cluster.VersionedRepository{
						Repository: "public.ecr.aws/eks-distro/etcd-io",
						Tag:        "v3.4.16-eks-1-21-9",
					},
					EtcdImage: v1alpha1.Image{
						URI: "public.ecr.aws/eks-distro/etcd-io/etcd:0.0.1",
					},
					EtcdURL:     "https://distro.eks.amazonaws.com/kubernetes-1-19/releases/4/artifacts/etcd/v3.4.14/etcd-linux-amd64-v3.4.14.tar.gz",
					EtcdVersion: "3.4.14",
					Pause: v1alpha1.Image{
						URI: "public.ecr.aws/eks-distro/kubernetes/pause:0.0.1",
					},
				},
				VersionsBundle: &v1alpha1.VersionsBundle{
					BottleRocketHostContainers: v1alpha1.BottlerocketHostContainersBundle{
						Admin: v1alpha1.Image{
							URI: "public.ecr.aws/eks-anywhere/bottlerocket-admin:0.0.1",
						},
						Control: v1alpha1.Image{
							URI: "public.ecr.aws/eks-anywhere/bottlerocket-control:0.0.1",
						},
						KubeadmBootstrap: v1alpha1.Image{
							URI: "public.ecr.aws/eks-anywhere/bottlerocket-bootstrap:0.0.1",
						},
					},
				},
			},
		}
	})

	controlPlane := &controlplanev1beta2.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cp-test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta2",
		},
	}

	workerNodeGroupConfig := &anywherev1.WorkerNodeGroupConfiguration{
		Name:  "wng-1",
		Count: ptr.Int(3),
		Taints: []v1.Taint{
			{
				Key:       "key2",
				Value:     "val2",
				Effect:    v1.TaintEffectNoSchedule,
				TimeAdded: nil,
			},
		},
		Labels: map[string]string{
			"key3": "val3",
		},
	}

	kubeadmConfigTemplate := &bootstrapv1beta2.KubeadmConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "md-0",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmConfigTemplate",
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta2",
		},
	}

	providerCluster := &providerCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "provider-cluster",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProviderCluster",
			APIVersion: "providercluster.cluster.x-k8s.io/v1beta1",
		},
	}

	providerMachineTemplate := &providerMachineTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "provider-template",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "ProviderMachineTemplate",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
	}

	unstackedEtcdCluster := &unstackedEtcdCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: "unstacked-etcd-cluster",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "UnstackedEtcdCluster",
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
		},
	}

	return apiBuilerTest{
		WithT:                   NewWithT(t),
		clusterSpec:             clusterSpec,
		workerNodeGroupConfig:   workerNodeGroupConfig,
		kubeadmConfigTemplate:   kubeadmConfigTemplate,
		providerCluster:         providerCluster,
		controlPlane:            controlPlane,
		providerMachineTemplate: providerMachineTemplate,
		unstackedEtcdCluster:    unstackedEtcdCluster,
	}
}

func wantCluster() *clusterv1beta2.Cluster {
	return &clusterv1beta2.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta2",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name":                        "test-cluster",
				"cluster.anywhere.eks.amazonaws.com/cluster-name":      "test-cluster",
				"cluster.anywhere.eks.amazonaws.com/cluster-namespace": "my-namespace",
			},
		},
		Spec: clusterv1beta2.ClusterSpec{
			ClusterNetwork: clusterv1beta2.ClusterNetwork{
				Pods: clusterv1beta2.NetworkRanges{
					CIDRBlocks: []string{
						"1.2.3.4/5",
					},
				},
				Services: clusterv1beta2.NetworkRanges{
					CIDRBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			ControlPlaneRef: clusterv1beta2.ContractVersionedObjectReference{
				APIGroup: "controlplane.cluster.x-k8s.io",
				Kind:     "KubeadmControlPlane",
				Name:     "cp-test",
			},
			InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
				APIGroup: "providercluster.cluster.x-k8s.io",
				Kind:     "ProviderCluster",
				Name:     "provider-cluster",
			},
		},
	}
}

func TestCluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := clusterapi.Cluster(tt.clusterSpec, tt.providerCluster, tt.controlPlane, nil)
	want := wantCluster()
	tt.Expect(got).To(Equal(want))
}

type kubeadmControlPlaneOpt func(k *controlplanev1beta2.KubeadmControlPlane)

func wantKubeadmControlPlane(opts ...kubeadmControlPlaneOpt) *controlplanev1beta2.KubeadmControlPlane {
	replicas := int32(3)
	kcp := &controlplanev1beta2.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta2",
			Kind:       "KubeadmControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "eksa-system",
		},
		Spec: controlplanev1beta2.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1beta2.KubeadmControlPlaneMachineTemplate{
				Spec: controlplanev1beta2.KubeadmControlPlaneMachineTemplateSpec{
					InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
						APIGroup: "infrastructure.cluster.x-k8s.io",
						Kind:     "ProviderMachineTemplate",
						Name:     "provider-template",
					},
				},
			},
			KubeadmConfigSpec: bootstrapv1beta2.KubeadmConfigSpec{
				ClusterConfiguration: bootstrapv1beta2.ClusterConfiguration{
					ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
					DNS: bootstrapv1beta2.DNS{
						ImageRepository: "public.ecr.aws/eks-distro/coredns",
						ImageTag:        "v1.8.4-eks-1-21-9",
					},
					Etcd: bootstrapv1beta2.Etcd{
						Local: bootstrapv1beta2.LocalEtcd{
							ImageRepository: "public.ecr.aws/eks-distro/etcd-io",
							ImageTag:        "v3.4.16-eks-1-21-9",
							ExtraArgs: clusterapi.ExtraArgs{
								"cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							}.ToArgs(),
						},
					},
					APIServer: bootstrapv1beta2.APIServer{
						CertSANs: []string{"foo.bar", "11.11.11.11"},
					},
					ControllerManager: bootstrapv1beta2.ControllerManager{
						ExtraArgs:    clusterapi.SecureTlsCipherSuitesExtraArgs().ToArgs(),
						ExtraVolumes: []bootstrapv1beta2.HostPathMount{},
					},
					Scheduler: bootstrapv1beta2.Scheduler{},
				},
				InitConfiguration: bootstrapv1beta2.InitConfiguration{
					NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
						KubeletExtraArgs: clusterapi.SecureTlsCipherSuitesExtraArgs().
							Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(anywherev1.ControlPlaneConfiguration{
								Labels: map[string]string{"key1": "val1", "key2": "val2"},
							})).ToArgs(),
						Taints: &[]v1.Taint{
							{
								Key:       "key1",
								Value:     "val1",
								Effect:    v1.TaintEffectNoExecute,
								TimeAdded: nil,
							},
						},
					},
				},
				JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
					NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
						KubeletExtraArgs: clusterapi.SecureTlsCipherSuitesExtraArgs().
							Append(clusterapi.ControlPlaneNodeLabelsExtraArgs(anywherev1.ControlPlaneConfiguration{
								Labels: map[string]string{"key1": "val1", "key2": "val2"},
							})).ToArgs(),
						Taints: &[]v1.Taint{
							{
								Key:       "key1",
								Value:     "val1",
								Effect:    v1.TaintEffectNoExecute,
								TimeAdded: nil,
							},
						},
					},
				},
				PreKubeadmCommands:  []string{},
				PostKubeadmCommands: []string{},
				Files:               []bootstrapv1beta2.File{},
			},
			Replicas: &replicas,
			Version:  "v1.21.5-eks-1-21-9",
		},
	}

	for _, opt := range opts {
		opt(kcp)
	}

	return kcp
}

func TestKubeadmControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	got, err := clusterapi.KubeadmControlPlane(tt.clusterSpec, tt.providerMachineTemplate)
	tt.Expect(err).To(Succeed())
	want := wantKubeadmControlPlane()
	tt.Expect(got).To(Equal(want))
}

func wantKubeadmConfigTemplate() *bootstrapv1beta2.KubeadmConfigTemplate {
	return &bootstrapv1beta2.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta2",
			Kind:       "KubeadmConfigTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-wng-1-1",
			Namespace: "eksa-system",
		},
		Spec: bootstrapv1beta2.KubeadmConfigTemplateSpec{
			Template: bootstrapv1beta2.KubeadmConfigTemplateResource{
				Spec: bootstrapv1beta2.KubeadmConfigSpec{
					ClusterConfiguration: bootstrapv1beta2.ClusterConfiguration{
						ControllerManager: bootstrapv1beta2.ControllerManager{
							ExtraArgs: clusterapi.ExtraArgs{}.ToArgs(),
						},
						APIServer: bootstrapv1beta2.APIServer{
							ExtraArgs: clusterapi.ExtraArgs{}.ToArgs(),
							CertSANs:  []string{"foo.bar", "11.11.11.11"},
						},
					},
					JoinConfiguration: bootstrapv1beta2.JoinConfiguration{
						NodeRegistration: bootstrapv1beta2.NodeRegistrationOptions{
							KubeletExtraArgs: clusterapi.WorkerNodeLabelsExtraArgs(anywherev1.WorkerNodeGroupConfiguration{
								Labels: map[string]string{"key3": "val3"},
							}).ToArgs(),
							Taints: &[]v1.Taint{
								{
									Key:       "key2",
									Value:     "val2",
									Effect:    v1.TaintEffectNoSchedule,
									TimeAdded: nil,
								},
							},
						},
					},
					PreKubeadmCommands:  []string{},
					PostKubeadmCommands: []string{},
					Files:               []bootstrapv1beta2.File{},
				},
			},
		},
	}
}

func TestKubeadmConfigTemplate(t *testing.T) {
	tt := newApiBuilerTest(t)
	got, err := clusterapi.KubeadmConfigTemplate(tt.clusterSpec, *tt.workerNodeGroupConfig)
	tt.Expect(err).To(Succeed())
	want := wantKubeadmConfigTemplate()
	tt.Expect(got).To(Equal(want))
}

type machineDeploymentOpt func(m *clusterv1beta2.MachineDeployment)

func wantMachineDeployment(opts ...machineDeploymentOpt) *clusterv1beta2.MachineDeployment {
	replicas := int32(3)
	version := "v1.21.5-eks-1-21-9"
	md := &clusterv1beta2.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta2",
			Kind:       "MachineDeployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-wng-1",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name":                        "test-cluster",
				"cluster.anywhere.eks.amazonaws.com/cluster-name":      "test-cluster",
				"cluster.anywhere.eks.amazonaws.com/cluster-namespace": "my-namespace",
			},
			Annotations: map[string]string{},
		},
		Spec: clusterv1beta2.MachineDeploymentSpec{
			ClusterName: "test-cluster",
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1beta2.MachineTemplateSpec{
				ObjectMeta: clusterv1beta2.ObjectMeta{
					Labels: map[string]string{
						"cluster.x-k8s.io/cluster-name": "test-cluster",
					},
				},
				Spec: clusterv1beta2.MachineSpec{
					Bootstrap: clusterv1beta2.Bootstrap{
						ConfigRef: clusterv1beta2.ContractVersionedObjectReference{
							APIGroup: "bootstrap.cluster.x-k8s.io",
							Kind:     "KubeadmConfigTemplate",
							Name:     "md-0",
						},
					},
					ClusterName: "test-cluster",
					InfrastructureRef: clusterv1beta2.ContractVersionedObjectReference{
						APIGroup: "infrastructure.cluster.x-k8s.io",
						Kind:     "ProviderMachineTemplate",
						Name:     "provider-template",
					},
					Version: version,
				},
			},
			Replicas: &replicas,
		},
	}

	for _, opt := range opts {
		opt(md)
	}

	return md
}

func TestMachineDeployment(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := clusterapi.MachineDeployment(tt.clusterSpec, *tt.workerNodeGroupConfig, tt.kubeadmConfigTemplate, tt.providerMachineTemplate)
	tt.Expect(got).To(BeComparableTo(wantMachineDeployment()))
}

func TestClusterName(t *testing.T) {
	tests := []struct {
		name    string
		cluster *anywherev1.Cluster
		want    string
	}{
		{
			name:    "no name",
			cluster: &anywherev1.Cluster{},
			want:    "",
		},
		{
			name: "with name",
			cluster: &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "my-cluster",
				},
			},
			want: "my-cluster",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(clusterapi.ClusterName(tt.cluster)).To(Equal(tt.want))
		})
	}
}

func wantEtcdCluster() *etcdv1.EtcdadmCluster {
	replicas := int32(3)
	return &etcdv1.EtcdadmCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "etcdcluster.cluster.x-k8s.io/v1beta1",
			Kind:       "EtcdadmCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-etcd",
			Namespace: "eksa-system",
		},
		Spec: etcdv1.EtcdadmClusterSpec{
			Replicas: &replicas,
			EtcdadmConfigSpec: etcdbootstrapv1.EtcdadmConfigSpec{
				EtcdadmBuiltin:     true,
				CipherSuites:       "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
				PreEtcdadmCommands: []string{},
			},
			InfrastructureTemplate: v1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "ProviderMachineTemplate",
				Name:       "provider-template",
			},
		},
	}
}

func TestEtcdadmCluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	tt.clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
		Count: 3,
	}
	got := clusterapi.EtcdadmCluster(tt.clusterSpec, tt.providerMachineTemplate)
	want := wantEtcdCluster()
	tt.Expect(got).To(Equal(want))
}
