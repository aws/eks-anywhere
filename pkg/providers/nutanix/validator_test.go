package nutanix

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/utils"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/constants"
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

//go:embed testdata/datacenterConfig_no_credentialRef.yaml
var nutanixDatacenterConfigSpecWithNoCredentialRef string

//go:embed testdata/datacenterConfig_invalid_credentialRef_kind.yaml
var nutanixDatacenterConfigSpecWithInvalidCredentialRefKind string

//go:embed testdata/datacenterConfig_empty_credentialRef_name.yaml
var nutanixDatacenterConfigSpecWithEmptyCredentialRefName string

//go:embed testdata/datacenterConfig_with_failure_domains.yaml
var nutanixDatacenterConfigSpecWithFailureDomain string

//go:embed testdata/datacenterConfig_with_failure_domains_invalid_name.yaml
var nutanixDatacenterConfigSpecWithFailureDomainInvalidName string

//go:embed testdata/datacenterConfig_with_failure_domains_invalid_cluster.yaml
var nutanixDatacenterConfigSpecWithFailureDomainInvalidCluster string

//go:embed testdata/datacenterConfig_with_failure_domains_invalid_subnet.yaml
var nutanixDatacenterConfigSpecWithFailureDomainInvalidSubnet string

//go:embed testdata/datacenterConfig_with_failure_domains_invalid_wg.yaml
var nutanixDatacenterConfigSpecWithFailureDomainInvalidWorkerMachineGroups string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPs string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips_invalid_cidr.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidCIDR string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips_invalid_ip.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIP string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips_invalid_ip_range1.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange1 string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips_invalid_ip_range2.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange2 string

//go:embed testdata/datacenterConfig_with_ccm_exclude_node_ips_invalid_ip_range3.yaml
var nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange3 string

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
					ClusterReference: &v3.Reference{
						Kind: utils.StringPtr("cluster"),
						UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
						Name: utils.StringPtr("prism-cluster"),
					},
				},
			},
		},
	}
}

func fakeClusterListForDCTest(filter string) (*v3.ClusterListIntentResponse, error) {
	data := &v3.ClusterListIntentResponse{
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
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("4d69ca7d-022f-49d1-a454-74535993bda4"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster-1"),
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

	result := &v3.ClusterListIntentResponse{
		Entities: []*v3.ClusterIntentResponse{},
	}

	if filter != "" {
		str := strings.Replace(filter, "name==", "", -1)
		for _, cluster := range data.Entities {
			if str == *cluster.Spec.Name {
				result.Entities = append(result.Entities, cluster)
			}
		}
	} else {
		result.Entities = data.Entities
	}

	return result, nil
}

func fakeSubnetListForDCTest(filter string) (*v3.SubnetListIntentResponse, error) {
	data := &v3.SubnetListIntentResponse{
		Entities: []*v3.SubnetIntentResponse{
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("2d166190-7759-4dc6-b835-923262d6b497"),
				},
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet"),
					ClusterReference: &v3.Reference{
						Kind: utils.StringPtr("cluster"),
						UUID: utils.StringPtr("4d69ca7d-022f-49d1-a454-74535993bda4"),
						Name: utils.StringPtr("prism-cluster-1"),
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("2d166190-7759-4dc6-b835-923262d6b497"),
				},
				Spec: &v3.Subnet{
					Name: utils.StringPtr("prism-subnet-1"),
					ClusterReference: &v3.Reference{
						Kind: utils.StringPtr("cluster"),
						UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
						Name: utils.StringPtr("prism-cluster"),
					},
				},
			},
		},
	}

	result := &v3.SubnetListIntentResponse{
		Entities: []*v3.SubnetIntentResponse{},
	}

	if filter != "" {
		filters := strings.Split(filter, ";")
		str := strings.Replace(filters[0], "name==", "", -1)
		for _, subnet := range data.Entities {
			if str == *subnet.Spec.Name {
				result.Entities = append(result.Entities, subnet)
			}
		}
	} else {
		result.Entities = data.Entities
	}

	return result, nil
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
					Name: "project",
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
			name: "invald bootType",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				machineConf.Spec.BootType = "abc"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "boot type abc is not supported, only legacy and uefi are supported",
		},
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(nil, errors.New("cluster not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to list clusters",
		},
		{
			name: "list cluster request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(&v3.ClusterListIntentResponse{}, nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(clusters, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) cluster with name",
		},
		{
			name: "empty subnet name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				machineConf.Spec.Subnet.Name = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing subnet name",
		},
		{
			name: "empty subnet uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				machineConf.Spec.Subnet.Type = "notanidentifier"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid subnet identifier type",
		},
		{
			name: "list subnet request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("subnet not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to list subnets",
		},
		{
			name: "list subnet request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(&v3.SubnetListIntentResponse{}, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find subnet by name",
		},
		{
			name: "duplicate subnets found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				subnets := fakeSubnetList()
				subnets.Entities = append(subnets.Entities, subnets.Entities[0])
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(subnets, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) subnet with name",
		},
		{
			name: "empty image name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Name = nil
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing image name",
		},
		{
			name: "empty image uuid",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				machineConf.Spec.Image.Type = "notanidentifier"
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid image identifier type",
		},
		{
			name: "list image request failed",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(nil, errors.New("image not found"))
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to list images",
		},
		{
			name: "list image request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(&v3.ImageListIntentResponse{}, nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find image by name",
		},
		{
			name: "duplicate image found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				images := fakeImageList()
				images.Entities = append(images.Entities, images.Entities[0])
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(images, nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(clusters, nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "",
		},
		{
			name: "empty project name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				mockClient.EXPECT().ListAllProject(gomock.Any(), gomock.Any()).Return(nil, errors.New("project not found"))
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: utils.StringPtr("notaproject"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to list projects",
		},
		{
			name: "list project request did not find match",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				mockClient.EXPECT().ListAllProject(gomock.Any(), gomock.Any()).Return(&v3.ProjectListResponse{}, nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: utils.StringPtr("notaproject"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "failed to find project by name",
		},
		{
			name: "duplicate project found",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				projects := fakeProjectList()
				projects.Entities = append(projects.Entities, projects.Entities[0])
				mockClient.EXPECT().ListAllProject(gomock.Any(), gomock.Any()).Return(projects, nil)
				machineConf.Spec.Project = &anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierName,
					Name: utils.StringPtr("project"),
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "found more than one (2) project with name",
		},
		{
			name: "empty category key",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
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
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				categoryKey := v3.CategoryKeyStatus{
					Name: utils.StringPtr("key"),
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
		{
			name: "invalid gpu identifier type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "invalid",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "invalid GPU identifier type",
		},
		{
			name: "missing GPU type",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing GPU type",
		},
		{
			name: "missing GPU device ID",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "deviceID",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing GPU device ID",
		},
		{
			name: "missing GPU name",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "missing GPU name",
		},
		{
			name: "machine config is not associated with any worker node group",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Name = "test-wn"
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "NVIDIA A40-1Q",
					},
				}
				return NewValidator(&ClientCache{}, validator, &http.Client{Transport: transport})
			},
			expectedError: "not associated with any worker node group",
		},
		{
			name: "GPUs are not supported for control plane machine",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Name = "test-cp"
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "NVIDIA A40-1Q",
					},
				}
				return NewValidator(&ClientCache{}, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPUs are not supported for control plane machine",
		},
		{
			name: "GPUs are not supported for external etcd machine",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Name = "test-etcd"
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "NVIDIA A40-1Q",
					},
				}
				return NewValidator(&ClientCache{}, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPUs are not supported for external etcd machine",
		},
		{
			name: "validation pass",
			setup: func(machineConf *anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterList(), nil).Times(2)
				mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).Return(fakeSubnetList(), nil)
				mockClient.EXPECT().ListAllImage(gomock.Any(), gomock.Any()).Return(fakeImageList(), nil)
				machineConf.Name = "eksa-unit-test"
				machineConf.Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "NVIDIA A40-1Q",
					},
				}
				return NewValidator(&ClientCache{}, validator, &http.Client{Transport: transport})
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cluster := &anywherev1.Cluster{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: anywherev1.ClusterSpec{
					ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
						MachineGroupRef: &anywherev1.Ref{
							Kind: "NutanixMachineConfig",
							Name: "test-cp",
						},
					},
					ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
						MachineGroupRef: &anywherev1.Ref{
							Kind: "NutanixMachineConfig",
							Name: "test-etcd",
						},
					},
					WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
						{
							MachineGroupRef: &anywherev1.Ref{
								Kind: "NutanixMachineConfig",
								Name: "eksa-unit-test",
							},
						},
					},
				},
			}
			machineConfig := &anywherev1.NutanixMachineConfig{}
			err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfig)
			require.NoError(t, err)

			mockClient := mocknutanix.NewMockClient(ctrl)
			validator := tc.setup(machineConfig, mockClient, mockCrypto.NewMockTlsValidator(ctrl), mocknutanix.NewMockRoundTripper(ctrl))
			err = validator.ValidateMachineConfig(context.Background(), mockClient, cluster, machineConfig)
			if tc.expectedError != "" {
				assert.Contains(t, err.Error(), tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func fakeHostList() *v3.HostListResponse {
	return &v3.HostListResponse{
		Entities: []*v3.HostResponse{
			{
				Status: &v3.HostStatus{
					ClusterReference: &v3.ReferenceValues{
						UUID: "a15f6966-bfc7-4d1e-8575-224096fc1cdb",
					},
					Resources: &v3.HostResources{
						GPUList: []*v3.GPU{
							{
								Assignable: false,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: false,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(557),
								Name:       "NVIDIA A40-1Q",
								Mode:       "VIRTUAL",
							},
						},
					},
				},
			},
			{
				Status: &v3.HostStatus{
					ClusterReference: &v3.ReferenceValues{
						UUID: "4d69ca7d-022f-49d1-a454-74535993bda4",
					},
					Resources: &v3.HostResources{
						GPUList: []*v3.GPU{
							{
								Assignable: false,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: false,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
							{
								Assignable: true,
								DeviceID:   utils.Int64Ptr(8757),
								Name:       "Ampere 40",
								Mode:       "PASSTHROUGH_COMPUTE",
							},
						},
					},
				},
			},
			{
				Status: &v3.HostStatus{
					ClusterReference: &v3.ReferenceValues{
						UUID: "e0b1dfc7-5447-410f-b708-f9603e9be79a",
					},
					Resources: &v3.HostResources{},
				},
			},
		},
	}
}

func fakeEmptyHostList() *v3.HostListResponse {
	return &v3.HostListResponse{
		Entities: []*v3.HostResponse{},
	}
}

func fakeClusterListForFreeGPUTest() *v3.ClusterListIntentResponse {
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
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("4d69ca7d-022f-49d1-a454-74535993bda4"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster-1"),
				},
				Status: &v3.ClusterDefStatus{
					Resources: &v3.ClusterObj{
						Config: &v3.ClusterConfig{
							ServiceList: []*string{utils.StringPtr("AOS")},
						},
					},
				},
			},
			{
				Metadata: &v3.Metadata{
					UUID: utils.StringPtr("e0b1dfc7-5447-410f-b708-f9603e9be79a"),
				},
				Spec: &v3.Cluster{
					Name: utils.StringPtr("prism-cluster-2"),
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

func TestNutanixValidatorValidateFreeGPU(t *testing.T) {
	ctrl := gomock.NewController(t)

	tests := []struct {
		name          string
		setup         func(map[string]*anywherev1.NutanixMachineConfig, *mocknutanix.MockClient, *mockCrypto.MockTlsValidator, *mocknutanix.MockRoundTripper) *Validator
		expectedError string
	}{
		{
			name: "not enough GPU resources available by name",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["cp"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["cp"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
				}
				machineConfigs["worker"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPU with name Ampere 40 not found",
		},
		{
			name: "not enough GPU resources available by name in different PE (UUID)",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["cp"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("e0b1dfc7-5447-410f-b708-f9603e9be79a"),
				}
				machineConfigs["cp"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{}
				machineConfigs["worker"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPU with name Ampere 40 not found on cluster with UUID a15f6966-bfc7-4d1e-8575-224096fc1cdb",
		},
		{
			name: "not enough GPU resources available by deviceID in different PE (UUID)",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["cp"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("e0b1dfc7-5447-410f-b708-f9603e9be79a"),
				}
				machineConfigs["cp"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{}
				machineConfigs["worker"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPU with device ID 8757 not found on cluster with UUID a15f6966-bfc7-4d1e-8575-224096fc1cdb",
		},
		{
			name: "not enough GPU resources available by deviceID",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["cp"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["cp"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
				}
				machineConfigs["worker"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("a15f6966-bfc7-4d1e-8575-224096fc1cdb"),
				}
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "GPU with device ID 8757 not found",
		},
		{
			name: "no GPU resources found",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeEmptyHostList(), nil)
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "no GPUs found",
		},
		{
			name: "no GPU resources found: ListAllHost failed",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(nil, fmt.Errorf("failed to list hosts"))
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
					{
						Type: "name",
						Name: "Ampere 40",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "no GPUs found",
		},
		{
			name: "mixed passthrough and vGPU mode GPUs in a machine config",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type: "name",
						Name: "NVIDIA A40-1Q",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "all GPUs in a machine config must be of the same mode, vGPU or passthrough",
		},
		{
			name: "GPUs validation successful",
			setup: func(machineConfigs map[string]*anywherev1.NutanixMachineConfig, mockClient *mocknutanix.MockClient, validator *mockCrypto.MockTlsValidator, transport *mocknutanix.MockRoundTripper) *Validator {
				mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).Return(fakeClusterListForFreeGPUTest(), nil).AnyTimes()
				mockClient.EXPECT().ListAllHost(gomock.Any()).Return(fakeHostList(), nil)
				machineConfigs["cp"].Spec.Cluster = anywherev1.NutanixResourceIdentifier{
					Type: anywherev1.NutanixIdentifierUUID,
					UUID: utils.StringPtr("4d69ca7d-022f-49d1-a454-74535993bda4"),
				}
				machineConfigs["cp"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
					{
						Type:     "deviceID",
						DeviceID: utils.Int64Ptr(8757),
					},
				}

				machineConfigs["worker"].Spec.GPUs = []anywherev1.NutanixGPUIdentifier{
					{
						Type: "name",
						Name: "Ampere 40",
					},
				}
				clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
				return NewValidator(clientCache, validator, &http.Client{Transport: transport})
			},
			expectedError: "",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			machineConfigsNames := []string{"cp", "etcd", "worker"}
			machineConfigs := make(map[string]*anywherev1.NutanixMachineConfig)

			for _, name := range machineConfigsNames {
				machineConfigs[name] = &anywherev1.NutanixMachineConfig{}
				err := yaml.Unmarshal([]byte(nutanixMachineConfigSpec), machineConfigs[name])
				machineConfigs[name].Name = machineConfigs[name].Name + "-" + name
				require.NoError(t, err)
			}

			mockClient := mocknutanix.NewMockClient(ctrl)
			validator := tc.setup(machineConfigs, mockClient, mockCrypto.NewMockTlsValidator(ctrl), mocknutanix.NewMockRoundTripper(ctrl))
			spec := &cluster.Spec{
				Config: &cluster.Config{
					Cluster: &anywherev1.Cluster{
						Spec: anywherev1.ClusterSpec{
							ControlPlaneConfiguration: anywherev1.ControlPlaneConfiguration{
								Count: 1,
								MachineGroupRef: &anywherev1.Ref{
									Name: "eksa-unit-test-cp",
									Kind: constants.NutanixMachineConfigKind,
								},
							},
							WorkerNodeGroupConfigurations: []anywherev1.WorkerNodeGroupConfiguration{
								{
									Count: utils.IntPtr(2),
									MachineGroupRef: &anywherev1.Ref{
										Name: "eksa-unit-test-worker",
										Kind: constants.NutanixMachineConfigKind,
									},
								},
							},
							ExternalEtcdConfiguration: &anywherev1.ExternalEtcdConfiguration{
								Count: 1,
								MachineGroupRef: &anywherev1.Ref{
									Name: "eksa-unit-test-etcd",
									Kind: constants.NutanixMachineConfigKind,
								},
							},
						},
					},
					NutanixMachineConfigs: machineConfigs,
				},
			}
			err := validator.validateFreeGPU(context.Background(), mockClient, spec)
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
		{
			name:       "valid failure domains",
			dcConfFile: nutanixDatacenterConfigSpecWithFailureDomain,
			expectErr:  false,
		},
		{
			name:       "failure domain with invalid name",
			dcConfFile: nutanixDatacenterConfigSpecWithFailureDomainInvalidName,
			expectErr:  true,
		},
		{
			name:       "failure domain with invalid cluster",
			dcConfFile: nutanixDatacenterConfigSpecWithFailureDomainInvalidCluster,
			expectErr:  true,
		},
		{
			name:       "failure domains with invalid subnet",
			dcConfFile: nutanixDatacenterConfigSpecWithFailureDomainInvalidSubnet,
			expectErr:  true,
		},
		{
			name:       "failure domains with invalid workerMachineGroups",
			dcConfFile: nutanixDatacenterConfigSpecWithFailureDomainInvalidWorkerMachineGroups,
			expectErr:  true,
		},
		{
			name:       "valid ccmExcludeNodeIPs",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPs,
			expectErr:  false,
		},
		{
			name:       "ccmExcludeNodeIPs invalid CIDR",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidCIDR,
			expectErr:  true,
		},
		{
			name:       "ccmExcludeNodeIPs invalid IP",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIP,
			expectErr:  true,
		},
		{
			name:       "ccmExcludeNodeIPs invalid IP range: wrong number of IPs",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange1,
			expectErr:  true,
		},
		{
			name:       "ccmExcludeNodeIPs invalid IP range: wrong IP range",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange2,
			expectErr:  true,
		},
		{
			name:       "ccmExcludeNodeIPs invalid IP range: wrong IP types",
			dcConfFile: nutanixDatacenterConfigSpecWithCCMExcludeNodeIPsInvalidIPRange3,
			expectErr:  true,
		},
	}

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()
	mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string) (*v3.ClusterListIntentResponse, error) {
			return fakeClusterListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string, _ []*prismgoclient.AdditionalFilter) (*v3.SubnetListIntentResponse, error) {
			return fakeSubnetListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().GetSubnet(gomock.Any(), gomock.Eq("2d166190-7759-4dc6-b835-923262d6b497")).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().GetSubnet(gomock.Any(), gomock.Not("2d166190-7759-4dc6-b835-923262d6b497")).Return(nil, fmt.Errorf("")).AnyTimes()
	mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Eq("4d69ca7d-022f-49d1-a454-74535993bda4")).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Not("4d69ca7d-022f-49d1-a454-74535993bda4")).Return(nil, fmt.Errorf("")).AnyTimes()

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
			clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
			clusterSpec.NutanixDatacenter = dcConf

			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), clientCache.clients["test"], clusterSpec)
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
			clusterSpec := test.NewFullClusterSpec(t, "testdata/eksa-cluster.yaml")
			clusterSpec.NutanixDatacenter = dcConf

			err := yaml.Unmarshal([]byte(tc.dcConfFile), dcConf)
			require.NoError(t, err)

			err = validator.ValidateDatacenterConfig(context.Background(), clientCache.clients["test"], clusterSpec)
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
	clusterSpec.Cluster.Spec.ExternalEtcdConfiguration = &anywherev1.ExternalEtcdConfiguration{
		MachineGroupRef: &anywherev1.Ref{
			Name: "invalid-etcd-name",
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

	clusterSpec.NutanixMachineConfigs["eksa-unit-test-cp"].Spec.Image = anywherev1.NutanixResourceIdentifier{
		Name: utils.StringPtr("kubernetes_1_22"),
		Type: anywherev1.NutanixIdentifierName,
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

func TestValidateMachineConfigFailureDomainsWrongCount(t *testing.T) {
	ctx := context.Background()
	clusterConfigFile := "testdata/eksa-cluster-multi-worker-fds.yaml"
	clusterSpec := test.NewFullClusterSpec(t, clusterConfigFile)

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()
	mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string) (*v3.ClusterListIntentResponse, error) {
			return fakeClusterListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string) (*v3.SubnetListIntentResponse, error) {
			return fakeSubnetListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().GetSubnet(gomock.Any(), gomock.Eq("2d166190-7759-4dc6-b835-923262d6b497")).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Eq("4d69ca7d-022f-49d1-a454-74535993bda4")).Return(nil, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	for i := 0; i < len(clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations); i++ {
		clusterSpec.Cluster.Spec.WorkerNodeGroupConfigurations[i].Count = utils.IntPtr(20)
	}

	err := validator.validateFailureDomains(ctx, clientCache.clients["test"], clusterSpec)
	if err == nil {
		t.Fatalf("validation should not pass: %v", err)
	}
}

func TestValidateControlPlaneIP(t *testing.T) {
	tests := []struct {
		name          string
		clusterConfig string
		setIPfunc     func(*cluster.Spec) *cluster.Spec
		expectErr     bool
	}{
		{
			name:          "valid control plane IP",
			clusterConfig: "testdata/eksa-cluster.yaml",
			setIPfunc: func(clusterSpec *cluster.Spec) *cluster.Spec {
				clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "10.199.198.12"
				return clusterSpec
			},
			expectErr: false,
		},
		{
			name:          "invalid control plane IP",
			clusterConfig: "testdata/eksa-cluster.yaml",
			setIPfunc: func(clusterSpec *cluster.Spec) *cluster.Spec {
				clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host = "not-an-ip"
				return clusterSpec
			},
			expectErr: true,
		},
	}

	ctrl := gomock.NewController(t)
	mockClient := mocknutanix.NewMockClient(ctrl)
	mockClient.EXPECT().GetCurrentLoggedInUser(gomock.Any()).Return(&v3.UserIntentResponse{}, nil).AnyTimes()
	mockClient.EXPECT().ListAllCluster(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string) (*v3.ClusterListIntentResponse, error) {
			return fakeClusterListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().ListAllSubnet(gomock.Any(), gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ context.Context, filter string) (*v3.SubnetListIntentResponse, error) {
			return fakeSubnetListForDCTest(filter)
		},
	).AnyTimes()
	mockClient.EXPECT().GetSubnet(gomock.Any(), gomock.Eq("2d166190-7759-4dc6-b835-923262d6b497")).Return(nil, nil).AnyTimes()
	mockClient.EXPECT().GetCluster(gomock.Any(), gomock.Eq("4d69ca7d-022f-49d1-a454-74535993bda4")).Return(nil, nil).AnyTimes()

	mockTLSValidator := mockCrypto.NewMockTlsValidator(ctrl)
	mockTLSValidator.EXPECT().ValidateCert(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	mockTransport := mocknutanix.NewMockRoundTripper(ctrl)
	mockTransport.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{}, nil).AnyTimes()

	mockHTTPClient := &http.Client{Transport: mockTransport}
	clientCache := &ClientCache{clients: map[string]Client{"test": mockClient}}
	validator := NewValidator(clientCache, mockTLSValidator, mockHTTPClient)

	// Run tests
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			clusterSpec := test.NewFullClusterSpec(t, tc.clusterConfig)
			clusterSpec = tc.setIPfunc(clusterSpec)
			err := validator.validateControlPlaneIP(clusterSpec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host)
			if tc.expectErr {
				assert.Error(t, err, tc.name)
			} else {
				assert.NoError(t, err, tc.name)
			}
		})
	}
}
