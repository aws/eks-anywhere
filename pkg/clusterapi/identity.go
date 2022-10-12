package clusterapi

import (
	bootstrapv1 "sigs.k8s.io/cluster-api/bootstrap/kubeadm/api/v1beta1"
	controlplanev1 "sigs.k8s.io/cluster-api/controlplane/kubeadm/api/v1beta1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
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

var awsIamMounts = []bootstrapv1.HostPathMount{
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
}

var awsIamFiles = []bootstrapv1.File{
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
}

func configureAWSIAMAuthInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, awsIamConfig *v1alpha1.AWSIamConfig) {
	if awsIamConfig == nil {
		return
	}

	apiServerExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs
	for k, v := range AwsIamAuthExtraArgs(awsIamConfig) {
		apiServerExtraArgs[k] = v
	}

	kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes = append(
		kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraVolumes,
		awsIamMounts...,
	)

	kcp.Spec.KubeadmConfigSpec.Files = append(kcp.Spec.KubeadmConfigSpec.Files, awsIamFiles...)
}

func configureOIDCInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, oidcConfig *v1alpha1.OIDCConfig) {
	if oidcConfig == nil {
		return
	}

	apiServerExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs
	for k, v := range OIDCToExtraArgs(oidcConfig) {
		apiServerExtraArgs[k] = v
	}
}

func configurePodIamAuthInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, podIamConfig *v1alpha1.PodIAMConfig) {
	if podIamConfig == nil {
		return
	}

	apiServerExtraArgs := kcp.Spec.KubeadmConfigSpec.ClusterConfiguration.APIServer.ExtraArgs
	for k, v := range PodIAMAuthExtraArgs(podIamConfig) {
		apiServerExtraArgs[k] = v
	}
}

func SetIdentityAuthInKubeadmControlPlane(kcp *controlplanev1.KubeadmControlPlane, clusterSpec *cluster.Spec) {
	configureOIDCInKubeadmControlPlane(kcp, clusterSpec.OIDCConfig)
	configureAWSIAMAuthInKubeadmControlPlane(kcp, clusterSpec.AWSIamConfig)
	configurePodIamAuthInKubeadmControlPlane(kcp, clusterSpec.Cluster.Spec.PodIAMConfig)
}
