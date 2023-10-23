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

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	mockCrypto "github.com/aws/eks-anywhere/pkg/crypto/mocks"
	mocknutanix "github.com/aws/eks-anywhere/pkg/providers/nutanix/mocks"
	"github.com/aws/eks-anywhere/pkg/utils/ptr"
)

//go:embed testdata/datacenterConfig_with_trust_bundle.yaml
var nutanixDatacenterConfigSpecWithTrustBundle string

//go:embed testdata/datacenterConfig_with_invalid_port.yaml
var nutanixDatacenterConfigSpecWithInvalidPort string

//go:embed testdata/datacenterConfig_with_invalid_endpoint.yaml
var nutanixDatacenterConfigSpecWithInvalidEndpoint string

//go:embed testdata/datacenterConfig_with_insecure.yaml
var nutanixDatacenterConfigSpecWithInsecure string

//go:embed testdata/datacenterConfig_no_credentialRef.yaml
var nutanixDatacenterConfigSpecWithNoCredentialRef string

//go:embed testdata/datacenterConfig_invalid_credentialRef_kind.yaml
var nutanixDatacenterConfigSpecWithInvalidCredentialRefKind string

//go:embed testdata/datacenterConfig_empty_credentialRef_name.yaml
var nutanixDatacenterConfigSpecWithEmptyCredentialRefName string

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

func fakeProjectList() *v3.ProjectListResponse {
	return &v3.ProjectListResponse{
		Entities: []*v3.Project{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("5c9a0641-1025-40ed-9e1d-0d0a23043e57"),
				},
				Spec: &v3.ProjectSpec{
					Name: "prism-image",
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
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "vCPU sockets 0 must be greater than or equal to 1",
		},
		{
			name: "invalid vcpus per socket",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.VCPUsPerSocket = 0
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "vCPUs per socket 0 must be greater than or equal to 1",
		},
		{
			name: "memory size less than min required",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.MemorySize = resource.MustParse("100Mi")
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "MemorySize must be greater than or equal to 2048Mi",
		},
		{
			name: "invalid system size",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.SystemDiskSize = resource.MustParse("100Mi")
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "SystemDiskSize must be greater than or equal to 20Gi",
		},
		{
			name: "empty cluster name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Name = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing cluster name",
		},
		{
			name: "empty cluster uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Type = anywherev1.NutanixIdentifierUUID
				machineConf.Spec.Cluster.UUID = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing cluster uuid",
		},
		{
			name: "invalid cluster identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.Cluster.Type = "notanidentifier"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid cluster identifier type",
		},
		{
			name: "list cluster request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(nil, errors.New("cluster not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find cluster by name",
		},
		{
			name: "list cluster request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(&v3.ClusterListIntentResponse{}, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find cluster by name",
		},
		{
			name: "duplicate clusters found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				clusters := fakeClusterList()
				clusters.Entities = append(clusters.Entities, clusters.Entities[0])
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) cluster with name",
		},
		{
			name: "empty subnet name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Name = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing subnet name",
		},
		{
			name: "empty subnet uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Type = anywherev1.NutanixIdentifierUUID
				machineConf.Spec.Subnet.UUID = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing subnet uuid",
		},
		{
			name: "invalid subnet identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				machineConf.Spec.Subnet.Type = "notanidentifier"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid subnet identifier type",
		},
		{
			name: "list subnet request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(nil, errors.New("subnet not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find subnet by name",
		},
		{
			name: "list subnet request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(&v3.SubnetListIntentResponse{}, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
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
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) subnet with name",
		},
		{
			name: "empty image name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Name = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
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
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing image uuid",
		},
		{
			name: "invalid image identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Type = "notanidentifier"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid image identifier type",
		},
		{
			name: "list image request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(nil, errors.New("image not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find image by name",
		},
		{
			name: "list image request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(&v3.ImageListIntentResponse{}, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
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
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
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
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "",
		},
		{
			name: "empty project name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: nil,
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing project name",
		},
		{
			name: "empty project uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: nil,
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing project uuid",
		},
		{
			name: "invalid project identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: "notatype",
					UUID: nil,
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid project identifier type",
		},
		{
			name: "list project request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				mockClient.EXPECT().ListProject(gomock.Any(), gomock.Any()).Return(nil, errors.New("project not found"))
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: ptr.String("notaproject"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find project by name",
		},
		{
			name: "list project request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				mockClient.EXPECT().ListProject(gomock.Any(), gomock.Any()).Return(&v3.ProjectListResponse{}, nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: ptr.String("notaproject"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find project by name",
		},
		{
			name: "duplicate project found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				projects := fakeProjectList()
				projects.Entities = append(projects.Entities, projects.Entities[0])
				mockClient.EXPECT().ListProject(gomock.Any(), gomock.Any()).Return(projects, nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: ptr.String("project"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) project with name",
		},
		{
			name: "empty category key",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.AdditionalCategories = []anywherev1.NutanixCategoryIdentifier{
					{
						Key:   "",
						Value: "",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing category key",
		},
		{
			name: "empty category value",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.AdditionalCategories = []anywherev1.NutanixCategoryIdentifier{
					{
						Key:   "key",
						Value: "",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing category value",
		},
		{
			name: "get category key failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				mockClient.EXPECT().GetCategoryKey(gomock.Any(), gomock.Any()).Return(nil, errors.New("category key not found"))
				machineConf.Spec.AdditionalCategories = []anywherev1.NutanixCategoryIdentifier{
					{
						Key:   "nonexistent",
						Value: "value",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find category with key",
		},
		{
			name: "get category value failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil)
				mockClient.EXPECT().ListSubnet(gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				categoryKey := v3.CategoryKeyStatus{
					Name: ptr.String("key"),
				}
				mockClient.EXPECT().GetCategoryKey(gomock.Any(), gomock.Any()).Return(&categoryKey, nil)
				mockClient.EXPECT().GetCategoryValue(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("category value not found"))
				machineConf.Spec.AdditionalCategories = []anywherev1.NutanixCategoryIdentifier{
					{
						Key:   "key",
						Value: "nonexistent",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find category value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			machineConfig := &anywherev1.NutanixMachineConfig{}
			err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
			require.NoError(t, err)

			mockClient := mocknutanix.NewMockClient(ctrl)
			validator := tc.setup(machineConfig, mockClient, mockCrypto.NewMockTlsValidator(ctrl), mocknutanix.NewMockRoundTripper(ctrl))
			err = validator.ValidateMachineConfig(context.Background(), mockClient, machineConfig)
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
		{
			name:       "nil credentialRef",
			dcConfFile: nutanixDatacenterConfigSpecWithNoCredentialRef,
			expectErr:  true,
		},
		{
			name:       "invalid credentialRef kind",
			dcConfFile: nutanixDatacenterConfigSpecWithInvalidCredentialRefKind,
			expectErr:  true,
		},
		{
			name:       "empty credentialRef name",
			dcConfFile: nutanixDatacenterConfigSpecWithEmptyCredentialRefName,
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
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)
	require.NotNil(t, validator)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dcConf := &anywherev1.NutanixDatacenterConfig{}
			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), clientCache.clients["test"], dcConf)
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
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)
	require.NotNil(t, validator)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dcConf := &anywherev1.NutanixDatacenterConfig{}
			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), clientCache.clients["test"], dcConf)
			if tc.expectErr {
				assert.Error(t, err, tc.name)
			} else {
				assert.NoError(t, err, tc.name)
			}
		})
	}
}

func TestValidateClusterMachineConfigsError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	clusterSpec.Cluster.Spec.KubernetesVersion = "1.22"

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsCPNotFoundError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	clusterSpec.Cluster.Spec.ControlPlaneConfiguration.MachineGroupRef.Name = "invalid-cp-name"

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsEtcdNotFoundError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration.MachineGroupRef.Name = "invalid-etcd-name"

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsCPError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	clusterSpec.NutanixMachineConfigs["eksa-unit-test-cp"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsEtcdError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	clusterSpec.NutanixMachineConfigs["eksa-unit-test"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsModularUpgradeError(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)
	kube122 := v1alpha1.KubernetesVersion("1.22")
	clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[0].KubernetesVersion = &kube122

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateClusterMachineConfigsSuccess(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multiple-machineconfigs.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)

	clusterSpec.Cluster.Spec.KubernetesVersion = "1.22"
	clusterSpec.NutanixMachineConfigs["eksa-unit-test-cp"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")
	clusterSpec.NutanixMachineConfigs["eksa-unit-test"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")
	clusterSpec.NutanixMachineConfigs["eksa-unit-test-md-1"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")
	clusterSpec.NutanixMachineConfigs["eksa-unit-test-md-2"].Spec.Image.Name = utils.StringPtr("kubernetes_1_22")

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	err := validator.checkImageNameMatchesKubernetesVersion(ctx, clusterSpec, clientCache.clients["test"])
	if err != nil {
		t.Fatalf("validation should pass: %v", err)
	}
}
