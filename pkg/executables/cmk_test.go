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

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
	"github.com/golang/mock/gomock"
)

const (
	cmkConfigFileName             = "cmk_tmp.ini"
	resourceName                  = "TEST_RESOURCE"
	cloudStackb64EncodedSecretKey = "CLOUDSTACK_B64ENCODED_SECRET"
	cloudmonkeyInsecureKey        = "CLOUDMONKEY_INSECURE"
)

//go:embed testdata/cloudstack_secret_file.ini
var cloudstackSecretFile []byte
var cloudStackb64EncodedSecretPreviousValue string
var cloudmonkeyInsecureKeyPreviousValue string

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
		cmkFunc               func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error)
		cmkResponseError      error
		wantErr               bool
		shouldSecondCallOccur bool
		wantResultCount       int
	}{
		{
			testName:          "listzones success",
			jsonResponseFile1: "testdata/cmk_list_zone_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListZones(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListZones(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListZones(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listzones json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "zones", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListZones(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listdiskofferings success",
			jsonResponseFile1: "testdata/cmk_list_diskoffering_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListDiskOfferings(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListDiskOfferings(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListDiskOfferings(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listdiskofferings json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "diskofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListDiskOfferings(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaccounts success",
			jsonResponseFile1: "testdata/cmk_list_account_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAccounts(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAccounts(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAccounts(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listaccounts json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "accounts", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAccounts(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listserviceofferings success",
			jsonResponseFile1: "testdata/cmk_list_serviceoffering_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListServiceOfferings(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListServiceOfferings(ctx, resourceName)
				return len(response), err
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
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListServiceOfferings(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listserviceofferings json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListServiceOfferings(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listtemplates success",
			jsonResponseFile1: "testdata/cmk_list_template_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListTemplates(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listtemplates success on id filter",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_template_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListTemplates(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       1,
		},
		{
			testName:          "listtemplates no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			jsonResponseFile2: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName)},
			argumentsExecCall2: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListTemplates(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: true,
			wantResultCount:       0,
		},
		{
			testName:          "listtemplates json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListTemplates(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               true,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaffinitygroups success on id filter",
			jsonResponseFile1: "testdata/cmk_list_affinitygroup_singular.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAffinityGroupsById(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       1,
		},
		{
			testName:          "listaffinitygroups no results",
			jsonResponseFile1: "testdata/cmk_list_empty_response.json",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAffinityGroupsById(ctx, resourceName)
				return len(response), err
			},
			cmkResponseError:      nil,
			wantErr:               false,
			shouldSecondCallOccur: false,
			wantResultCount:       0,
		},
		{
			testName:          "listaffinitygroups json parse exception",
			jsonResponseFile1: "testdata/cmk_non_json_response.txt",
			argumentsExecCall1: []string{"-c", configFilePath,
				"list", "affinitygroups", fmt.Sprintf("id=\"%s\"", resourceName)},
			cmkFunc: func(cmk TestCmkClient, ctx context.Context) (responseLength int, err error) {
				response, err := cmk.ListAffinityGroupsById(ctx, resourceName)
				return len(response), err
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
			resultCount, err := tt.cmkFunc(cmk, ctx)
			if tt.wantErr && err != nil {
				return
			}
			if err != nil {
				t.Fatalf("Cmk.ListZones() error: %v", err)
			}

			if resultCount != tt.wantResultCount {
				t.Fatalf("Cmk call returned = %d results, want %d", resultCount, tt.wantResultCount)
			}
		})
	}
	restoreEnv()
}

type TestCmkClient interface {
	ValidateCloudStackConnection(ctx context.Context) error
	ListTemplates(ctx context.Context, template string) ([]types.CmkTemplate, error)
	ListServiceOfferings(ctx context.Context, offering string) ([]types.CmkServiceOffering, error)
	ListDiskOfferings(ctx context.Context, offering string) ([]types.CmkDiskOffering, error)
	ListZones(ctx context.Context, offering string) ([]types.CmkZone, error)
	ListAccounts(ctx context.Context, accountName string) ([]types.CmkAccount, error)
	ListAffinityGroupsById(ctx context.Context, affinityGroupId string) ([]types.CmkAffinityGroup, error)
}
