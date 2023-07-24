package defaulting_test

import (
	"context"
	"fmt"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/defaulting"
	eksaerrors "github.com/aws/eks-anywhere/pkg/errors"
)

func ExampleRunner_RunAll() {
	r := defaulting.NewRunner[cluster.Spec]()

	r.Register(
		func(ctx context.Context, spec cluster.Spec) (cluster.Spec, error) {
			if spec.Cluster.Spec.KubernetesVersion == "" {
				spec.Cluster.Spec.KubernetesVersion = anywherev1.Kube124
			}
			return spec, nil
		},
		func(ctx context.Context, spec cluster.Spec) (cluster.Spec, error) {
			if spec.Cluster.Spec.ControlPlaneConfiguration.Count == 0 {
				spec.Cluster.Spec.ControlPlaneConfiguration.Count = 3
			}
			return spec, nil
		},
	)

	ctx := context.Background()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Spec.KubernetesVersion = "1.24"
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 5
	})
	updatedSpec, agg := r.RunAll(ctx, *spec)
	if agg != nil {
		printErrors(agg)
		return
	}

	fmt.Println("Cluster config is valid")
	fmt.Printf("Cluster is for kube version: %s\n", updatedSpec.Cluster.Spec.KubernetesVersion)
	fmt.Printf("Cluster CP replicas is: %d\n", updatedSpec.Cluster.Spec.ControlPlaneConfiguration.Count)

	// Output:
	// Cluster config is valid
	// Cluster is for kube version: 1.24
	// Cluster CP replicas is: 5
}

func printErrors(agg eksaerrors.Aggregate) {
	fmt.Println("Failed assigning cluster spec defaults")
	for _, err := range agg.Errors() {
		msg := "- " + err.Error()
		fmt.Println(msg)
	}
}
