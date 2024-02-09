package awsiamauth

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesClient provides Kubernetes API access.
type KubernetesClient interface {
	Apply(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetAPIServerURL(ctx context.Context, cluster *types.Cluster) (string, error)
	GetClusterCACert(ctx context.Context, cluster *types.Cluster, clusterName string) ([]byte, error)
}

// Installer provides the necessary behavior for installing the AWS IAM Authenticator.
type Installer struct {
	certgen         crypto.CertificateGenerator
	templateBuilder *TemplateBuilder
	clusterID       uuid.UUID
	k8s             KubernetesClient
	writer          filewriter.FileWriter
}

// NewInstaller creates a new installer instance.
func NewInstaller(
	certgen crypto.CertificateGenerator,
	clusterID uuid.UUID,
	k8s KubernetesClient,
	writer filewriter.FileWriter,
) *Installer {
	return &Installer{
		certgen:         certgen,
		templateBuilder: &TemplateBuilder{},
		clusterID:       clusterID,
		k8s:             k8s,
		writer:          writer,
	}
}

// CreateAndInstallAWSIAMAuthCASecret creates a Kubernetes Secret in cluster containing a
// self-signed certificate and key for a cluster identified by clusterName.
func (i *Installer) CreateAndInstallAWSIAMAuthCASecret(ctx context.Context, managementCluster *types.Cluster, clusterName string) error {
	secret, err := i.generateCertKeyPairSecret(clusterName)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator ca secret: %v", err)
	}

	if err = i.k8s.Apply(ctx, managementCluster, secret); err != nil {
		return fmt.Errorf("applying aws-iam-authenticator ca secret: %v", err)
	}

	return nil
}

// InstallAWSIAMAuth installs AWS IAM Authenticator deployment manifests into the workload cluster.
// It writes a Kubeconfig to disk for kubectl access using AWS IAM Authentication.
func (i *Installer) InstallAWSIAMAuth(
	ctx context.Context,
	management, workload *types.Cluster,
	spec *cluster.Spec,
) error {
	manifest, err := i.generateManifest(spec)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator manifest: %v", err)
	}

	if err = i.k8s.Apply(ctx, workload, manifest); err != nil {
		return fmt.Errorf("applying aws-iam-authenticator manifest: %v", err)
	}

	if err = i.GenerateKubeconfig(ctx, management, workload, spec); err != nil {
		return err
	}
	return nil
}

// UpgradeAWSIAMAuth upgrades an AWS IAM Authenticator deployment in cluster.
func (i *Installer) UpgradeAWSIAMAuth(ctx context.Context, cluster *types.Cluster, spec *cluster.Spec) error {
	awsIamAuthManifest, err := i.generateManifestForUpgrade(spec)
	if err != nil {
		return fmt.Errorf("generating manifest: %v", err)
	}

	err = i.k8s.Apply(ctx, cluster, awsIamAuthManifest)
	if err != nil {
		return fmt.Errorf("applying manifest: %v", err)
	}

	return nil
}

func (i *Installer) generateManifest(clusterSpec *cluster.Spec) ([]byte, error) {
	return i.templateBuilder.GenerateManifest(clusterSpec, i.clusterID)
}

func (i *Installer) generateManifestForUpgrade(clusterSpec *cluster.Spec) ([]byte, error) {
	return i.templateBuilder.GenerateManifest(clusterSpec, uuid.Nil)
}

func (i *Installer) generateCertKeyPairSecret(managementClusterName string) ([]byte, error) {
	return i.templateBuilder.GenerateCertKeyPairSecret(i.certgen, managementClusterName)
}

func (i *Installer) generateInstallerKubeconfig(clusterSpec *cluster.Spec, serverURL, tlsCert string) ([]byte, error) {
	return i.templateBuilder.GenerateKubeconfig(clusterSpec, i.clusterID, serverURL, tlsCert)
}

// GenerateKubeconfig generates the AWS IAM auth kubeconfig.
func (i *Installer) GenerateKubeconfig(
	ctx context.Context,
	management, workload *types.Cluster,
	spec *cluster.Spec,
) error {
	fileName := fmt.Sprintf("%s-aws.kubeconfig", workload.Name)

	serverURL, err := i.k8s.GetAPIServerURL(ctx, workload)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	tlsCert, err := i.k8s.GetClusterCACert(
		ctx,
		management,
		workload.Name,
	)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	awsIamAuthKubeconfigContent, err := i.generateInstallerKubeconfig(spec, serverURL, string(tlsCert))
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	writtenFile, err := i.writer.Write(
		fileName,
		awsIamAuthKubeconfigContent,
		filewriter.PersistentFile,
		filewriter.Permission0600,
	)
	if err != nil {
		return fmt.Errorf("writing aws-iam-authenticator kubeconfig to %s: %v", writtenFile, err)
	}

	logger.V(3).Info("Generated aws-iam-authenticator kubeconfig", "kubeconfig", writtenFile)

	return nil
}
