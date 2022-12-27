package nutanix

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	mockCrypto "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	mocknutanix "github.com/aws/eks-anywhere/pkg/providers/nutanix/mocks"
)

//go:embed testdata/datacenterConfig_with_trust_bundle.yaml
var nutanixDatacenterConfigSpecWithTrustBundle string

//go:embed testdata/datacenterConfig_with_invalid_port.yaml
var nutanixDatacenterConfigSpecWithInvalidPort string

//go:embed testdata/datacenterConfig_with_invalid_endpoint.yaml
var nutanixDatacenterConfigSpecWithInvalidEndpoint string

//go:embed testdata/datacenterConfig_with_insecure.yaml
var nutanixDatacenterConfigSpecWithInsecure string

func fakeClusterList() *v3.ClusterListIntentResponse {
	return &v3.ClusterListIntentResponse{
		Entities: []*v3.ClusterIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
		},
	}
}

func fakeSubnetList() *v3.SubnetListIntentResponse {
	return &v3.SubnetListIntentResponse{
		Entities: []*v3.SubnetIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("b15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				},
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet"),
				},
			},
		},
	}
}

func fakeImageList() *v3.ImageListIntentResponse {
	return &v3.ImageListIntentResponse{
		Entities: []*v3.ImageIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("c15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				},
				Spec: &v3.Image{
					Name: utils.StringPtr("prism-image"),
				},
			},
		},
	}
}

func TestNutanixValidatorValidateMachineConfig(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name          string
		setup         func(*anywherev1.NutanixMachineConfig, *mocknutanix.MockClient, *mockCrypto.MockTlsValidator, *mocknutanix.MockRoundTripper) *Validator
		expectedError string
	}{
		{
			name: "invalid vcpu sockets",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.VCPUSockets = 0
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "vCPU sockets 0 must be greater than or equal to 1",
		},
		{
			name: "invalid vcpus per socket",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.VCPUsPerSocket = 0
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "vCPUs per socket 0 must be greater than or equal to 1",
		},
		{
			name: "memory size less than min required",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.MemorySize = resource.MustParse("100Mi")
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "MemorySize must be greater than or equal to 2048Mi",
		},
		{
			name: "invalid system size",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.SystemDiskSize = resource.MustParse("100Mi")
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "SystemDiskSize must be greater than or equal to 20Gi",
		},
		{
			name: "empty cluster name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Name = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing cluster name",
		},
		{
			name: "empty cluster uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Type = anywherev1.NutanixIdentifierUUID
				machineConf.Spec.Cluster.UUID = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing cluster uuid",
		},
		{
			name: "invalid cluster identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Type = "notanidentifier"
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid cluster identifier type",
		},
		{
			name: "list cluster request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(nil, errors.New("cluster not found"))
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find cluster by name",
		},
		{
			name: "list cluster request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(&v3.ClusterListIntentResponse{}, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find cluster by name",
		},
		{
			name: "duplicate clusters found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				clusters := fakeClusterList()
				clusters.Entities = append(clusters.Entities, clusters.Entities[0])
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) cluster with name",
		},
		{
			name: "empty subnet name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Name = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing subnet name",
		},
		{
			name: "empty subnet uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Type = anywherev1.NutanixIdentifierUUID
				machineConf.Spec.Subnet.UUID = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing subnet uuid",
		},
		{
			name: "invalid subnet identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Type = "notanidentifier"
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid subnet identifier type",
		},
		{
			name: "list subnet request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(nil, errors.New("subnet not found"))
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find subnet by name",
		},
		{
			name: "list subnet request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(&v3.SubnetListIntentResponse{}, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find subnet by name",
		},
		{
			name: "duplicate subnets found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				subnets := fakeSubnetList()
				subnets.Entities = append(subnets.Entities, subnets.Entities[0])
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(subnets, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) subnet with name",
		},
		{
			name: "empty image name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Name = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing image name",
		},
		{
			name: "empty image uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Type = anywherev1.NutanixIdentifierUUID
				machineConf.Spec.Image.UUID = nil
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing image uuid",
		},
		{
			name: "invalid image identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Type = "notanidentifier"
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid image identifier type",
		},
		{
			name: "list image request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(nil, errors.New("image not found"))
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find image by name",
		},
		{
			name: "list image request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(&v3.ImageListIntentResponse{}, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find image by name",
		},
		{
			name: "duplicate image found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				images := fakeImageList()
				images.Entities = append(images.Entities, images.Entities[0])
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(images, nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) image with name",
		},
		{
			name: "filters out prism central clusters",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				clusters := fakeClusterList()
				tmp, err := json.Marshal(clusters.Entities[0])
				assert.NoError(t, err)
				var cluster v3.ClusterIntentResponse
				err = json.Unmarshal(tmp, &cluster)
				assert.NoError(t, err)
				cluster.Status.Resources.Config.ServiceList = []*string{utils.StringPtr("PRISM_CENTRAL")}
				clusters.Entities = append(clusters.Entities, &cluster)
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				return NewValidator(mockClient, validator, &http.Client{Transport: transport})
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			machineConfig := &anywherev1.NutanixMachineConfig{}
			err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
			require.NoError(t, err)

			mockClient := mocknutanix.NewMockClient(ctrl)
			validator := tc.setup(machineConfig, mockClient, mockCrypto.NewMockTlsValidator(ctrl), mocknutanix.NewMockRoundTripper(ctrl))
			err = validator.ValidateMachineConfig(context.Background(), machineConfig)
			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNutanixValidatorValidateDatacenterConfig(t *testing.T) {
	tests := []struct {
		name       string
		dcConfFile string
		expectErr  bool
	}{
		{
			name:       "valid datacenter config without trust bundle",
			dcConfFile: nutanixDatacenterConfigSpec,
		},
		{
			name:       "valid datacenter config with trust bundle",
			dcConfFile: nutanixDatacenterConfigSpecWithTrustBundle,
		},
		{
			name:       "valid datacenter config with insecure",
			dcConfFile: nutanixDatacenterConfigSpecWithInsecure,
		},
		{
			name:       "valid datacenter config with invalid port",
			dcConfFile: nutanixDatacenterConfigSpecWithInvalidPort,
			expectErr:  true,
		},
		{
			name:       "valid datacenter config with invalid endpoint",
			dcConfFile: nutanixDatacenterConfigSpecWithInvalidEndpoint,
			expectErr:  true,
		},
	}

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	validator := NewValidator(mockClient, mockTLSValidator, mockHTTPClient)
	require.NotNil(t, validator)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dcConf := &anywherev1.NutanixDatacenterConfig{}
			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), dcConf)
			if tc.expectErr {
				assert.Error(t, err, tc.name)
			} else {
				assert.NoError(t, err, tc.name)
			}
		})
	}
}

func TestNutanixValidatorValidateDatacenterConfigWithInvalidCreds(t *testing.T) {
	tests := []struct {
		name       string
		dcConfFile string
		expectErr  bool
	}{
		{
			name:       "valid datacenter config without trust bundle",
			dcConfFile: nutanixDatacenterConfigSpec,
			expectErr:  true,
		},
	}

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, errors.New("GetCurrentLoggedInUser returned error")).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	validator := NewValidator(mockClient, mockTLSValidator, mockHTTPClient)
	require.NotNil(t, validator)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dcConf := &anywherev1.NutanixDatacenterConfig{}
			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), dcConf)
			if tc.expectErr {
				assert.Error(t, err, tc.name)
			} else {
				assert.NoError(t, err, tc.name)
			}
		})
	}
}
