package upgradevalidations_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

func TestValidateEksaControllerReady(t *testing.T) {
	tests := []struct {
		name                   string
		wantErr                error
		upgradeVersion         v1alpha1.KubernetesVersion
		getDeploymentsResponse string
	}{
		{
			name:                   "FailureNoDeployment",
			wantErr:                errors.New("failed to find EKS-A controller deployment eksa-controller-manager in namespace eksa-system"),
			getDeploymentsResponse: "testdata/empty_get_deployments_response.json",
		},
		{
			name:                   "FailureReplicasNotReady",
			wantErr:                errors.New("EKS-A controller deployment eksa-controller-manager replicas in namespace eksa-system are not ready; ready=0, want=1"),
			getDeploymentsResponse: "testdata/eksa_controller_deployment_response_no_ready_replicas.json",
		},
		{
			name:                   "FailureZeroReplicas",
			wantErr:                errors.New("EKS-A controller deployment eksa-controller-manager in namespace eksa-system is scaled to 0 replicas; should be at least one replcias"),
			getDeploymentsResponse: "testdata/eksa_controller_deployment_response_no_replicas.json",
		},
		{
			name:                   "SuccessReplicasReady",
			wantErr:                nil,
			getDeploymentsResponse: "testdata/eksa_controller_deployment_response.json",
		},
	}

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getDeploymentsResponse)
			e.EXPECT().Execute(ctx, []string{"get", "deployments", "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)
			err := upgradevalidations.ValidateEksaSystemComponents(ctx, k, cluster)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
