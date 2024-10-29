package framework

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

func TestValidatePackageBundleControllerRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	testPbc := &packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			DefaultRegistry:      "123.ecr",
			DefaultImageRegistry: "123.ecr",
		},
	}
	respJSON, err := json.Marshal(testPbc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	expectedParam := []string{"get", "packagebundlecontroller.packages.eks.amazonaws.com", "test", "-o", "json", "--kubeconfig", "test/test-eks-a-cluster.kubeconfig", "--namespace", "eksa-packages", "--ignore-not-found=true"}
	t.Run("CuratedPackagesTest", func(t *testing.T) {
		e := &ClusterE2ETest{
			T:                   t,
			ClusterConfigFolder: "test",
			ClusterName:         "test",
			KubectlClient:       executables.NewKubectl(executable),
		}

		executable.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
		e.ValidatePackageBundleControllerRegistry()
	})
}
