package cli_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
)

func TestNewDeleteClusterDefaulter(t *testing.T) {
	g := NewWithT(t)

	ns := "test-ns"
	nsDefaulter := cluster.NewNamespaceDefaulter(ns)

	c := baseCluster()

	clusterSpec := &cluster.Spec{
		Config: &cluster.Config{
			Cluster: c,
		},
	}

	r := defaulting.NewRunner[*cluster.Spec]()
	r.Register(
		nsDefaulter.NamespaceDefault,
	)

	got := cli.NewDeleteClusterDefaulter(nsDefaulter)

	g.Expect(got).NotTo(BeNil())

	clusterSpec, err := got.Run(context.Background(), clusterSpec)

	g.Expect(err).To(BeNil())
	g.Expect(clusterSpec.Cluster.Namespace).To(ContainSubstring(ns))
}
