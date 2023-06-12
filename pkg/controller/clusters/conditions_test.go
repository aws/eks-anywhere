package clusters_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util/conditions"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	ctrlclusters "github.com/aws/eks-anywhere/pkg/controller/clusters"
)

func conditionChecker(result *v1beta1.Condition, err error, duration time.Duration) ctrlclusters.ConditionChecker {
	return func(ctx context.Context, client client.Client, clusterSpec *cluster.Spec) (*v1beta1.Condition, error) {
		time.Sleep(duration)
		return result, err
	}
}

func controlPlaneInitializedTrueCondition() *v1beta1.Condition {
	return conditions.TrueCondition(v1beta1.ReadyCondition)
}

func controlPlaneReadyFalseCondition() *v1beta1.Condition {
	return conditions.FalseCondition(v1beta1.ControlPlaneReadyCondition, "", v1beta1.ConditionSeverityInfo, "")
}

func readyFalseCondition() *v1beta1.Condition {
	return conditions.FalseCondition(v1beta1.ReadyCondition, "", v1beta1.ConditionSeverityInfo, "")
}

func TestClusterConditionFetcher(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	clusterSpec := test.NewClusterSpec()
	g := NewWithT(t)
	tests := []struct {
		name           string
		checkers       []ctrlclusters.ConditionChecker
		wantConditions []*v1beta1.Condition
		wantErr        []string
	}{
		{
			name: "run with no error",
			checkers: []ctrlclusters.ConditionChecker{
				conditionChecker(controlPlaneInitializedTrueCondition(), nil, 1*time.Second),
				conditionChecker(controlPlaneReadyFalseCondition(), nil, 2*time.Second),
				conditionChecker(readyFalseCondition(), nil, 3*time.Second),
			},
			wantConditions: []*v1beta1.Condition{
				controlPlaneInitializedTrueCondition(),
				controlPlaneReadyFalseCondition(),
				readyFalseCondition(),
			},
			wantErr: []string{},
		},
		{
			name: "run with error returned",
			checkers: []ctrlclusters.ConditionChecker{
				conditionChecker(nil, errors.New("test error"), 1*time.Second),
			},
			wantConditions: []*v1beta1.Condition{
				nil,
			},
			wantErr: []string{"test error"},
		},
		{
			name: "run with multiple errors aggregated returned",
			checkers: []ctrlclusters.ConditionChecker{
				conditionChecker(nil, errors.New("test error 1"), 1*time.Second),
				conditionChecker(nil, errors.New("test error 2"), 2*time.Second),
				conditionChecker(readyFalseCondition(), nil, 3*time.Second),
			},
			wantConditions: []*v1beta1.Condition{
				nil,
				nil,
				readyFalseCondition(),
			},
			wantErr: []string{"test error 1", "test error 2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cf := ctrlclusters.NewConditionFetcher(tt.checkers...)
			ctx := context.Background()
			conditions, err := cf.RunAll(ctx, client, clusterSpec)

			if len(tt.wantErr) == 0 {
				g.Expect(err).To(BeNil())
			} else {
				for _, wantErr := range tt.wantErr {
					g.Expect(err).To(MatchError(ContainSubstring(wantErr)))
				}
			}

			g.Expect(tt.wantConditions).To(Equal(conditions))

		})
	}
}
