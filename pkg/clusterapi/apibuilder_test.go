package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

type apiBuilerTest struct {
	*WithT
	clusterSpec             *cluster.Spec
	workerNodeGroupConfig   *anywherev1.WorkerNodeGroupConfiguration
	kubeadmConfigTemplate   *bootstrapv1.KubeadmConfigTemplate
	providerCluster         clusterapi.APIObject
	controlPlane            clusterapi.APIObject
	providerMachineTemplate clusterapi.APIObject
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
				},
				KubernetesVersion: "1.21",
			},
		}
		s.VersionsBundle.KubeDistro.Kubernetes.Repository = "public.ecr.aws/eks-distro/kubernetes"
		s.VersionsBundle.KubeDistro.Kubernetes.Tag = "v1.21.5-eks-1-21-9"
		s.VersionsBundle.KubeDistro.CoreDNS.Repository = "public.ecr.aws/eks-distro/coredns"
		s.VersionsBundle.KubeDistro.CoreDNS.Tag = "v1.8.4-eks-1-21-9"
		s.VersionsBundle.KubeDistro.Etcd.Repository = "public.ecr.aws/eks-distro/etcd-io"
		s.VersionsBundle.KubeDistro.Etcd.Tag = "v3.4.16-eks-1-21-9"
	})

	controlPlane := &controlplanev1.KubeadmControlPlane{
		ObjectMeta: metav1.ObjectMeta{
			Name: "cp-test",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmControlPlane",
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
		},
	}

	workerNodeGroupConfig := &anywherev1.WorkerNodeGroupConfiguration{
		Name:  "wng-1",
		Count: 3,
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

	kubeadmConfigTemplate := &bootstrapv1.KubeadmConfigTemplate{
		ObjectMeta: metav1.ObjectMeta{
			Name: "md-0",
		},
		TypeMeta: metav1.TypeMeta{
			Kind:       "KubeadmConfigTemplate",
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
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

	return apiBuilerTest{
		WithT:                   NewWithT(t),
		clusterSpec:             clusterSpec,
		workerNodeGroupConfig:   workerNodeGroupConfig,
		kubeadmConfigTemplate:   kubeadmConfigTemplate,
		providerCluster:         providerCluster,
		controlPlane:            controlPlane,
		providerMachineTemplate: providerMachineTemplate,
	}
}

// TODO: add unstacked etcd test
func TestCluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := clusterapi.Cluster(tt.clusterSpec, tt.providerCluster, tt.controlPlane)
	want := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
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
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"1.2.3.4/5",
					},
				},
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"1.2.3.4/5",
					},
				},
			},
			ControlPlaneRef: &v1.ObjectReference{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmControlPlane",
				Name:       "cp-test",
			},
			InfrastructureRef: &v1.ObjectReference{
				APIVersion: "providercluster.cluster.x-k8s.io/v1beta1",
				Kind:       "ProviderCluster",
				Name:       "provider-cluster",
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func wantKubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	replicas := int32(3)
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster",
			Namespace: "eksa-system",
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					Kind:       "ProviderMachineTemplate",
					Name:       "provider-template",
				},
			},
			KubeadmConfigSpec: bootstrapv1.KubeadmConfigSpec{
				ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
					ImageRepository: "public.ecr.aws/eks-distro/kubernetes",
					DNS: bootstrapv1.DNS{
						ImageMeta: bootstrapv1.ImageMeta{
							ImageRepository: "public.ecr.aws/eks-distro/coredns",
							ImageTag:        "v1.8.4-eks-1-21-9",
						},
					},
					Etcd: bootstrapv1.Etcd{
						Local: &bootstrapv1.LocalEtcd{
							ImageMeta: bootstrapv1.ImageMeta{
								ImageRepository: "public.ecr.aws/eks-distro/etcd-io",
								ImageTag:        "v3.4.16-eks-1-21-9",
							},
							ExtraArgs: map[string]string{},
						},
					},
					APIServer: bootstrapv1.APIServer{
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs:    map[string]string{},
							ExtraVolumes: []bootstrapv1.HostPathMount{},
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: tlsCipherSuitesArgs(),
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							"node-labels":       "key1=val1,key2=val2",
						},
						Taints: []v1.Taint{
							{
								Key:       "key1",
								Value:     "val1",
								Effect:    v1.TaintEffectNoExecute,
								TimeAdded: nil,
							},
						},
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"tls-cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
							"node-labels":       "key1=val1,key2=val2",
						},
						Taints: []v1.Taint{
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
				Files:               []bootstrapv1.File{},
			},
			Replicas: &replicas,
			Version:  "v1.21.5-eks-1-21-9",
		},
	}
}

// TODO: add unstacked etcd test
func TestKubeadmControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	got, err := clusterapi.KubeadmControlPlane(tt.clusterSpec, tt.providerMachineTemplate)
	tt.Expect(err).To(Succeed())
	want := wantKubeadmControlPlane()
	tt.Expect(got).To(Equal(want))
}

func wantKubeadmConfigTemplate() *bootstrapv1.KubeadmConfigTemplate {
	return &bootstrapv1.KubeadmConfigTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmConfigTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cluster-wng-1-1",
			Namespace: "eksa-system",
		},
		Spec: bootstrapv1.KubeadmConfigTemplateSpec{
			Template: bootstrapv1.KubeadmConfigTemplateResource{
				Spec: bootstrapv1.KubeadmConfigSpec{
					ClusterConfiguration: &bootstrapv1.ClusterConfiguration{
						ControllerManager: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{},
						},
						APIServer: bootstrapv1.APIServer{
							ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
								ExtraArgs: map[string]string{},
							},
						},
					},
					JoinConfiguration: &bootstrapv1.JoinConfiguration{
						NodeRegistration: bootstrapv1.NodeRegistrationOptions{
							KubeletExtraArgs: map[string]string{
								"node-labels": "key3=val3",
							},
							Taints: []v1.Taint{
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
					Files:               []bootstrapv1.File{},
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

func TestMachineDeployment(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := clusterapi.MachineDeployment(tt.clusterSpec, *tt.workerNodeGroupConfig, tt.kubeadmConfigTemplate, tt.providerMachineTemplate)
	replicas := int32(3)
	version := "v1.21.5-eks-1-21-9"
	want := clusterv1.MachineDeployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
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
		},
		Spec: clusterv1.MachineDeploymentSpec{
			ClusterName: "test-cluster",
			Selector: metav1.LabelSelector{
				MatchLabels: map[string]string{},
			},
			Template: clusterv1.MachineTemplateSpec{
				ObjectMeta: clusterv1.ObjectMeta{
					Labels: map[string]string{
						"cluster.x-k8s.io/cluster-name": "test-cluster",
					},
				},
				Spec: clusterv1.MachineSpec{
					Bootstrap: clusterv1.Bootstrap{
						ConfigRef: &v1.ObjectReference{
							APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
							Kind:       "KubeadmConfigTemplate",
							Name:       "md-0",
						},
					},
					ClusterName: "test-cluster",
					InfrastructureRef: v1.ObjectReference{
						APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
						Kind:       "ProviderMachineTemplate",
						Name:       "provider-template",
					},
					Version: &version,
				},
			},
			Replicas: &replicas,
		},
	}
	tt.Expect(got).To(Equal(want))
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
