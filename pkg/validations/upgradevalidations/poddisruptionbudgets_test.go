package upgradevalidations_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	policy "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/mocks"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidatePodDisruptionBudgets(t *testing.T) {
	type args struct {
		ctx     context.Context
		k       validations.KubectlClient
		cluster *types.Cluster
		pdbList *policy.PodDisruptionBudgetList
	}
	mockCtrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(mockCtrl)
	c := types.Cluster{
		KubeconfigFile: "test.kubeconfig",
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "PDBs exist on cluster",
			args: args{
				ctx:     context.Background(),
				k:       k,
				cluster: &c,
				pdbList: &policy.PodDisruptionBudgetList{
					Items: []policy.PodDisruptionBudget{
						{
							Spec: policy.PodDisruptionBudgetSpec{
								MinAvailable: &intstr.IntOrString{
									Type:   intstr.Int,
									IntVal: 0,
								},
							},
						},
					},
				},
			},
			wantErr: fmt.Errorf("one or more pod disruption budgets were detected on the cluster. Use the --skip-validations=%s flag if you wish to skip the validations for pod disruption budgets and proceed with the upgrade operation", validations.PDB),
		},
		{
			name: "PDBs don't exist on cluster",
			args: args{
				ctx:     context.Background(),
				k:       k,
				cluster: &c,
				pdbList: &policy.PodDisruptionBudgetList{},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		podDisruptionBudgets := &policy.PodDisruptionBudgetList{}
		k.EXPECT().List(tt.args.ctx, tt.args.cluster.KubeconfigFile, podDisruptionBudgets).DoAndReturn(func(_ context.Context, _ string, objs kubernetes.ObjectList) error {
			tt.args.pdbList.DeepCopyInto(objs.(*policy.PodDisruptionBudgetList))
			return nil
		})

		t.Run(tt.name, func(t *testing.T) {
			if err := upgradevalidations.ValidatePodDisruptionBudgets(tt.args.ctx, tt.args.k, tt.args.cluster); !reflect.DeepEqual(err, tt.wantErr) {
				t.Errorf("ValidatePodDisruptionBudgets() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidatePodDisruptionBudgetsFailure(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	k := mocks.NewMockKubectlClient(mockCtrl)
	c := types.Cluster{
		KubeconfigFile: "test.kubeconfig",
	}
	ctx := context.Background()
	pdbList := &policy.PodDisruptionBudgetList{}

	k.EXPECT().List(ctx, c.KubeconfigFile, pdbList).Return(errors.New("listing cluster pod disruption budgets for upgrade"))

	wantErr := errors.New("listing cluster pod disruption budgets for upgrade")

	err := upgradevalidations.ValidatePodDisruptionBudgets(ctx, k, &c)
	if err != nil && !strings.Contains(err.Error(), wantErr.Error()) {
		t.Errorf("ValidatePodDisruptionBudgets() error = %v, wantErr %v", err, wantErr)
	}
}
