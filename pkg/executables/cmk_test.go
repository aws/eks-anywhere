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
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

const (
	cmkConfigFileName = "cmk_tmp.ini"
	resourceName      = "TEST_RESOURCE"
	zoneName          = "zone1"
	accountName       = "account1"
	domainName        = "domain1"
)

var execConfig = executables.CmkExecConfig{
	CloudStackApiKey:        "test",
	CloudStackSecretKey:     "test",
	CloudStackManagementUrl: "http://1.1.1.1:8080/client/api",
	CloudMonkeyVerifyCert:   false,
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
		argumentsExecCall1    []string
		argumentsExecCall2    []string
		jsonResponseFile1     string
		jsonResponseFile2     string
		cmkFunc               func(cmk executables.Cmk, ctx context.Context) error
		cmkResponseError      error
		wantErr               bool
		shouldSecondCallOccur bool
		wantResultCount       int
	}{
		{
			testName:          "listzones success",
			jsonResponseFile1: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateZonePresent(ctx, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listzones success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateZonePresent(ctx, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "listzones no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateZonePresent(ctx, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listzones json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateZonePresent(ctx, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "validate disk offerings success",
			jsonResponseFile1: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, domainName, zoneName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listdiskofferings success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, domainName, zoneName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "listdiskofferings no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, domainName, zoneName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listdiskofferings json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateDiskOfferingPresent(ctx, domainName, zoneName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaccounts success",
			jsonResponseFile1: "testdata/cmk_list_account_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listaccounts success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_account_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("id=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "listaccounts no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("id=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listaccounts json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", accountName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAccountPresent(ctx, accountName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listserviceofferings success",
			jsonResponseFile1: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listserviceofferings success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "listserviceofferings no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listserviceofferings json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateServiceOfferingPresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "validatetemplate success",
			jsonResponseFile1: "testdata/cmk_list_template_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "validatetemplate success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_template_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "validatetemplate no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			argumentsExecCall2: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "validatetemplate json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateTemplatePresent(ctx, zoneName, domainName, accountName, resourceName)
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaffinitygroups success on id filter",
			jsonResponseFile1: "testdata/cmk_list_affinitygroup_singular.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, zoneName, domainName, accountName, []string{resourceName})
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listaffinitygroups no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, zoneName, domainName, accountName, []string{resourceName})
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaffinitygroups json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{
				"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName),
			},
			cmkFunc: func(cmk executables.Cmk, ctx context.Context) error {
				return cmk.ValidateAffinityGroupsPresent(ctx, zoneName, domainName, accountName, []string{resourceName})
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			fileContent1 := test.ReadFile(t, tt.jsonResponseFile1)

			ctx := context.Background()
			mockCtrl := gomock.NewController(t)

			var tctx testContext
			tctx.SaveContext()
			defer tctx.RestoreContext()

			executable := mockexecutables.NewMockExecutable(mockCtrl)
			executable.EXPECT().Execute(ctx, tt.argumentsExecCall1).
				Return(*bytes.NewBufferString(fileContent1), tt.cmkResponseError)
			if tt.shouldSecondCallOccur {
				fileContent2 := test.ReadFile(t, tt.jsonResponseFile2)
				executable.EXPECT().Execute(ctx, tt.argumentsExecCall2).
					Return(*bytes.NewBufferString(fileContent2), tt.cmkResponseError)
			}
			cmk := executables.NewCmk(executable, writer, execConfig)
			err := tt.cmkFunc(*cmk, ctx)
			if tt.wantErr && err != nil {
				return
			}
			if err != nil {
				t.Fatalf("Cmk error: %v", err)
			}
		})
	}
}
