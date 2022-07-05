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
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const (
	cmkConfigFileName  = "cmk_test_name.ini"
	cmkConfigFileName2 = "cmk_test_name_2.ini"
	accountName        = "account1"
	rootDomain         = "ROOT"
	rootDomainId       = "5300cdac-74d5-11ec-8696-c81f66d3e965"
	domain             = "foo/domain1"
	domainName         = "domain1"
	domainId           = "7700cdac-74d5-11ec-8696-c81f66d3e965"
	domain2            = "foo/bar/domain1"
	domain2Name        = "domain1"
	domain2Id          = "8800cdac-74d5-11ec-8696-c81f66d3e965"
	zoneId             = "4e3b338d-87a6-4189-b931-a1747edeea8f"
)

var execConfig = decoder.CloudStackExecConfig{
	Profiles: []decoder.CloudStackProfileConfig{
		{
			Name:          "test_name",
			ApiKey:        "test",
			SecretKey:     "test",
			ManagementUrl: "http://1.1.1.1:8080/client/api",
		},
	},
}

var execConfigWithMultipleProfiles = decoder.CloudStackExecConfig{
	Profiles: []decoder.CloudStackProfileConfig{
		execConfig.Profiles[0],
		{
			Name:          "test_name_2",
			ApiKey:        "test_2",
			SecretKey:     "test_2",
			ManagementUrl: "http://1.1.1.1:8080/client/api_2",
		},
	},
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

var diskOfferingCustomSizeInGB = v1alpha1.CloudStackResourceDiskOffering{
	CloudStackResourceIdentifier: v1alpha1.CloudStackResourceIdentifier{
		Id: "TEST_RESOURCE",
	},
	CustomSize: 1,
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
	c := executables.NewCmk(executable, writer, execConfig.Profiles)
	err := c.ValidateCloudStackConnection(ctx, execConfig.Profiles[0].Name)
	if err != nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() error = %v, want nil", err)
	}
}

func TestValidateMultipleCloudStackProfiles(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, nil)
	configFilePath2, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName2))
	expectedArgs2 := []string{"-c", configFilePath2, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs2).Return(bytes.Buffer{}, nil)

	c := executables.NewCmk(executable, writer, execConfigWithMultipleProfiles.Profiles)
	err := c.ValidateCloudStackConnection(ctx, execConfigWithMultipleProfiles.Profiles[0].Name)
	if err != nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() error = %v, want nil", err)
	}
	err = c.ValidateCloudStackConnection(ctx, execConfigWithMultipleProfiles.Profiles[1].Name)
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
	c := executables.NewCmk(executable, writer, execConfig.Profiles)
	err := c.ValidateCloudStackConnection(ctx, execConfig.Profiles[0].Name)
	if err == nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() didn't throw expected error")
	}
}

func TestCmkCleanupVms(t *testing.T) {
	_, writer := test.NewWriter(t)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	clusterName := "test"
	tests := []struct {
		testName           string
		argumentsExecCalls [][]string
		jsonResponseFile   string
		cmkFunc            func(cmk executables.Cmk, ctx context.Context) error
		cmkResponseError   error
		wantErr            bool
	}{
		{
			testName:         "listvirtualmachines json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCalls: [][]string{{
				"-c", configFilePath,
				"list", "virtualmachines", fmt.Sprintf("keyword=\"%s\"", clusterName), "listall=true",
			}},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.CleanupVms(ctx, execConfig.Profiles[0].Name, clusterName, false)
			},
			cmkResponseError: nil,
			wantErr:          true,
		},
		{
			testName:         "dry run succeeds",
			jsonResponseFile: "testdata/cmk_list_virtualmachine_singular.json",
			argumentsExecCalls: [][]string{
				{
					"-c", configFilePath,
					"list", "virtualmachines", fmt.Sprintf("keyword=\"%s\"", clusterName), "listall=true",
				},
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.CleanupVms(ctx, execConfig.Profiles[0].Name, clusterName, true)
			},
			cmkResponseError: nil,
			wantErr:          false,
		},
		{
			testName:         "listvirtualmachines no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCalls: [][]string{{
				"-c", configFilePath,
				"list", "virtualmachines", fmt.Sprintf("keyword=\"%s\"", clusterName), "listall=true",
			}},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.CleanupVms(ctx, execConfig.Profiles[0].Name, clusterName, false)
			},
			cmkResponseError: nil,
			wantErr:          true,
		},
		{
			testName:         "listaffinitygroups json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCalls: [][]string{{
				"-c", configFilePath,
				"list", "virtualmachines", fmt.Sprintf("keyword=\"%s\"", clusterName), "listall=true",
			}},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.CleanupVms(ctx, execConfig.Profiles[0].Name, clusterName, false)
			},
			cmkResponseError: nil,
			wantErr:          true,
		},
		{
			testName:         "full runthrough succeeds",
			jsonResponseFile: "testdata/cmk_list_virtualmachine_singular.json",
			argumentsExecCalls: [][]string{
				{
					"-c", configFilePath,
					"list", "virtualmachines", fmt.Sprintf("keyword=\"%s\"", clusterName), "listall=true",
				},
				{
					"-c", configFilePath, "stop", "virtualmachine", "id=\"30e8b0b1-f286-4372-9f1f-441e199a3f49\"",
					"forced=true",
				},
				{
					"-c", configFilePath, "destroy", "virtualmachine", "id=\"30e8b0b1-f286-4372-9f1f-441e199a3f49\"",
					"expunge=true",
				},
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.CleanupVms(ctx, execConfig.Profiles[0].Name, clusterName, false)
			},
			cmkResponseError: nil,
			wantErr:          false,
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
			for _, argsList := range tt.argumentsExecCalls {
				executable.EXPECT().Execute(ctx, argsList).
					Return(*bytes.NewBufferString(fileContent), tt.cmkResponseError)
			}
			cmk := executables.NewCmk(executable, writer, execConfig.Profiles)
			err := tt.cmkFunc(*cmk, ctx)
			if tt.wantErr && err != nil || !tt.wantErr && err == nil {
				return
			}
			t.Fatalf("Cmk error: %v", err)
		})
	}
}

func TestCmkListOperations(t *testing.T) {
	_, writer := test.NewWriter(t)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	tests := []struct {
		testName          string
		argumentsExecCall []string
		jsonResponseFile  string
		cmkFunc           func(cmk executables.Cmk, ctx context.Context) error
		cmkResponseError  error
		wantErr           bool
		wantResultCount   int
	}{
		{
			testName:         "listdomain success on name root",
			jsonResponseFile: "testdata/cmk_list_domain_root.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", rootDomain), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				domainId, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, rootDomain)
				if domainId != rootDomainId {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", rootDomainId, domainId)
				}
				return err
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  0,
		},
		{
			testName:         "listdomain success on name filter",
			jsonResponseFile: "testdata/cmk_list_domain_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				actualDomainId, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, domain)
				if actualDomainId != domainId {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", domainId, actualDomainId)
				}
				return err
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  0,
		},
		{
			testName:         "listdomain failure on multiple returns",
			jsonResponseFile: "testdata/cmk_list_domain_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, domainName)
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdomain success on multiple returns",
			jsonResponseFile: "testdata/cmk_list_domain_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domain2Name), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				domainId, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, domain2)
				if domainId != domain2Id {
					t.Fatalf("Expected domain id: %s, actual domain id: %s", domain2Id, domainId)
				}
				return err
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  0,
		},
		{
			testName:         "listdomains json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, domain)
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdomains no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "domains", fmt.Sprintf("name=\"%s\"", domainName), "listall=true",
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateDomainAndGetId(ctx, execConfig.Profiles[0].Name, domain)
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listaccounts success on name filter",
			jsonResponseFile: "testdata/cmk_list_account_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, execConfig.Profiles[0].Name, accountName, domainId)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  0,
		},
		{
			testName:         "listaccounts json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, execConfig.Profiles[0].Name, accountName, domainId)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listaccounts no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName), fmt.Sprintf("domainid=\"%s\"", domainId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, execConfig.Profiles[0].Name, accountName, domainId)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listzones success on name filter",
			jsonResponseFile: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[0])
				return err
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listzones success on id filter",
			jsonResponseFile: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceId.Id),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[2])
				return err
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listzones failure on multple results",
			jsonResponseFile: "testdata/cmk_list_zone_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceId.Id),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[2])
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listzones failure on none results",
			jsonResponseFile: "testdata/cmk_list_zone_none.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceId.Id),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[2])
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listzones failure on cmk failure",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[0])
				return err
			},
			cmkResponseError: errors.New("cmk calling return exception"),
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listzones no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[0])
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listzones json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName.Name),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				_, err := cmk.ValidateZoneAndGetId(ctx, execConfig.Profiles[0].Name, zones[0])
				return err
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listnetworks success on name filter",
			jsonResponseFile: "testdata/cmk_list_network_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listnetworks failure on multiple results",
			jsonResponseFile: "testdata/cmk_list_network_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listnetworks failure on none results",
			jsonResponseFile: "testdata/cmk_list_network_none.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listnetworks failure on cmk failure",
			jsonResponseFile: "testdata/cmk_list_network_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: errors.New("cmk calling return exception"),
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listnetworks no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listnetworks json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "networks", fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName), fmt.Sprintf("zoneid=\"%s\"", "TEST_RESOURCE"),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateNetworkPresent(ctx, execConfig.Profiles[0].Name, domainId, zones[2].Network, zones[2].Id, accountName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listserviceofferings success on name filter",
			jsonResponseFile: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listserviceofferings success on id filter",
			jsonResponseFile: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, resourceId)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listserviceofferings no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, resourceId)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listserviceofferings json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdiskofferings success on name filter",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceName)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listdiskofferings success on id filter",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listdiskofferings no results",
			jsonResponseFile: "testdata/cmk_list_diskoffering_empty.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdiskofferings no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdiskofferings multiple results",
			jsonResponseFile: "testdata/cmk_list_diskoffering_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  4,
		},
		{
			testName:         "listdiskofferings customized results with customSizeInGB > 0",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular_customized.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingCustomSizeInGB)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listdiskofferings non-customized results with customSizeInGB > 0",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingCustomSizeInGB)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listdiskofferings non-customized results with customSizeInGB > 0",
			jsonResponseFile: "testdata/cmk_list_diskoffering_singular_customized.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "listdiskofferings throw exception",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceID)
			},
			cmkResponseError: errors.New("cmk calling return exception"),
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listdiskofferings json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, execConfig.Profiles[0].Name, zoneId, diskOfferingResourceName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "validatetemplate success on name filter",
			jsonResponseFile: "testdata/cmk_list_template_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:          "validatetemplate failure when passing invalid profile",
			jsonResponseFile:  "testdata/cmk_list_template_singular.json",
			argumentsExecCall: nil,
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, "xxx", domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "validatetemplate success on id filter",
			jsonResponseFile: "testdata/cmk_list_template_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceId)
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "validatetemplate failure on multiple results",
			jsonResponseFile: "testdata/cmk_list_template_multiple.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceId)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "validatetemplate failure on none results",
			jsonResponseFile: "testdata/cmk_list_template_none.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceId)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "validatetemplate failure on cmk failure",
			jsonResponseFile: "testdata/cmk_list_template_none.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceId)
			},
			cmkResponseError: errors.New("cmk calling return exception"),
			wantErr:          true,
			wantResultCount:  1,
		},
		{
			testName:         "validatetemplate no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "validatetemplate json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName.Name), fmt.Sprintf("zoneid=\"%s\"", zoneId), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, execConfig.Profiles[0].Name, domainId, zoneId, accountName, resourceName)
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listaffinitygroups success on id filter",
			jsonResponseFile: "testdata/cmk_list_affinitygroup_singular.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, execConfig.Profiles[0].Name, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError: nil,
			wantErr:          false,
			wantResultCount:  1,
		},
		{
			testName:         "listaffinitygroups no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, execConfig.Profiles[0].Name, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
		},
		{
			testName:         "listaffinitygroups json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			argumentsExecCall: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceId.Id), fmt.Sprintf("domainid=\"%s\"", domainId), fmt.Sprintf("account=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, execConfig.Profiles[0].Name, domainId, accountName, []string{resourceId.Id})
			},
			cmkResponseError: nil,
			wantErr:          true,
			wantResultCount:  0,
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
			if tt.argumentsExecCall != nil {
				executable.EXPECT().Execute(ctx, tt.argumentsExecCall).
					Return(*bytes.NewBufferString(fileContent), tt.cmkResponseError)
			}
			cmk := executables.NewCmk(executable, writer, execConfig.Profiles)
			err := tt.cmkFunc(*cmk, ctx)
			if tt.wantErr && err != nil || !tt.wantErr && err == nil {
				return
			}
			t.Fatalf("Cmk error: %v", err)
		})
	}
}

func TestCmkGetManagementApiEndpoint(t *testing.T) {
	_, writer := test.NewWriter(t)
	mockCtrl := gomock.NewController(t)
	tt := NewWithT(t)

	var tctx testContext
	tctx.SaveContext()
	defer tctx.RestoreContext()

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	cmk := executables.NewCmk(executable, writer, execConfigWithMultipleProfiles.Profiles)

	endpoint, err := cmk.GetManagementApiEndpoint("test_name")
	tt.Expect(err).To(BeNil())
	tt.Expect(endpoint).To(Equal("http://1.1.1.1:8080/client/api"))

	endpoint, err = cmk.GetManagementApiEndpoint("test_name_2")
	tt.Expect(err).To(BeNil())
	tt.Expect(endpoint).To(Equal("http://1.1.1.1:8080/client/api_2"))

	_, err = cmk.GetManagementApiEndpoint("xxx")
	tt.Expect(err).NotTo(BeNil())
}
