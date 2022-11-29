package validator_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/features"
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
	ipValidator := validator.NewIPValidator(validator.CustomNetClient(client))

	g.Expect(ipValidator.ValidateControlPlaneIPUniqueness(cluster)).To(Succeed())
}

func TestValidateSupportedProvider(t *testing.T) {
	tests := []struct {
		name     string
		wantErr  error
		provider string
		envVars  map[string]string
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
			envVars:  map[string]string{"SNOW_PROVIDER": ""},
		},
		{
			name:     "SuccessSupportedVSphereProvider",
			wantErr:  nil,
			provider: constants.VSphereProviderName,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			features.ClearCache()
			for evName, evValue := range tc.envVars {
				tt.Setenv(evName, evValue)
			}
			mockCtrl := gomock.NewController(tt)
			provider := mockprovider.NewMockProvider(mockCtrl)
			provider.EXPECT().Name().Return(tc.provider).AnyTimes()
			err := validator.ValidateSupportedProviderCreate(provider)
			if !reflect.DeepEqual(err, tc.wantErr) {
				tt.Errorf("%s got = %v, \nwant %v", tc.name, err, tc.wantErr)
			}
		})
	}
}
