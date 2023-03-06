package cluster_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/retrier"
	clusterf "github.com/aws/eks-anywhere/test/framework/cluster"
)

func validations() []clusterf.StateValidation {
	r := retrier.NewWithMaxRetries(2, time.Second)
	return []clusterf.StateValidation{
		clusterf.RetriableStateValidation(r, func(ctx context.Context, vc clusterf.StateValidationConfig) error {
			if vc.ClusterSpec == nil {
				return fmt.Errorf("spec not defined")
			}
			return nil
		}),
		clusterf.RetriableStateValidation(r, func(ctx context.Context, vc clusterf.StateValidationConfig) error {
			if vc.ClusterSpec != nil && vc.ClusterSpec.Cluster.Name != "test-cluster" {
				return fmt.Errorf("cluster name not valid")
			}
			return nil
		}),
		clusterf.RetriableStateValidation(r, func(ctx context.Context, vc clusterf.StateValidationConfig) error {
			if vc.ClusterSpec != nil && vc.ClusterSpec.Cluster.Namespace != "test-namespace" {
				return fmt.Errorf("cluster namespace not valid")
			}
			return nil
		}),
	}
}

func TestClusterStateValidatorValidate(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name        string
		clusterSpec *cluster.Spec
		wantErr     []string
	}{
		{
			name: "validate success",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = &v1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster",
						Namespace: "test-namespace",
					},
				}
			}),
			wantErr: []string{},
		},
		{
			name:        "cluster spec nil",
			clusterSpec: nil,
			wantErr:     []string{"spec not defined"},
		},
		{
			name: "invalid cluster name in spec",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = &v1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-invalid-name",
						Namespace: "test-namespace",
					},
				}
			}),
			wantErr: []string{"cluster name not valid"},
		},
		{
			name: "invalid cluster name in spec",
			clusterSpec: test.NewClusterSpec(func(s *cluster.Spec) {
				s.Cluster = &v1alpha1.Cluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-cluster-name-invalid",
						Namespace: "test-namespace-invalid",
					},
				}
			}),
			wantErr: []string{"cluster name not valid", "cluster namespace not valid"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cv := clusterf.NewStateValidator(clusterf.StateValidationConfig{
				ClusterSpec: tt.clusterSpec,
			})
			cv.WithValidations(validations()...)
			ctx := context.Background()
			err := cv.Validate(ctx)
			if len(tt.wantErr) == 0 {
				g.Expect(err).To(BeNil())
			} else {
				for _, wantErr := range tt.wantErr {
					g.Expect(err).To(MatchError(ContainSubstring(wantErr)))
				}
			}
		})
	}
}
