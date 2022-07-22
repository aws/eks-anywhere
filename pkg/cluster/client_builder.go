package cluster

import (
	"context"
	"fmt"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
)

// Client is a kubernetes API client
type Client interface {
	Get(ctx context.Context, name, namespace string, obj kubernetes.Object) error
}

// ConfigClientProcessor updates a Config retrieving objects from
// the API server through a client
type ConfigClientProcessor func(ctx context.Context, client Client, c *Config) error

// ConfigClientBuilder allows to register processors to build a Config
// using a cluster client, retrieving the api objects from the API server
type ConfigClientBuilder struct {
	processors []ConfigClientProcessor
}

// NewConfigClientBuilder builds a new ConfigClientBuilder with
// no processors registered
func NewConfigClientBuilder() *ConfigClientBuilder {
	return &ConfigClientBuilder{}
}

// Register stores processors to be used during Build
func (b *ConfigClientBuilder) Register(processors ...ConfigClientProcessor) *ConfigClientBuilder {
	b.processors = append(b.processors, processors...)
	return b
}

// Build constructs a Config for a cluster using the registered processors
func (b *ConfigClientBuilder) Build(ctx context.Context, client Client, cluster *anywherev1.Cluster) (*Config, error) {
	c := &Config{
		Cluster: cluster,
	}

	for _, p := range b.processors {
		if err := p(ctx, client, c); err != nil {
			return nil, fmt.Errorf("building Config from a cluster client: %v", err)
		}
	}

	return c, nil
}
