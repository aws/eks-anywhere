package clusterapi

import (
	bootstrapv1beta2 "sigs.k8s.io/cluster-api/api/bootstrap/kubeadm/v1beta2"
	controlplanev1beta2 "sigs.k8s.io/cluster-api/api/controlplane/kubeadm/v1beta2"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

const awsIamKubeconfig = `
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
`

var awsIamMounts = []bootstrapv1beta2.HostPathMount{
	{
		Name:      "authconfig",
		HostPath:  "/var/lib/kubeadm/aws-iam-authenticator/",
		MountPath: "/etc/kubernetes/aws-iam-authenticator/",
		ReadOnly:  ptr.Bool(false),
	},
	{
		Name:      "awsiamcert",
		HostPath:  "/var/lib/kubeadm/aws-iam-authenticator/pki/",
		MountPath: "/var/aws-iam-authenticator/",
		ReadOnly:  ptr.Bool(false),
	},
}

var awsIamFiles = []bootstrapv1beta2.File{
	{
		Path:        "/var/lib/kubeadm/aws-iam-authenticator/kubeconfig.yaml",
		Owner:       "root:root",
		Permissions: "0640",
		Content:     awsIamKubeconfig,
	},
	{
		Path:        "/var/lib/kubeadm/aws-iam-authenticator/pki/cert.pem",
		Owner:       "root:root",
		Permissions: "0640",
		ContentFrom: bootstrapv1beta2.FileSource{
			Secret: bootstrapv1beta2.SecretFileSource{
				Name: "test-cluster-aws-iam-authenticator-ca",
				Key:  "cert.pem",
			},
		},
	},
	{
		Path:        "/var/lib/kubeadm/aws-iam-authenticator/pki/key.pem",
		Owner:       "root:root",
		Permissions: "0640",
		ContentFrom: bootstrapv1beta2.FileSource{
			Secret: bootstrapv1beta2.SecretFileSource{
				Name: "test-cluster-aws-iam-authenticator-ca",
				Key:  "key.pem",
			},
		},
	},
}

func configureAWSIAMAuthInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, awsIamConfig *v1alpha1.AWSIamConfig) {
	if awsIamConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs,
		AwsIamAuthExtraArgs(awsIamConfig).ToArgs()...,
	)

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes,
		awsIamMounts...,
	)

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, awsIamFiles...)
}

func configureOIDCInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, oidcConfig *v1alpha1.OIDCConfig) {
	if oidcConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs,
		OIDCToExtraArgs(oidcConfig).ToArgs()...,
	)
}

func configureAPIServerExtraArgsInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, apiServerExtraArgs map[string]string) {
	if apiServerExtraArgs == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs,
		ExtraArgs(apiServerExtraArgs).ToArgs()...,
	)
}

func configurePodIamAuthInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, podIamConfig *v1alpha1.PodIAMConfig) {
	if podIamConfig == nil {
		return
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs = SetPodIAMAuthInArgs(
		podIamConfig, kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs)
}

func SetIdentityAuthInKubeadmControlPlane(kcp *controlplanev1beta2.KubeadmControlPlane, clusterSpec *cluster.Spec) {
	configureOIDCInKubeadmControlPlane(kcp, clusterSpec.OIDCConfig)
	configureAWSIAMAuthInKubeadmControlPlane(kcp, clusterSpec.AWSIamConfig)
	configureAPIServerExtraArgsInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.ControlPlaneConfiguration.APIServerExtraArgs)
	configurePodIamAuthInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.PodIAMConfig)
}
