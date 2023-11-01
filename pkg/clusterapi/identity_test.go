package clusterapi_test

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
)

func TestConfigureAWSIAMAuthInKubeadmControlPlane(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name         string
		awsIamConfig *v1alpha1.AWSIamConfig
		want         *controlplanev1.KubeadmControlPlane
	}{
		{
			name:         "no iam auth",
			awsIamConfig: nil,
			want:         wantKubeadmControlPlane(),
		},
		{
			name: "with iam auth",
			awsIamConfig: &v1alpha1.AWSIamConfig{
				Spec: v1alpha1.AWSIamConfigSpec{
					AWSRegion:   "test-region",
					BackendMode: []string{"mode1", "mode2"},
					MapRoles: []v1alpha1.MapRoles{
						{
							RoleARN:  "test-role-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					MapUsers: []v1alpha1.MapUsers{
						{
							UserARN:  "test-user-arn",
							Username: "test",
							Groups:   []string{"group1", "group2"},
						},
					},
					Partition: "aws",
				},
			},
			want: &controlplanev1.KubeadmControlPlane{
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
									ExtraArgs: map[string]string{
										"cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
									},
								},
							},
							APIServer: bootstrapv1.APIServer{
								ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
									ExtraArgs: map[string]string{
										"authentication-token-webhook-config-file": "/etc/kubernetes/aws-iam-authenticator/kubeconfig.yaml",
									},
									ExtraVolumes: []bootstrapv1.HostPathMount{
										{
											Name:      "authconfig",
											HostPath:  "/var/lib/kubeadm/aws-iam-authenticator/",
											MountPath: "/etc/kubernetes/aws-iam-authenticator/",
											ReadOnly:  false,
										},
										{
											Name:      "awsiamcert",
											HostPath:  "/var/lib/kubeadm/aws-iam-authenticator/pki/",
											MountPath: "/var/aws-iam-authenticator/",
											ReadOnly:  false,
										},
									},
								},
								CertSANs: []string{"foo.bar", "11.11.11.11"},
							},
							ControllerManager: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    tlsCipherSuitesArgs(),
								ExtraVolumes: []bootstrapv1.HostPathMount{},
							},
							Scheduler: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    map[string]string{},
								ExtraVolumes: []bootstrapv1.HostPathMount{},
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
						Files: []bootstrapv1.File{
							{
								Path:        "/var/lib/kubeadm/aws-iam-authenticator/kubeconfig.yaml",
								Owner:       "root:root",
								Permissions: "0640",
								Content: `
# clusters refers to the remote service.
clusters:
  - name: aws-iam-authenticator
    cluster:
      certificate-authority: /var/aws-iam-authenticator/cert.pem
      server: https://localhost:21362/authenticate
# users refers to the API Server's webhook configuration
# (we don't need to authenticate the API server).
users:
  - name: apiserver
# kubeconfig files require a context. Provide one for the API Server.
current-context: webhook
contexts:
- name: webhook
  context:
    cluster: aws-iam-authenticator
    user: apiserver
`,
							},
							{
								Path:        "/var/lib/kubeadm/aws-iam-authenticator/pki/cert.pem",
								Owner:       "root:root",
								Permissions: "0640",
								ContentFrom: &bootstrapv1.FileSource{
									Secret: bootstrapv1.SecretFileSource{
										Name: "test-cluster-aws-iam-authenticator-ca",
										Key:  "cert.pem",
									},
								},
							},
							{
								Path:        "/var/lib/kubeadm/aws-iam-authenticator/pki/key.pem",
								Owner:       "root:root",
								Permissions: "0640",
								ContentFrom: &bootstrapv1.FileSource{
									Secret: bootstrapv1.SecretFileSource{
										Name: "test-cluster-aws-iam-authenticator-ca",
										Key:  "key.pem",
									},
								},
							},
						},
					},
					Replicas: &replicas,
					Version:  "v1.21.5-eks-1-21-9",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			g.clusterSpec.AWSIamConfig = tt.awsIamConfig
			clusterapi.SetIdentityAuthInKubeadmControlPlane(got, g.clusterSpec)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestConfigureOIDCInKubeadmControlPlane(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name       string
		oidcConfig *v1alpha1.OIDCConfig
		want       *controlplanev1.KubeadmControlPlane
	}{
		{
			name:       "no oidc",
			oidcConfig: nil,
			want:       wantKubeadmControlPlane(),
		},
		{
			name: "with oidc",
			oidcConfig: &v1alpha1.OIDCConfig{
				TypeMeta: metav1.TypeMeta{
					Kind:       "OIDCConfig",
					APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "eksa-unit-test",
				},
				Spec: v1alpha1.OIDCConfigSpec{
					ClientId:     "id1",
					GroupsClaim:  "claim1",
					GroupsPrefix: "prefix-for-groups",
					IssuerUrl:    "https://mydomain.com/issuer",
					RequiredClaims: []v1alpha1.OIDCConfigRequiredClaim{
						{
							Claim: "sub",
							Value: "test",
						},
					},
					UsernameClaim:  "username-claim",
					UsernamePrefix: "username-prefix",
				},
			},
			want: &controlplanev1.KubeadmControlPlane{
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
									ExtraArgs: map[string]string{
										"cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
									},
								},
							},
							APIServer: bootstrapv1.APIServer{
								ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
									ExtraArgs: map[string]string{
										"oidc-client-id":       "id1",
										"oidc-groups-claim":    "claim1",
										"oidc-groups-prefix":   "prefix-for-groups",
										"oidc-issuer-url":      "https://mydomain.com/issuer",
										"oidc-required-claim":  "sub=test",
										"oidc-username-claim":  "username-claim",
										"oidc-username-prefix": "username-prefix",
									},
									ExtraVolumes: []bootstrapv1.HostPathMount{},
								},
								CertSANs: []string{"foo.bar", "11.11.11.11"},
							},
							ControllerManager: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    tlsCipherSuitesArgs(),
								ExtraVolumes: []bootstrapv1.HostPathMount{},
							},
							Scheduler: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    map[string]string{},
								ExtraVolumes: []bootstrapv1.HostPathMount{},
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			g.clusterSpec.OIDCConfig = tt.oidcConfig
			clusterapi.SetIdentityAuthInKubeadmControlPlane(got, g.clusterSpec)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestConfigurePodIamAuthInKubeadmControlPlane(t *testing.T) {
	replicas := int32(3)
	tests := []struct {
		name         string
		podIAMConfig *v1alpha1.PodIAMConfig
		want         *controlplanev1.KubeadmControlPlane
	}{
		{
			name:         "no pod iam",
			podIAMConfig: nil,
			want:         wantKubeadmControlPlane(),
		},
		{
			name: "with pod iam",
			podIAMConfig: &v1alpha1.PodIAMConfig{
				ServiceAccountIssuer: "https://test",
			},
			want: &controlplanev1.KubeadmControlPlane{
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
									ExtraArgs: map[string]string{
										"cipher-suites": "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
									},
								},
							},
							APIServer: bootstrapv1.APIServer{
								ControlPlaneComponent: bootstrapv1.ControlPlaneComponent{
									ExtraArgs: map[string]string{
										"service-account-issuer": "https://test",
									},
									ExtraVolumes: []bootstrapv1.HostPathMount{},
								},
								CertSANs: []string{"foo.bar", "11.11.11.11"},
							},
							ControllerManager: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    tlsCipherSuitesArgs(),
								ExtraVolumes: []bootstrapv1.HostPathMount{},
							},
							Scheduler: bootstrapv1.ControlPlaneComponent{
								ExtraArgs:    map[string]string{},
								ExtraVolumes: []bootstrapv1.HostPathMount{},
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
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := newApiBuilerTest(t)
			got := wantKubeadmControlPlane()
			g.clusterSpec.Cluster.Spec.PodIAMConfig = tt.podIAMConfig
			clusterapi.SetIdentityAuthInKubeadmControlPlane(got, g.clusterSpec)
			g.Expect(got).To(Equal(tt.want))
		})
	}
}
