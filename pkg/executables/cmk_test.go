package executables_test

import (
	"bytes"
	"context"
	_ "embed"
	b64 "encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

const (
	cmkConfigFileName             = "cmk_tmp.ini"
	resourceName                  = "TEST_RESOURCE"
	zoneName					  = "zone1"
	accountName					  = "account1"
	domainName					  = "domain1"
	cloudStackb64EncodedSecretKey = "CLOUDSTACK_B64ENCODED_SECRET"
	cloudmonkeyInsecureKey        = "CLOUDMONKEY_INSECURE"
)

//go:embed testdata/cloudstack_secret_file.ini
var cloudstackSecretFile []byte

var (
	cloudStackb64EncodedSecretPreviousValue string
	cloudmonkeyInsecureKeyPreviousValue     string
)

func saveAndSetEnv() {
	cloudStackb64EncodedSecretPreviousValue = os.Getenv(cloudStackb64EncodedSecretKey)
	os.Setenv(cloudStackb64EncodedSecretKey, b64.StdEncoding.EncodeToString(cloudstackSecretFile))
	os.Setenv(cloudmonkeyInsecureKey, "false")
}

func restoreEnv() {
	os.Setenv(cloudStackb64EncodedSecretKey, cloudStackb64EncodedSecretPreviousValue)
	os.Setenv(cloudmonkeyInsecureKey, cloudmonkeyInsecureKeyPreviousValue)
}

func TestValidateCloudStackConnectionSuccess(t *testing.T) {
	saveAndSetEnv()
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, nil)
	c := executables.NewCmk(executable, writer)
	err := c.ValidateCloudStackConnection(ctx)
	if err != nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() error = %v, want nil", err)
	}
	restoreEnv()
}

func TestValidateCloudStackConnectionError(t *testing.T) {
	saveAndSetEnv()
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, errors.New("cmk test error"))
	c := executables.NewCmk(executable, writer)
	err := c.ValidateCloudStackConnection(ctx)
	if err == nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() didn't throw expected error")
	}
	restoreEnv()
}

func TestCmkListOperations(t *testing.T) {
	saveAndSetEnv()
	_, writer := test.NewWriter(t)
	configFilePath, _ := filepath.Abs(filepath.Join(writer.Dir(), "generated", cmkConfigFileName))
	tests := []struct {
		testName              string
		argumentsExecCall1    []string
		argumentsExecCall2    []string
		jsonResponseFile1     string
		jsonResponseFile2     string
		cmkFunc               func(cmk TestCmkClient, ctx context.Context) error
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) error {
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
			cmk := executables.NewCmk(executable, writer)
			err := tt.cmkFunc(cmk, ctx)
			if tt.wantErr && err != nil {
				return
			}
			if err != nil {
				t.Fatalf("Cmk.ListZones() error: %v", err)
			}
		})
	}
	restoreEnv()
}

type TestCmkClient interface {
	ValidateCloudStackConnection(ctx context.Context) error
	ValidateTemplatePresent(ctx context.Context, domain, zone, account, template string) error
	ValidateServiceOfferingPresent(ctx context.Context, domain, zone, account, serviceOffering string) error
	ValidateDiskOfferingPresent(ctx context.Context, domain, zone, account, diskOffering string) error
	ValidateZonePresent(ctx context.Context, zone string) error
	ValidateAccountPresent(ctx context.Context, account string) error
	ValidateAffinityGroupsPresent(ctx context.Context, domain, zone, account string, affinityGroupIds []string) error
}
