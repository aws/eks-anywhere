package executables_test

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const (
	cmkConfigFileName = "cmk_tmp.ini"
	accountName       = "account1"
	rootDomain        = "ROOT"
	rootDomainId      = "5300cdac-74d5-11ec-8696-c81f66d3e965"
	domain            = "foo/domain1"
	domainName        = "domain1"
	domainId          = "7700cdac-74d5-11ec-8696-c81f66d3e965"
	domain2           = "foo/bar/domain1"
	domain2Name       = "domain1"
	domain2Id         = "8800cdac-74d5-11ec-8696-c81f66d3e965"
	zoneId            = "4e3b338d-87a6-4189-b931-a1747edeea8f"
)

var execConfig = decoder.CloudStackExecConfig{
	ApiKey:        "test",
	SecretKey:     "test",
	ManagementUrl: "http://1.1.1.1:8080/client/api",
}

var zones = []v1alpha1.CloudStackZone{
	{Name: "TEST_RESOURCE", Network: v1alpha1.CloudStackResourceIdentifier{Name: "TEST_RESOURCE"}},
	{Name: "TEST_RESOURCE", Network: v1alpha1.CloudStackResourceIdentifier{Id: "TEST_RESOURCE"}},
	{Id: "TEST_RESOURCE", Network: v1alpha1.CloudStackResourceIdentifier{Name: "TEST_RESOURCE"}},
	{Id: "TEST_RESOURCE", Network: v1alpha1.CloudStackResourceIdentifier{Id: "TEST_RESOURCE"}},
}

var resourceName = v1alpha1.CloudStackResourceIdentifier{
	Name: "TEST_RESOURCE",
}

var resourceId = v1alpha1.CloudStackResourceIdentifier{
	Id: "TEST_RESOURCE",
}

var diskOfferingResourceName = v1alpha1.CloudStackResourceDiskOffering{
	CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
		Name: "TEST_RESOURCE",
	},
	MountPath:  "/TEST_RESOURCE",
	Device:     "/dev/vdb",
	Filesystem: "ext4",
	Label:      "data_disk",
}

var diskOfferingResourceID = v1alpha1.CloudStackResourceDiskOffering{
	CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
		Id: "TEST_RESOURCE",
	},
	MountPath:  "/TEST_RESOURCE",
	Device:     "/dev/vdb",
	Filesystem: "ext4",
	Label:      "data_disk",
}

func TestValidateCloudStackConnectionSuccess(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, nil)
	c := executables.NewCmk(executable, writer, execConfig)
	err := c.ValidateCloudStackConnection(ctx)
	if err != nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() error = %v, want nil", err)
	}
}

func TestValidateCloudStackConnectionError(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, errors.New("cmk test error"))
	c := executables.NewCmk(executable, writer, execConfig)
	err := c.ValidateCloudStackConnection(ctx)
	if err == nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() didn't throw expected error")
	}
}

func TestCmkListOperations(t *testing.T) {
	_, writer := test.NewWriter(t)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	tests := []struct {
		testName              string
		argumentsExecCall     []string
		jsonResponseFile      string
		cmkFunc               func(cmk executables.Cmk, ctx context.Context) error
		cmkResponseError      error
		wantErr               bool
		shouldSecondCallOccur bool
		wantResultCount       int
	}{
		{
			testName:         "listdomain success on name root",
			jsonResponseFile: "testdata/cmk_list_domain_root.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", rootDomain), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				domain, err := cmk.ValidateDomainPresent(ctx, rootDomain)
				if domain.Id != rootDomainId {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", rootDomainId, domain.Id)
				}
				return err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdomain success on name filter",
			jsonResponseFile: "testdata/cmk_list_domain_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				domain, err := cmk.ValidateDomainPresent(ctx, domain)
				if domain.Id != domainId {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", domainId, domain.Id)
				}
				return err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdomain failure on multiple returns",
			jsonResponseFile: "testdata/cmk_list_domain_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainPresent(ctx, domainName)
				return err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdomain success on multiple returns",
			jsonResponseFile: "testdata/cmk_list_domain_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domain2Name), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				domain, err := cmk.ValidateDomainPresent(ctx, domain2)
				if domain.Id != domain2Id {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", domain2Id, domain.Id)
				}
				return err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdomains json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainPresent(ctx, domain)
				return err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listdomains no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainPresent(ctx, domain)
				return err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listaccounts success on name filter",
			jsonResponseFile: "testdata/cmk_list_account_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName, domainId)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listaccounts json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName, domainId)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listaccounts no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName, domainId)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listzones success on name filter",
			jsonResponseFile: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZonesPresent(ctx, []v1alpha1.CloudStackZone{zones[0]})
				return err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "listzones success on id filter",
			jsonResponseFile: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceId.Id),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZonesPresent(ctx, []v1alpha1.CloudStackZone{zones[2]})
				return err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:         "listzones no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZonesPresent(ctx, zones)
				return err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listzones json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZonesPresent(ctx, zones)
				return err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listnetworks success on name filter",
			jsonResponseFile: "testdata/cmk_list_network_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, domainId, zones[2], []v1alpha1.CloudStackResourceIdentifier{}, accountName, false)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "listnetworks success on id filter",
			jsonResponseFile: "testdata/cmk_list_network_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, domainId, zones[3], []v1alpha1.CloudStackResourceIdentifier{}, accountName, false)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:         "listnetworks no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, domainId, zones[2], []v1alpha1.CloudStackResourceIdentifier{}, accountName, false)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listnetworks json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, domainId, zones[2], []v1alpha1.CloudStackResourceIdentifier{}, accountName, false)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listserviceofferings success on name filter",
			jsonResponseFile: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneId, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "listserviceofferings success on id filter",
			jsonResponseFile: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneId, resourceId)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:         "listserviceofferings no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneId, resourceId)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listserviceofferings json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneId, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listdiskofferings success on name filter",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "listdiskofferings success on id filter",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceID)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:         "listdiskofferings no results",
			jsonResponseFile: "testdata/cmk_list_diskoffering_empty.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceID)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdiskofferings multiple results",
			jsonResponseFile: "testdata/cmk_list_diskoffering_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceID)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       4,
		},
		{
			testName:         "listdiskofferings throw exception",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceID)
			},
			cmkResponseError:      errors.New("cmk calling return exception"),
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "listdiskofferings json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, zoneId, diskOfferingResourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "validatetemplate success on name filter",
			jsonResponseFile: "testdata/cmk_list_template_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "validatetemplate success on id filter",
			jsonResponseFile: "testdata/cmk_list_template_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, domainId, zoneId, accountName, resourceId)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:         "validatetemplate no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:         "validatetemplate json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listaffinitygroups success on id filter",
			jsonResponseFile: "testdata/cmk_list_affinitygroup_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:         "listaffinitygroups no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:         "listaffinitygroups json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent := test.ReadFile(t, tt.jsonResponseFile)

			ctx := context.Background()
			mockCtrl := gomock.NewController(t)

			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()

			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(ctx, tt.argumentsExecCall).
				Return(*bytes.NewBufferString(fileContent), tt.cmkResponseError)
			cmk := executables.NewCmk(executable, writer, execConfig)
			err := tt.cmkFunc(*cmk, ctx)
			if tt.wantErr && err != nil || !tt.wantErr && err == nil {
				return
			}
			t.Fatalf("Cmk error: %v", err)
		})
	}
}
