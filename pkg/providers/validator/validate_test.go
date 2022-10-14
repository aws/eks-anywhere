package validator_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/networkutils/mocks"
	mockprovider "github.com/aws/eks-anywhere/pkg/providers/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/validator"
)

func TestValidateControlPlaneIpUniqueness(t *testing.T) {
	g := NewWithT(t)
	cluster := &v1alpha1.Cluster{
		Spec: v1alpha1.ClusterSpec{
			ControlPlaneConfiguration: v1alpha1.ControlPlaneConfiguration{
				Endpoint: &v1alpha1.Endpoint{
					Host: "1.2.3.4",
				},
			},
		},
	}
	ctrl := gomock.NewController(t)
	client := mocks.NewMockNetClient(ctrl)
	client.EXPECT().DialTimeout(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(5).
		Return(nil, errors.New("no connection"))
	g.Expect(validator.ValidateControlPlaneIpUniqueness(cluster, client)).To(Succeed())
}

func TestValidateSupportedProvider(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  error
		provider string
	}{
		{
			name:     "SuccessSupportedCloudstackProvider",
			wantErr:  nil,
			provider: constants.CloudStackProviderName,
		},
		{
			name:     "FailureUnsupportedSnowProvider",
			wantErr:  errors.New("provider snow is not supported in this release"),
			provider: constants.SnowProviderName,
		},
		{
			name:     "SuccessSupportedVSphereProvider",
			wantErr:  nil,
			provider: constants.VSphereProviderName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			mockCtrl := gomock.NewController(t)
			provider := mockprovider.NewMockProvider(mockCtrl)
			provider.EXPECT().Name().Return(tc.provider).AnyTimes()
			err := validator.ValidateSupportedProviderCreate(provider)
			if !reflect.DeepEqual(err, tc.wantErr) {
				t.Errorf("%v got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
