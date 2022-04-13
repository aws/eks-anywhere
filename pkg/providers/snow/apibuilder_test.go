package snow

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	snowv1 "github.com/aws/eks-anywhere/pkg/providers/snow/api/v1beta1"
)

type apiBuilerTest struct {
	*WithT
	clusterSpec    *cluster.Spec
	machineConfigs map[string]*v1alpha1.SnowMachineConfig
}

func newApiBuilerTest(t *testing.T) apiBuilerTest {
	return apiBuilerTest{
		WithT:          NewWithT(t),
		clusterSpec:    givenClusterSpec(),
		machineConfigs: givenMachineConfigs(),
	}
}

func TestCAPICluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	snowCluster := SnowCluster(tt.clusterSpec)
	controlPlaneMachineTemplate := SnowMachineTemplate(tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	kubeadmControlPlane, err := KubeadmControlPlane(tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	got := CAPICluster(tt.clusterSpec, snowCluster, kubeadmControlPlane)
	want := &clusterv1.Cluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "cluster.x-k8s.io/v1beta1",
			Kind:       "Cluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
			Labels: map[string]string{
				"cluster.x-k8s.io/cluster-name": "snow-test",
			},
		},
		Spec: clusterv1.ClusterSpec{
			ClusterNetwork: &clusterv1.ClusterNetwork{
				Pods: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"10.1.0.0/16",
					},
				},
				Services: &clusterv1.NetworkRanges{
					CIDRBlocks: []string{
						"10.96.0.0/12",
					},
				},
			},
			ControlPlaneRef: &v1.ObjectReference{
				APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmControlPlane",
				Name:       "snow-test",
			},
			InfrastructureRef: &v1.ObjectReference{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "AWSSnowCluster",
				Name:       "snow-test",
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func wantKubeadmControlPlane() *controlplanev1.KubeadmControlPlane {
	wantReplicas := int32(3)
	return &controlplanev1.KubeadmControlPlane{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "controlplane.cluster.x-k8s.io/v1beta1",
			Kind:       "KubeadmControlPlane",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
		},
		Spec: controlplanev1.KubeadmControlPlaneSpec{
			MachineTemplate: controlplanev1.KubeadmControlPlaneMachineTemplate{
				InfrastructureRef: v1.ObjectReference{
					APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
					Kind:       "AWSSnowMachineTemplate",
					Name:       "test-cp",
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
							ExtraArgs: map[string]string{
								"listen-peer-urls":   "https://0.0.0.0:2380",
								"listen-client-urls": "https://0.0.0.0:2379",
							},
						},
					},
					APIServer: bootstrapv1.APIServer{
						ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
							ExtraArgs: map[string]string{},
						},
					},
					ControllerManager: bootstrapv1.ControlPlaneComponent{
						ExtraArgs: map[string]string{},
					},
				},
				InitConfiguration: &bootstrapv1.InitConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"provider-id": "aws-snow:////'{{ ds.meta_data.instance_id }}'",
						},
					},
				},
				JoinConfiguration: &bootstrapv1.JoinConfiguration{
					NodeRegistration: bootstrapv1.NodeRegistrationOptions{
						KubeletExtraArgs: map[string]string{
							"provider-id": "aws-snow:////'{{ ds.meta_data.instance_id }}'",
						},
					},
				},
				PreKubeadmCommands: []string{
					"/etc/eks/bootstrap.sh public.ecr.aws/l0g8r8j6/plunder-app/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433 1.2.3.4",
				},
				PostKubeadmCommands: []string{
					"/etc/eks/bootstrap-after.sh public.ecr.aws/l0g8r8j6/plunder-app/kube-vip:v0.3.7-eks-a-v0.0.0-dev-build.1433 1.2.3.4",
				},
				Files: []bootstrapv1.File{},
			},
			Replicas: &wantReplicas,
			Version:  "v1.21.5-eks-1-21-9",
		},
	}
}

func wantRegistryMirrorCommands() []string {
	return []string{
		"cat /etc/containerd/config_append.toml >> /etc/containerd/config.toml",
		"sudo systemctl daemon-reload",
		"sudo systemctl restart containerd",
	}
}

func TestKubeadmControlPlane(t *testing.T) {
	tt := newApiBuilerTest(t)
	controlPlaneMachineTemplate := SnowMachineTemplate(tt.machineConfigs[tt.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
	got, err := KubeadmControlPlane(tt.clusterSpec, controlPlaneMachineTemplate)
	tt.Expect(err).To(Succeed())

	want := wantKubeadmControlPlane()
	tt.Expect(got).To(Equal(want))
}

func TestKubeadmControlPlaneWithRegistryMirror(t *testing.T) {
	tests := []struct {
		name                 string
		registryMirrorConfig *v1alpha1.RegistryMirrorConfiguration
		wantFiles            []bootstrapv1.File
	}{
		{
			name: "with ca cert",
			registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint:      "1.2.3.4",
				Port:          "443",
				CACertContent: "xyz",
			},
			wantFiles: []bootstrapv1.File{
				{
					Path:  "/etc/containerd/config_append.toml",
					Owner: "root:root",
					Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]
  [plugins."io.containerd.grpc.v1.cri".registry.configs."1.2.3.4:443".tls]
    ca_file = "/etc/containerd/certs.d/1.2.3.4:443/ca.crt"`,
				},
				{
					Path:    "/etc/containerd/certs.d/1.2.3.4:443/ca.crt",
					Owner:   "root:root",
					Content: "xyz",
				},
			},
		},
		{
			name: "without ca cert",
			registryMirrorConfig: &v1alpha1.RegistryMirrorConfiguration{
				Endpoint: "1.2.3.4",
				Port:     "443",
			},
			wantFiles: []bootstrapv1.File{
				{
					Path:  "/etc/containerd/config_append.toml",
					Owner: "root:root",
					Content: `[plugins."io.containerd.grpc.v1.cri".registry.mirrors]
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."public.ecr.aws"]
    endpoint = ["https://1.2.3.4:443"]`,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			g.clusterSpec.Cluster.Spec.RegistryMirrorConfiguration = tt.registryMirrorConfig
			controlPlaneMachineTemplate := SnowMachineTemplate(g.machineConfigs[g.clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name])
			got, err := KubeadmControlPlane(g.clusterSpec, controlPlaneMachineTemplate)
			g.Expect(err).To(Succeed())
			want := wantKubeadmControlPlane()
			want.Spec.KubeadmConfigSpec.Files = tt.wantFiles
			want.Spec.KubeadmConfigSpec.PreKubeadmCommands = append(wantRegistryMirrorCommands(), want.Spec.KubeadmConfigSpec.PreKubeadmCommands...)
			g.Expect(got).To(Equal(want))
		})
	}
}

func TestKubeadmConfigTemplates(t *testing.T) {
	tt := newApiBuilerTest(t)
	got, err := KubeadmConfigTemplates(tt.clusterSpec)
	tt.Expect(err).To(Succeed())

	want := map[string]*bootstrapv1.KubeadmConfigTemplate{
		"md-0": {
			TypeMeta: metav1.TypeMeta{
				APIVersion: "bootstrap.cluster.x-k8s.io/v1beta1",
				Kind:       "KubeadmConfigTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "md-0",
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
									"provider-id": "aws-snow:////'{{ ds.meta_data.instance_id }}'",
								},
							},
						},
						PreKubeadmCommands: []string{
							"/etc/eks/bootstrap.sh",
						},
						PostKubeadmCommands: []string{},
						Files:               []bootstrapv1.File{},
					},
				},
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func TestMachineDeployments(t *testing.T) {
	tt := newApiBuilerTest(t)
	kubeadmConfigTemplates, err := KubeadmConfigTemplates(tt.clusterSpec)
	tt.Expect(err).To(Succeed())

	workerMachineTemplates := SnowMachineTemplates(tt.clusterSpec, tt.machineConfigs)
	got := MachineDeployments(tt.clusterSpec, kubeadmConfigTemplates, workerMachineTemplates)
	wantVersion := "v1.21.5-eks-1-21-9"
	wantReplicas := int32(3)
	want := map[string]*clusterv1.MachineDeployment{
		"md-0": {
			TypeMeta: metav1.TypeMeta{
				APIVersion: "cluster.x-k8s.io/v1beta1",
				Kind:       "MachineDeployment",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "md-0",
				Namespace: "eksa-system",
				Labels: map[string]string{
					"cluster.x-k8s.io/cluster-name": "snow-test",
				},
			},
			Spec: clusterv1.MachineDeploymentSpec{
				ClusterName: "snow-test",
				Selector: metav1.LabelSelector{
					MatchLabels: map[string]string{},
				},
				Template: clusterv1.MachineTemplateSpec{
					ObjectMeta: clusterv1.ObjectMeta{
						Labels: map[string]string{
							"cluster.x-k8s.io/cluster-name": "snow-test",
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
						ClusterName: "snow-test",
						InfrastructureRef: v1.ObjectReference{
							APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
							Kind:       "AWSSnowMachineTemplate",
							Name:       "test-wn",
						},
						Version: &wantVersion,
					},
				},
				Replicas: &wantReplicas,
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func TestSnowCluster(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := SnowCluster(tt.clusterSpec)
	want := &snowv1.AWSSnowCluster{
		TypeMeta: metav1.TypeMeta{
			Kind:       "AWSSnowCluster",
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "snow-test",
			Namespace: "eksa-system",
		},
		Spec: snowv1.AWSSnowClusterSpec{
			Region: "snow",
			ControlPlaneEndpoint: clusterv1.APIEndpoint{
				Host: "1.2.3.4",
				Port: 6443,
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func TestSnowMachineTemplate(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := SnowMachineTemplate(tt.machineConfigs["test-wn"])
	wantAMIID := "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea"
	wantSSHKey := "default"
	wantPhysicalNetworkConnector := "SFP_PLUS"
	want := &snowv1.AWSSnowMachineTemplate{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
			Kind:       "AWSSnowMachineTemplate",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-wn",
			Namespace: "eksa-system",
		},
		Spec: snowv1.AWSSnowMachineTemplateSpec{
			Template: snowv1.AWSSnowMachineTemplateResource{
				Spec: snowv1.AWSSnowMachineSpec{
					IAMInstanceProfile: "control-plane.cluster-api-provider-aws.sigs.k8s.io",
					InstanceType:       "sbe-c.xlarge",
					SSHKeyName:         &wantSSHKey,
					AMI: snowv1.AWSResourceReference{
						ID: &wantAMIID,
					},
					CloudInit: snowv1.CloudInit{
						InsecureSkipSecretsManager: true,
					},
					PhysicalNetworkConnectorType: &wantPhysicalNetworkConnector,
				},
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}

func TestSnowMachineTemplates(t *testing.T) {
	tt := newApiBuilerTest(t)
	got := SnowMachineTemplates(tt.clusterSpec, tt.machineConfigs)
	wantAMIID := "eks-d-v1-21-5-ubuntu-ami-02833ca9a8f29c2ea"
	wantSSHKey := "default"
	wantPhysicalNetworkConnector := "SFP_PLUS"
	want := map[string]*snowv1.AWSSnowMachineTemplate{
		"test-wn": {
			TypeMeta: metav1.TypeMeta{
				APIVersion: "infrastructure.cluster.x-k8s.io/v1beta1",
				Kind:       "AWSSnowMachineTemplate",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-wn",
				Namespace: "eksa-system",
			},
			Spec: snowv1.AWSSnowMachineTemplateSpec{
				Template: snowv1.AWSSnowMachineTemplateResource{
					Spec: snowv1.AWSSnowMachineSpec{
						IAMInstanceProfile: "control-plane.cluster-api-provider-aws.sigs.k8s.io",
						InstanceType:       "sbe-c.xlarge",
						SSHKeyName:         &wantSSHKey,
						AMI: snowv1.AWSResourceReference{
							ID: &wantAMIID,
						},
						CloudInit: snowv1.CloudInit{
							InsecureSkipSecretsManager: true,
						},
						PhysicalNetworkConnectorType: &wantPhysicalNetworkConnector,
					},
				},
			},
		},
	}
	tt.Expect(got).To(Equal(want))
}
