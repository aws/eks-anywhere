package cluster_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/cluster/mocks"
)

func TestConfigClientBuilderBuildSuccess(t *testing.T) {
	g := NewWithT(t)

	builder := cluster.NewConfigClientBuilder()

	timesCalled := 0
	processor := func(ctx context.Context, client cluster.Client, c *cluster.Config) error {
		timesCalled++
		return nil
	}

	// register twice, expect two calls
	builder.Register(processor, processor)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{}

	config, err := builder.Build(ctx, client, cluster)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(config).NotTo(BeNil())
	g.Expect(config.Cluster).To(Equal(cluster))
	g.Expect(timesCalled).To(Equal(2), "processor should be called 2 times")
}

func TestConfigClientBuilderBuildError(t *testing.T) {
	g := NewWithT(t)

	builder := cluster.NewConfigClientBuilder()

	timesCalled := 0
	processor := func(ctx context.Context, client cluster.Client, c *cluster.Config) error {
		timesCalled++
		return nil
	}

	processorError := func(ctx context.Context, client cluster.Client, c *cluster.Config) error {
		return errors.New("processor error")
	}

	builder.Register(processor, processorError)

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mocks.NewMockClient(ctrl)
	cluster := &anywherev1.Cluster{}

	_, err := builder.Build(ctx, client, cluster)
	g.Expect(err).To(MatchError(ContainSubstring("processor error")))
	g.Expect(timesCalled).To(Equal(1), "processor should be called 1 times")
}
