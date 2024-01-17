package kubeconfig

import (
	"bytes"
	"context"
	"io"
	"time"

	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/clusterapi"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

// ClientFactory builds Kubernetes clients.
type ClientFactory interface {
	// BuildClientFromKubeconfig builds a Kubernetes client from a kubeconfig file.
	BuildClientFromKubeconfig(kubeconfigPath string) (kubernetes.Client, error)
}

// ClusterAPIKubeconfigSecretWriter reads the kubeconfig secret on a cluster and copies the contents to a writer.
type ClusterAPIKubeconfigSecretWriter struct {
	client  ClientFactory
	timeout time.Duration
	backoff time.Duration
}

// WriterOpt allows to configure [KubeconfigWriter].
type WriterOpt func(*ClusterAPIKubeconfigSecretWriter)

// WithTimeout sets the optional timeout for a KubeconfigWriter.
func WithTimeout(timeout time.Duration) WriterOpt {
	return func(writer *ClusterAPIKubeconfigSecretWriter) {
		writer.timeout = timeout
	}
}

// WithBackoff sets the optional backoff duration for a KubeconfigWriter.
func WithBackoff(backoff time.Duration) WriterOpt {
	return func(writer *ClusterAPIKubeconfigSecretWriter) {
		writer.backoff = backoff
	}
}

// NewClusterAPIKubeconfigSecretWriter creates a ClusterAPIKubeconfigSecretWriter.
func NewClusterAPIKubeconfigSecretWriter(unauthClient ClientFactory, opts ...WriterOpt) ClusterAPIKubeconfigSecretWriter {
	kr := &ClusterAPIKubeconfigSecretWriter{
		client:  unauthClient,
		timeout: time.Minute,
		backoff: time.Second,
	}

	for _, o := range opts {
		o(kr)
	}

	return *kr
}

// WriteKubeconfig retrieves the contents of the specified cluster's kubeconfig from a secret and copies it to an io.Writer.
func (kr ClusterAPIKubeconfigSecretWriter) WriteKubeconfig(ctx context.Context, clusterName, kubeconfigPath string, w io.Writer) error {
	rawKubeconfig, err := kr.GetClusterKubeconfig(ctx, clusterName, kubeconfigPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(w, bytes.NewReader(rawKubeconfig)); err != nil {
		return err
	}

	return nil
}

// GetClusterKubeconfig gets the cluster's kubeconfig from the secret.
func (kr ClusterAPIKubeconfigSecretWriter) GetClusterKubeconfig(ctx context.Context, clusterName, kubeconfigPath string) ([]byte, error) {
	kubeconfigSecret := &corev1.Secret{}
	var kubeClient kubernetes.Client
	kubeClient, err := kr.client.BuildClientFromKubeconfig(kubeconfigPath)
	if err != nil {
		return nil, err
	}

	err = retrier.New(
		kr.timeout,
		retrier.WithRetryPolicy(retrier.BackOffPolicy(kr.backoff)),
	).Retry(func() error {
		if err = kubeClient.Get(ctx, clusterapi.ClusterKubeconfigSecretName(clusterName), constants.EksaSystemNamespace, kubeconfigSecret); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return kubeconfigSecret.Data["value"], nil
}
