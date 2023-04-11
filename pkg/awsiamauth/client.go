package awsiamauth

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/types"
)

// Client is a Kubernetes client.
type Client interface {
	ApplyKubeSpecFromBytes(ctx context.Context, cluster *types.Cluster, data []byte) error
	GetApiServerUrl(ctx context.Context, cluster *types.Cluster) (string, error)
	GetObject(ctx context.Context, resourceType string, name string, namespace string, kubeconfig string, obj runtime.Object) error
}

// RetrierClient wraps basic kubernetes API operations around a retrier.
type RetrierClient struct {
	client  Client
	retrier retrier.Retrier
}

// RetrierClientOpt allows to customize a RetrierClient
// on construction.
type RetrierClientOpt func(*RetrierClient)

// RetrierClientRetrier allows to use a custom retrier.
func RetrierClientRetrier(retrier retrier.Retrier) RetrierClientOpt {
	return func(u *RetrierClient) {
		u.retrier = retrier
	}
}

// NewRetrierClient constructs a new RetrierClient.
func NewRetrierClient(client Client, opts ...RetrierClientOpt) RetrierClient {
	c := &RetrierClient{
		client:  client,
		retrier: *retrier.NewWithMaxRetries(10, time.Second),
	}

	for _, opt := range opts {
		opt(c)
	}

	return *c
}

// Apply creates/updates the data objects for a cluster.
func (c RetrierClient) Apply(ctx context.Context, cluster *types.Cluster, data []byte) error {
	return c.retrier.Retry(
		func() error {
			return c.client.ApplyKubeSpecFromBytes(ctx, cluster, data)
		},
	)
}

// GetAPIServerURL gets the api server url from K8s config.
func (c RetrierClient) GetAPIServerURL(ctx context.Context, cluster *types.Cluster) (string, error) {
	var url string
	err := c.retrier.Retry(
		func() error {
			var err error
			url, err = c.client.GetApiServerUrl(ctx, cluster)
			return err
		},
	)
	if err != nil {
		return "", err
	}

	return url, nil
}

// GetClusterCACert gets the ca cert for a cluster from a secret.
func (c RetrierClient) GetClusterCACert(ctx context.Context, cluster *types.Cluster, clusterName string) ([]byte, error) {
	secret := &corev1.Secret{}
	secretName := fmt.Sprintf("%s-ca", clusterName)
	err := c.retrier.Retry(
		func() error {
			return c.client.GetObject(ctx, "secret", secretName, constants.EksaSystemNamespace, cluster.KubeconfigFile, secret)
		},
	)
	if err != nil {
		return nil, err
	}

	if crt, ok := secret.Data["tls.crt"]; ok {
		b64EncodedCrt := make([]byte, base64.StdEncoding.EncodedLen(len(crt)))
		base64.StdEncoding.Encode(b64EncodedCrt, crt)
		return b64EncodedCrt, nil
	}

	return nil, fmt.Errorf("tls.crt not found in secret [%s]", secretName)
}
