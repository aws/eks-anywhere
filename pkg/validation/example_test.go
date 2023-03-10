package validation_test

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/test"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cli"
	"github.com/aws/eks-anywhere/pkg/cluster"
	eksaerrors "github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/validation"
)

func ExampleRunner_RunAll_validations() {
	log := test.NewNullLogger()
	r := validation.NewRunner[*cluster.Spec](validation.WithMaxJobs(10))
	cpValidator := newControlPlaneValidator(log)

	r.Register(
		func(ctx context.Context, spec *cluster.Spec) error {
			if spec.Cluster.Spec.KubernetesVersion == "" {
				return errors.New("kubernetesVersion can't be empty")
			}

			return nil
		},
		validation.Sequentially(
			func(ctx context.Context, spec *cluster.Spec) error {
				if spec.Cluster.Name == "" {
					return validation.WithRemediation(
						errors.New("cluster name is empty"),
						"set a name for your cluster",
					)
				}

				return nil
			},
			cpValidator.validateCount,
		),
	)

	ctx := context.Background()
	spec := test.NewClusterSpec(func(s *cluster.Spec) {
		s.Cluster.Name = ""
		s.Cluster.Spec.KubernetesVersion = anywherev1.Kube124
		s.Cluster.Spec.ControlPlaneConfiguration.Count = 0
	})
	agg := r.RunAll(ctx, spec)
	if agg != nil {
		printErrors(agg)
		return
	}

	fmt.Println("Cluster config is valid")

	// Output:
	// Invalid cluster config
	// - cluster name is empty. Try to set a name for your cluster
	// - control plane node count can't be 0
}

func printErrors(agg eksaerrors.Aggregate) {
	fmt.Println("Invalid cluster config")
	for _, err := range agg.Errors() {
		msg := "- " + err.Error()
		if validation.IsRemediable(err) {
			msg += ". Try to " + validation.Remediation(err)
		}
		fmt.Println(msg)
	}
}

type controlPlaneValidator struct {
	log logr.Logger
}

func newControlPlaneValidator(log logr.Logger) *controlPlaneValidator {
	return &controlPlaneValidator{log: log}
}

func (v *controlPlaneValidator) validateCount(ctx context.Context, spec *cluster.Spec) error {
	if spec.Cluster.Spec.ControlPlaneConfiguration.Count == 0 {
		cli.ValidationFailed(v.log, "Control plane invalid")
		return errors.New("control plane node count can't be 0")
	}

	cli.ValidationPassed(v.log, "Control plane valid")
	return nil
}
