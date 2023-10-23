package framework

import (
	"context"
	"crypto"
	"crypto/rsa"
	"crypto/x509"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	jose "github.com/go-jose/go-jose/v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/templater"
)

const (
	irsaS3BucketVar = "T_IRSA_S3_BUCKET"
	kmsIAMRoleVar   = "T_KMS_IAM_ROLE"
	kmsImageVar     = "T_KMS_IMAGE"
	kmsKeyArn       = "T_KMS_KEY_ARN"
	kmsKeyRegion    = "T_KMS_KEY_REGION"
	kmsSocketVar    = "T_KMS_SOCKET"

	defaultRegion = "us-west-2"
	keysFilename  = "keys.json"

	// SSHKeyPath is the path where the SSH private key is stored on the test-runner instance.
	SSHKeyPath = "/tmp/ssh_key"
)

//go:embed config/pod-identity-webhook.yaml
var podIdentityWebhookManifest []byte

//go:embed config/aws-kms-encryption-provider.yaml
var kmsProviderManifest string

type keyResponse struct {
	Keys []jose.JSONWebKey `json:"keys"`
}

// etcdEncryptionTestVars stores all the environment variables needed by etcd encryption tests.
type etcdEncryptionTestVars struct {
	KmsKeyRegion string
	S3Bucket     string
	KmsIamRole   string
	KmsImage     string
	KmsKeyArn    string
	KmsSocket    string
}

// RequiredEtcdEncryptionEnvVars returns the environment variables required .
func RequiredEtcdEncryptionEnvVars() []string {
	return []string{irsaS3BucketVar, kmsIAMRoleVar, kmsImageVar, kmsKeyArn, kmsSocketVar}
}

func getEtcdEncryptionVarsFromEnv() *etcdEncryptionTestVars {
	return &etcdEncryptionTestVars{
		KmsKeyRegion: os.Getenv(kmsKeyRegion),
		S3Bucket:     os.Getenv(irsaS3BucketVar),
		KmsIamRole:   os.Getenv(kmsIAMRoleVar),
		KmsImage:     os.Getenv(kmsImageVar),
		KmsKeyArn:    os.Getenv(kmsKeyArn),
		KmsSocket:    os.Getenv(kmsSocketVar),
	}
}

// WithPodIamConfig is a ClusterE2ETestOpt that adds pod IAM config to the cluster.
func WithPodIamConfig() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, RequiredEtcdEncryptionEnvVars())
		e.clusterFillers = append(e.clusterFillers, api.WithPodIamFiller(getIssuerURL()))
	}
}

// WithEtcdEncrytion is a ClusterE2ETestOpt that adds etcd encryption config to the Cluster.
func WithEtcdEncrytion() ClusterE2ETestOpt {
	return func(e *ClusterE2ETest) {
		checkRequiredEnvVars(e.T, RequiredEtcdEncryptionEnvVars())
		encryptionVars := getEtcdEncryptionVarsFromEnv()
		kms := &v1alpha1.KMS{
			CacheSize:           v1alpha1.DefaultKMSCacheSize,
			Name:                "test-kms-config",
			SocketListenAddress: fmt.Sprintf("unix://%s", encryptionVars.KmsSocket),
			Timeout:             &v1alpha1.DefaultKMSTimeout,
		}
		e.UpdateClusterConfig(api.ClusterToConfigFiller(api.WithEtcdEncryptionFiller(kms, []string{"secrets"})))
	}
}

// ValidateEtcdEncryption validates that etcd encryption is properly configured by creating a secret
// and SSHing into the ETCD nodes and ensuring the secret is not stored in plaintext.
func (e *ClusterE2ETest) ValidateEtcdEncryption() {
	ctx := context.Background()
	secretName := "my-very-secure-secret"
	secretVal := "confidential"
	secret := &v1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "default",
		},
		StringData: map[string]string{
			"data": secretName,
		},
	}
	if err := e.KubectlClient.Apply(ctx, e.KubeconfigFilePath(), secret); err != nil {
		e.T.Fatalf("Error creating secret to validate etcd encryption: %v", err)
	}

	machines, err := e.KubectlClient.GetCAPIMachines(ctx, e.Cluster(), e.ClusterName)
	if err != nil {
		e.T.Fatalf("Error getting CAPI machines to validate etcd encryption: %v", err)
	}

	etcdIPs := make([]string, 0)
	for _, m := range machines {
		if _, exists := m.Labels["cluster.x-k8s.io/etcd-cluster"]; exists && len(m.Status.Addresses) > 0 {
			etcdIPs = append(etcdIPs, m.Status.Addresses[0].Address)
		}
	}

	ssh := buildSSH(e.T)
	cmd := []string{
		"sudo", "ETCDCTL_API=3", "etcdctl",
		"--cacert=/etc/etcd/pki/ca.crt",
		"--cert=/etc/etcd/pki/etcdctl-etcd-client.crt",
		"--key=/etc/etcd/pki/etcdctl-etcd-client.key",
		"get", fmt.Sprintf("/registry/secrets/default/%s", secretName), "| hexdump -C",
	}
	for _, etcdIP := range etcdIPs {
		out, err := ssh.RunCommand(ctx, SSHKeyPath, getSSHUsernameByProvider(e.Provider.Name()), etcdIP, cmd...)
		if err != nil {
			e.T.Fatalf("Error verifying the secret is encrypted in etcd: %v", err)
		}
		e.T.Log(out)
		if strings.Contains(out, secretVal) && !strings.Contains(out, "k8s:enc:kms:v1") {
			e.T.Fatal("The secure secret is not encrypted")
		}
		e.T.Logf("The secret is encrypted with KMS on etcd node %s", etcdIP)
	}
}

func getSSHUsernameByProvider(provider string) string {
	switch provider {
	case "cloudstack":
		return "capc"
	case "nutanix":
		return "eksa"
	default:
		return "ec2-user"
	}
}

// PostClusterCreateEtcdEncryptionSetup performs operations on the cluster to prepare it for etcd encryption.
// These operations include:
// - Adding Cluster SA cert to the OIDC provider's keys.
// - Deploying Pod Identity Webhook.
// - Deploying AWS KMS Provider.
func (e *ClusterE2ETest) PostClusterCreateEtcdEncryptionSetup() {
	ctx := context.Background()
	envVars := getEtcdEncryptionVarsFromEnv()
	awsSession, err := session.NewSession(&aws.Config{
		Region: aws.String(defaultRegion),
	})
	if err != nil {
		e.T.Fatalf("creating aws session for tests: %v", err)
	}

	if err := e.addClusterCertToIrsaOidcProvider(ctx, envVars, awsSession); err != nil {
		e.T.Fatal(err)
	}

	if err := e.deployPodIdentityWebhook(ctx, envVars); err != nil {
		e.T.Fatal(err)
	}

	if err := e.deployKMSProvider(ctx, envVars); err != nil {
		e.T.Fatal(err)
	}
}

func getIssuerURL() string {
	etcdEncryptionConfig := getEtcdEncryptionVarsFromEnv()
	return fmt.Sprintf("https://s3.%s.amazonaws.com/%s", defaultRegion, etcdEncryptionConfig.S3Bucket)
}

func (e *ClusterE2ETest) deployPodIdentityWebhook(ctx context.Context, envVars *etcdEncryptionTestVars) error {
	e.T.Log("Deploying Pod Identity Webhook")
	if err := e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), podIdentityWebhookManifest); err != nil {
		return fmt.Errorf("deploying pod identity webhook: %v", err)
	}
	return nil
}

func (e *ClusterE2ETest) deployKMSProvider(ctx context.Context, envVars *etcdEncryptionTestVars) error {
	e.T.Log("Deploying AWS KMS Encryption Provider")
	values := map[string]string{
		"kmsImage":           envVars.KmsImage,
		"kmsIamRole":         envVars.KmsIamRole,
		"kmsKeyArn":          envVars.KmsKeyArn,
		"kmsKeyRegion":       envVars.KmsKeyRegion,
		"kmsSocket":          envVars.KmsSocket,
		"serviceAccountName": "kms-encrypter-decrypter",
	}

	if e.OSFamily != v1alpha1.Bottlerocket {
		values["deployOnlyOnControlPlane"] = "true"
	}

	manifest, err := templater.Execute(kmsProviderManifest, values)
	if err != nil {
		return fmt.Errorf("templating kms provider manifest: %v", err)
	}

	if err = e.KubectlClient.ApplyKubeSpecFromBytes(ctx, e.Cluster(), manifest); err != nil {
		return fmt.Errorf("deploying kms provider: %v", err)
	}
	return nil
}

func (e *ClusterE2ETest) addClusterCertToIrsaOidcProvider(ctx context.Context, envVars *etcdEncryptionTestVars, awsSession *session.Session) error {
	e.T.Log("Adding Cluster Cert to OIDC keys.json on S3")
	// Fetch the cluster's service account cert
	saCert, err := e.getClusterSACert(ctx)
	if err != nil {
		return err
	}

	// download the current keys json from S3 to add the current cluster's cert
	content, err := s3.Download(awsSession, keysFilename, envVars.S3Bucket)
	if err != nil {
		return fmt.Errorf("downloading %s from s3: %v", keysFilename, err)
	}

	resp := &keyResponse{}
	if err = json.Unmarshal(content, resp); err != nil {
		return fmt.Errorf("unmarshaling %s into json: %v", keysFilename, err)
	}

	newKey, err := getJSONWebKeyFromCertFile(saCert)
	if err != nil {
		return err
	}
	resp.Keys = append(resp.Keys, *newKey)

	keysJSON, err := json.MarshalIndent(resp, "", "    ")
	if err != nil {
		return fmt.Errorf("marshaling keys.json: %v", err)
	}

	// upload the modified keys json to s3 with the public read access
	if err = s3.Upload(awsSession, keysJSON, keysFilename, envVars.S3Bucket, s3.WithPublicRead()); err != nil {
		return fmt.Errorf("upload new keys.json to s3: %v", err)
	}

	return nil
}

func (e *ClusterE2ETest) getClusterSACert(ctx context.Context) ([]byte, error) {
	secret, err := e.KubectlClient.GetSecretFromNamespace(ctx, e.KubeconfigFilePath(), fmt.Sprintf("%s-sa", e.ClusterName), constants.EksaSystemNamespace)
	if err != nil {
		return nil, err
	}

	cert, found := secret.Data["tls.crt"]
	if !found {
		return nil, errors.New("cluster SA secret doesn't contain tls.crt")
	}

	return cert, nil
}

func getJSONWebKeyFromCertFile(cert []byte) (*jose.JSONWebKey, error) {
	block, _ := pem.Decode(cert)
	if block == nil {
		return nil, errors.New("decoding public key")
	}
	pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parsing public key content: %v", err)
	}
	switch pubKey.(type) {
	case *rsa.PublicKey:
	default:
		return nil, errors.New("public key is not RSA")
	}

	var alg jose.SignatureAlgorithm
	switch pubKey.(type) {
	case *rsa.PublicKey:
		alg = jose.RS256
	default:
		return nil, fmt.Errorf("invalid public key type %T, must be *rsa.PrivateKey", pubKey)
	}

	kid, err := keyIDFromPublicKey(pubKey)
	if err != nil {
		return nil, err
	}

	return &jose.JSONWebKey{
		Key:       pubKey,
		KeyID:     kid,
		Algorithm: string(alg),
		Use:       "sig",
	}, nil
}

func keyIDFromPublicKey(publicKey interface{}) (string, error) {
	publicKeyDERBytes, err := x509.MarshalPKIXPublicKey(publicKey)
	if err != nil {
		return "", fmt.Errorf("failed to serialize public key to DER format: %v", err)
	}
	hasher := crypto.SHA256.New()
	hasher.Write(publicKeyDERBytes)
	publicKeyDERHash := hasher.Sum(nil)
	return base64.RawURLEncoding.EncodeToString(publicKeyDERHash), nil
}
