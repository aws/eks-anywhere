package upgradevalidations_test

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/clients/kubernetes"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/validations"
	"github.com/aws/eks-anywhere/pkg/validations/upgradevalidations"
)

type UnAuthKubectlClient struct {
	*executables.Kubectl
	*kubernetes.UnAuthClient
}

func TestValidateVersionSkew(t *testing.T) {
	tests := []struct {
		name                  string
		wantErr               error
		upgradeVersion        v1alpha1.KubernetesVersion
		serverVersionResponse string
	}{
		{
			name:                  "FailureTwoMinorVersions",
			wantErr:               errors.New("WARNING: version difference between upgrade version (1.20) and server version (1.18) do not meet the supported version increment of +1"),
			upgradeVersion:        v1alpha1.Kube120,
			serverVersionResponse: "testdata/kubectl_version_server_118.json",
		},
		{
			name:                  "FailureMinusOneMinorVersion",
			wantErr:               errors.New("WARNING: version difference between upgrade version (1.19) and server version (1.20) do not meet the supported version increment of +1"),
			upgradeVersion:        v1alpha1.Kube119,
			serverVersionResponse: "testdata/kubectl_version_server_120.json",
		},
		{
			name:                  "SuccessSameVersion",
			wantErr:               nil,
			upgradeVersion:        v1alpha1.Kube119,
			serverVersionResponse: "testdata/kubectl_version_server_119.json",
		},
		{
			name:                  "SuccessOneMinorVersion",
			wantErr:               nil,
			upgradeVersion:        v1alpha1.Kube120,
			serverVersionResponse: "testdata/kubectl_version_server_119.json",
		},
	}

	k, ctx, cluster, e := validations.NewKubectl(t)
	uk := kubernetes.NewUnAuthClient(k)

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			fileContent := test.ReadFile(t, tc.serverVersionResponse)
			e.EXPECT().Execute(ctx, []string{"version", "-o", "json", "--kubeconfig", cluster.KubeconfigFile}).Return(*bytes.NewBufferString(fileContent), nil)
			err := upgradevalidations.ValidateServerVersionSkew(ctx, tc.upgradeVersion, cluster, UnAuthKubectlClient{k, uk})
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
