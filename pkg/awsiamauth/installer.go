package awsiamauth

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/kubeconfig"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/types"
)

// KubernetesClient provides Kubernetes API access.
type KubernetesClient interface {
	Apply(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetAPIServerURL(ctx context.Context, cluster *types.Cluster) (string, error)
	GetClusterCACert(ctx context.Context, cluster *types.Cluster, clusterName string) ([]byte, error)
	GetAWSIAMKubeconfigSecretValue(ctx context.Context, cluster *types.Cluster, clusterName string) ([]byte, error)
}

// Installer provides the necessary behavior for installing the AWS IAM Authenticator.
type Installer struct {
	k8s              KubernetesClient
	writer           filewriter.FileWriter
	kubeconfigWriter kubeconfig.Writer
}

// NewInstaller creates a new installer instance.
func NewInstaller(
	k8s KubernetesClient,
	writer filewriter.FileWriter,
	kubeconfigWriter kubeconfig.Writer,
) *Installer {
	return &Installer{
		k8s:              k8s,
		writer:           writer,
		kubeconfigWriter: kubeconfigWriter,
	}
}

// GenerateWorkloadKubeconfig generates the AWS IAM auth kubeconfig.
func (i *Installer) GenerateWorkloadKubeconfig(
	ctx context.Context,
	management, workload *types.Cluster,
	spec *cluster.Spec,
) error {
	fileName := fmt.Sprintf("%s-aws.kubeconfig", workload.Name)

	fsOptions := []filewriter.FileOptionsFunc{filewriter.PersistentFile, filewriter.Permission0600}
	fh, path, err := i.writer.Create(
		fileName,
		fsOptions...,
	)
	if err != nil {
		return err
	}

	defer fh.Close()

	decodedKubeconfigSecretValue, err := i.k8s.GetAWSIAMKubeconfigSecretValue(
		ctx,
		management,
		workload.Name,
	)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	err = i.kubeconfigWriter.WriteKubeconfigContent(ctx, workload.Name, decodedKubeconfigSecretValue, fh)
	if err != nil {
		return fmt.Errorf("writing aws-iam-authenticator kubeconfig to %s: %v", path, err)
	}

	logger.V(3).Info("Generated aws-iam-authenticator kubeconfig", "kubeconfig", path)
	return nil
}

// GenerateManagementKubeconfig generates the AWS IAM auth kubeconfig.
func (i *Installer) GenerateManagementKubeconfig(
	ctx context.Context,
	cluster *types.Cluster,
) error {
	fileName := fmt.Sprintf("%s-aws.kubeconfig", cluster.Name)

	fsOptions := []filewriter.FileOptionsFunc{filewriter.PersistentFile, filewriter.Permission0600}
	fh, path, err := i.writer.Create(
		fileName,
		fsOptions...,
	)
	if err != nil {
		return err
	}

	defer fh.Close()

	decodedKubeconfigSecretValue, err := i.k8s.GetAWSIAMKubeconfigSecretValue(
		ctx,
		cluster,
		cluster.Name,
	)
	if err != nil {
		return fmt.Errorf("generating aws-iam-authenticator kubeconfig: %v", err)
	}

	err = i.kubeconfigWriter.WriteKubeconfigContent(ctx, cluster.Name, decodedKubeconfigSecretValue, fh)
	if err != nil {
		return fmt.Errorf("writing aws-iam-authenticator kubeconfig to %s: %v", path, err)
	}

	logger.V(3).Info("Generated aws-iam-authenticator kubeconfig", "kubeconfig", path)
	return nil
}
