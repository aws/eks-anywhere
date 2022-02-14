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
	network     v1alpha1.CloudStackResourceRef
	domain      string
	zone        v1alpha1.CloudStackResourceRef
	account     string
	cloudmonkey *mocks.MockCloudMonkeyClient
	factory     *templates.Factory
	ctx         context.Context
	dummyError  error
}

type createTest struct {
	*test
	domain        string
	zone          v1alpha1.CloudStackResourceRef
	account       string
	machineConfig *v1alpha1.CloudStackMachineConfig
}

func newTest(t *testing.T) *test {
	ctrl := gomock.NewController(t)
	test := &test{
		t:           t,
		domain:      "domain1",
		zone:        v1alpha1.CloudStackResourceRef{
			Type:  "name",
			Value: "zone1",
		},
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
			Template:        v1alpha1.CloudStackResourceRef{
				Value: "centos7-k8s-118",
				Type: "name",
			},
			ComputeOffering: v1alpha1.CloudStackResourceRef{
				Value: "m4-large",
				Type: "name",
			},
			Users: []v1alpha1.UserConfiguration{{
				Name:              "mySshUsername",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			}},
		},
	}
}

func newCreateTest(t *testing.T) *createTest {
	test := newTest(t)
	return &createTest{
		test:          test,
		domain:        "domain1",
		zone:          v1alpha1.CloudStackResourceRef{
			Value: "zone1",
			Type: "name",
		},
		account:       "admin",
		machineConfig: newMachineConfig(t),
	}
}

func (ct *createTest) validateMachineResources() error {
	return ct.factory.ValidateMachineResources(ct.ctx, ct.machineConfig)
}
