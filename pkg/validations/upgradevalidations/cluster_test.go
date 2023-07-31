package upgradevalidations_test

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"testing"

	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

const testclustername string = "testcluster"

type UnAuthKubectlClient struct {
	*executables.Kubectl
	*kubernetes.UnAuthClient
}

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

	k, ctx, cluster, e := validations.NewKubectl(t)
	uk := kubernetes.NewUnAuthClient(k)

	cluster.Name = testclustername
	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.getClusterResponse)
			e.EXPECT().Execute(ctx, []string{"get", capiClustersResourceType, "-o", "json", "--kubeconfig", cluster.KubeconfigFile, "--namespace", constants.EksaSystemNamespace}).Return(*bytes.NewBufferString(fileContent), nil)
			err := upgradevalidations.ValidateClusterObjectExists(ctx, UnAuthKubectlClient{k, uk}, cluster)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}

var capiClustersResourceType = fmt.Sprintf("clusters.%s", clusterv1.GroupVersion.Group)
