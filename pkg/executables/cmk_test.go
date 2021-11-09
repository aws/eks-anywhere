package executables_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/golang/mock/gomock"
)

const (
	cmkConfigFileName  	= "cmk_tmp.ini"
	templateName		= "TEST_TEMPLATE"
	offeringName 		= "TEST_OFFERING"
)

func TestValidateCloudStackConnectionSuccess(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, nil)
	c := executables.NewCmk(executable, writer)
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
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	expectedArgs := []string{"-c", configFilePath, "sync"}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, errors.New("cmk test error"))
	c := executables.NewCmk(executable, writer)
	err := c.ValidateCloudStackConnection(ctx)
	if err == nil {
		t.Fatalf("Cmk.ValidateCloudStackConnection() didn't throw expected error")
	}
}

func TestRegisterSshKeyPairSuccess(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	keyName := "testKeyname"
	keyValue := "ssh-rsa key-value"

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	expectedArgs := []string{"-c", configFilePath, "register", "sshkeypair", fmt.Sprintf("name=\"%s\"", keyName),
		fmt.Sprintf("publickey=\"%s\"", keyValue)}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, nil)
	c := executables.NewCmk(executable, writer)
	err := c.RegisterSSHKeyPair(ctx, keyName, keyValue)
	if err != nil {
		t.Fatalf("Awscli.CreateAccessKey() error = %v, want nil", err)
	}
}

func TestRegisterSshKeyPairError(t *testing.T) {
	_, writer := test.NewWriter(t)
	ctx := context.Background()
	mockCtrl := gomock.NewController(t)
	keyName := "testKeyname"
	keyValue := "ssh-rsa key-value"

	executable := mockexecutables.NewMockExecutable(mockCtrl)
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	expectedArgs := []string{"-c", configFilePath, "register", "sshkeypair", fmt.Sprintf("name=\"%s\"", keyName),
		fmt.Sprintf("publickey=\"%s\"", keyValue)}
	executable.EXPECT().Execute(ctx, expectedArgs).Return(bytes.Buffer{}, errors.New("cmk test error"))
	c := executables.NewCmk(executable, writer)
	err := c.RegisterSSHKeyPair(ctx, keyName, keyValue)
	if err == nil {
		t.Fatalf("Cmk.RegisterSshKeyPair() didn't throw expected error")
	}
}

func TestListTemplates(t *testing.T) {
	_, writer := test.NewWriter(t)
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	tests := []struct {
		testName         	string
		jsonResponseFile 	string
		wantErr				bool
		wantResultCount		int
	}{
		{
			testName:         "success",
			jsonResponseFile: "testdata/cmk_list_template_singular.json",
			wantErr: false,
			wantResultCount: 1,
		},
		{
			testName:         "no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			wantErr: false,
			wantResultCount: 0,
		},
		{
			testName:         "json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			wantErr: true,
			wantResultCount: 0,
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
			executable.EXPECT().Execute(ctx, []string{"-c", configFilePath,
				"list", "templates", "templatefilter=all", "listall=true", fmt.Sprintf("name=\"%s\"", templateName)}).
				Return(*bytes.NewBufferString(fileContent), nil)
			cmk := executables.NewCmk(executable, writer)
			templates, err := cmk.ListTemplates(ctx, templateName)
			if tt.wantErr {
				return
			}
			if err != nil {
				t.Fatalf("Cmk.ListTemplates() error: %v", err)
			}

			if len(templates) != tt.wantResultCount {
				t.Fatalf("Cmk.ListTemplates returned = %d results, want %d", len(templates), tt.wantResultCount)
			}
		})
	}
}

func TestListServiceOfferings(t *testing.T) {
	_, writer := test.NewWriter(t)
	configFilePath := filepath.Join(writer.Dir(), "generated", cmkConfigFileName)
	tests := []struct {
		testName         	string
		jsonResponseFile 	string
		wantErr				bool
		wantResultCount		int
	}{
		{
			testName:         "success",
			jsonResponseFile: "testdata/cmk_list_serviceoffering_singular.json",
			wantErr: false,
			wantResultCount: 1,
		},
		{
			testName:         "no results",
			jsonResponseFile: "testdata/cmk_list_empty_response.json",
			wantErr: false,
			wantResultCount: 0,
		},
		{
			testName:         "json parse exception",
			jsonResponseFile: "testdata/cmk_non_json_response.txt",
			wantErr: true,
			wantResultCount: 0,
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
			executable.EXPECT().Execute(ctx, []string{"-c", configFilePath,
				"list", "serviceofferings", fmt.Sprintf("name=\"%s\"", offeringName)}).
				Return(*bytes.NewBufferString(fileContent), nil)
			cmk := executables.NewCmk(executable, writer)
			templates, err := cmk.ListServiceOfferings(ctx, offeringName)
			if tt.wantErr {
				return
			}
			if err != nil {
				t.Fatalf("Cmk.ListServiceOfferings() error: %v", err)
			}

			if len(templates) != tt.wantResultCount {
				t.Fatalf("Cmk.ListServiceOfferings returned = %d results, want %d", len(templates), tt.wantResultCount)
			}
		})
	}
}
