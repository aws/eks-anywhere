package createvalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"sigs.k8s.io/cluster-api/api/v1alpha3"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/createvalidations"
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
			name:               "SuccessNoClusters",
			wantErr:            nil,
			getClusterResponse: "testdata/empty_get_cluster_response.json",
		},
		{
			name:               "FailureClusterNameExists",
			wantErr:            errors.New("cluster name testcluster already exists"),
			getClusterResponse: "testdata/cluster_name_exists.json",
		},
		{
			name:               "SuccessClusterNotInList",
			wantErr:            nil,
			getClusterResponse: "testdata/name_not_in_list.json",
		},
	}

	k, ctx, cluster, e := validations.NewKubectl(t)
	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(ctx, []string{"get", capiClustersResourceType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)
			err := createvalidations.ValidateClusterNameIsUnique(ctx, k, cluster, testclustername)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

var capiClustersResourceType = fmt.Sprintf("clusters.%s", v1alpha3.GroupVersion.Group)
