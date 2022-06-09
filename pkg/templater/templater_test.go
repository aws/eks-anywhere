package templater_test

import (
	"os"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/templater"
)

func TestTemplaterWriteToFileSuccess(t *testing.T) {
	type dataStruct struct {
		Key1, Key2, Key3, KeyAndValue3 string
		Conditional                    bool
	}

	tests := []struct {
		testName     string
		templateFile string
		data         dataStruct
		fileName     string
		wantFilePath string
		wantErr      bool
	}{
		{
			testName:     "with conditional true",
			templateFile: "testdata/test1_template.yaml",
			data: dataStruct{
				Key1:        "value_1",
				Key2:        "value_2",
				Key3:        "value_3",
				Conditional: true,
			},
			fileName:     "file_tmp.yaml",
			wantFilePath: "testdata/test1_conditional_true_want.yaml",
			wantErr:      false,
		},
		{
			testName:     "with conditional false",
			templateFile: "testdata/test1_template.yaml",
			data: dataStruct{
				Key1:        "value_1",
				Key2:        "value_2",
				Key3:        "value_3",
				Conditional: false,
			},
			fileName:     "file_tmp.yaml",
			wantFilePath: "testdata/test1_conditional_false_want.yaml",
			wantErr:      false,
		},
		{
			testName:     "with indent",
			templateFile: "testdata/test_indent_template.yaml",
			data: dataStruct{
				Key1:         "value_1",
				Key2:         "value_2",
				KeyAndValue3: "key3: value_3",
				Conditional:  true,
			},
			fileName:     "file_tmp.yaml",
			wantFilePath: "testdata/test_indent_want.yaml",
			wantErr:      false,
		},
		{
			testName:     "with marshal",
			templateFile: "testdata/test_marshal_template.yaml",
			data: dataStruct{
				Key1:         "value_1",
				Key2:         "value_2",
				KeyAndValue3: "key3: value_3",
				Conditional:  true,
			},
			fileName:     "file_tmp.yaml",
			wantFilePath: "testdata/test_marshal_want.yaml",
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			_, writer := test.NewWriter(t)
			tr := templater.New(writer)
			templateContent := test.ReadFile(t, tt.templateFile)
			gotFilePath, err := tr.WriteToFile(templateContent, tt.data, tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Templater.WriteToFile() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !strings.HasSuffix(gotFilePath, tt.fileName) {
				t.Errorf("Templater.WriteToFile()  = %v, want to end with %v", gotFilePath, tt.fileName)
			}

			test.AssertFilesEquals(t, gotFilePath, tt.wantFilePath)
		})
	}
}

func TestTemplaterWriteToFileError(t *testing.T) {
	folder := "tmp_folder"
	defer os.RemoveAll(folder)

	writer, err := filewriter.NewWriter(folder)
	if err != nil {
		t.Fatalf("failed creating writer error = #{err}")
	}

	type dataStruct struct {
		Key1, Key2, Key3 string
		Conditional      bool
	}

	tests := []struct {
		testName     string
		templateFile string
		data         dataStruct
		fileName     string
	}{
		{
			testName:     "invalid template",
			templateFile: "testdata/invalid_template.yaml",
			data: dataStruct{
				Key1:        "value_1",
				Key2:        "value_2",
				Key3:        "value_3",
				Conditional: true,
			},
			fileName: "file_tmp.yaml",
		},
		{
			testName:     "data doesn't exist",
			templateFile: "testdata/key4_template.yaml",
			data: dataStruct{
				Key1:        "value_1",
				Key2:        "value_2",
				Key3:        "value_3",
				Conditional: false,
			},
			fileName: "file_tmp.yaml",
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tr := templater.New(writer)
			templateContent := test.ReadFile(t, tt.templateFile)
			gotFilePath, err := tr.WriteToFile(templateContent, tt.data, tt.fileName)
			if err == nil {
				t.Errorf("Templater.WriteToFile() error = nil")
			}

			if gotFilePath != "" {
				t.Errorf("Templater.WriteToFile() = %v, want nil", gotFilePath)
			}
		})
	}
}
