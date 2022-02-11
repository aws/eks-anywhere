package templates_test

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/internal/templates"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/internal/templates/mocks"
)

type test struct {
	t           *testing.T
	network     string
	domain      string
	zone        string
	account     string
	cloudmonkey *mocks.MockCloudMonkeyClient
	factory     *templates.Factory
	ctx         context.Context
	dummyError  error
}

type createTest struct {
	*test
	domain        string
	zone          string
	account       string
	machineConfig *v1alpha1.CloudStackMachineConfig
}

func newTest(t *testing.T) *test {
	ctrl := gomock.NewController(t)
	test := &test{
		t:           t,
		domain:      "domain1",
		zone:        "zone1",
		account:     "admin",
		cloudmonkey: mocks.NewMockCloudMonkeyClient(ctrl),
		ctx:         context.Background(),
		dummyError:  errors.New("error from cloudmonkey"),
	}
	f := templates.NewFactory(
		test.cloudmonkey,
		test.network,
		test.domain,
		test.zone,
		test.account,
	)
	test.factory = f
	return test
}

func newMachineConfig(t *testing.T) *v1alpha1.CloudStackMachineConfig {
	return &v1alpha1.CloudStackMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind: v1alpha1.CloudStackMachineConfigKind,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "eksa-unit-test",
		},
		Spec: v1alpha1.CloudStackMachineConfigSpec{
			Template:        "centos7-k8s-118",
			ComputeOffering: "m4-large",
			DiskOffering:    "ssd-100GB",
			OSFamily:        v1alpha1.Ubuntu,
			Users: []v1alpha1.UserConfiguration{{
				Name:              "mySshUsername",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			}},
			Details: map[string]string{
				"foo": "bar",
				"key": "value",
			},
		},
	}
}

func newCreateTest(t *testing.T) *createTest {
	test := newTest(t)
	return &createTest{
		test:          test,
		domain:        "domain1",
		zone:          "zone1",
		account:       "admin",
		machineConfig: newMachineConfig(t),
	}
}

func (ct *createTest) validateMachineResources() error {
	return ct.factory.ValidateMachineResources(ct.ctx, ct.machineConfig)
}
