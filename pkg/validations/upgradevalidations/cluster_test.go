package upgradevalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

const testclustername string = "testcluster"

func TestValidateClusterPresent(t *testing.T) {
	tests := []struct {
		name               string
		wantErr            error
		upgradeVersion     v1alpha1.KubernetesVersion
		getClusterResponse string
	}{
		{
			name:               "FailureNoClusters",
			wantErr:            errors.New("no CAPI cluster objects present on workload cluster testcluster"),
			getClusterResponse: "testdata/empty_get_cluster_response.json",
		},
		{
			name:               "FailureClusterNotPresent",
			wantErr:            errors.New("couldn't find CAPI cluster object for cluster with name testcluster"),
			getClusterResponse: "testdata/no_target_cluster_response.json",
		},
		{
			name:               "SuccessClusterPresent",
			wantErr:            nil,
			getClusterResponse: "testdata/target_cluster_response.json",
		},
	}

	k, ctx, cluster, e := newKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(ctx, []string{"get", capiClustersResourceName, "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)
			err := upgradevalidations.ValidateClusterObjectExists(ctx, k, cluster)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

var capiClustersResourceName = fmt.Sprintf("clusters.%s", v1alpha3.GroupVersion.Group)
