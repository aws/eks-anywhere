package filewriter_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

func TestTmpWriterWriteValid(t *testing.T) {
	folder := "tmp_folder"
	folder2 := "tmp_folder_2"
	err := os.MkdirAll(folder2, os.ModePerm)
	if err != nil {
		t.Fatalf("error setting up test: %v", err)
	}
	defer os.RemoveAll(folder)
	defer os.RemoveAll(folder2)

	tests := []struct {
		testName string
		dir      string
		fileName string
		content  []byte
	}{
		{
			testName: "dir doesn't exist",
			dir:      folder,
			fileName: "TestTmpWriterWriteValid-success.yaml",
			content: []byte(`
			fake content
			blablab
			`),
		},
		{
			testName: "dir exists",
			dir:      folder2,
			fileName: "test",
			content: []byte(`
			fake content
			blablab
			`),
		},
		{
			testName: "empty file name",
			dir:      folder,
			fileName: "test",
			content: []byte(`
			fake content
			blablab
			`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tr, err := filewriter.NewWriter(tt.dir)
			if err != nil {
				t.Fatalf("failed creating tmpWriter error = %v", err)
			}

			gotPath, err := tr.Write(tt.fileName, tt.content)
			if err != nil {
				t.Fatalf("tmpWriter.Write() error = %v", err)
			}

			if !strings.HasPrefix(gotPath, tt.dir) {
				t.Errorf("tmpWriter.Write() = %v, want to start with %v", gotPath, tt.dir)
			}

			if !strings.HasSuffix(gotPath, tt.fileName) {
				t.Errorf("tmpWriter.Write() = %v, want to end with %v", gotPath, tt.fileName)
			}

			content, err := ioutil.ReadFile(gotPath)
			if err != nil {
				t.Fatalf("error reading written file: %v", err)
			}

			if string(content) != string(tt.content) {
				t.Errorf("Write file content = %v, want %v", content, tt.content)
			}
		})
	}
}

func TestTmpWriterWithDir(t *testing.T) {
	rootFolder := "folder_root"
	subFolder := "subFolder"
	defer os.RemoveAll(rootFolder)

	tr, err := filewriter.NewWriter(rootFolder)
	if err != nil {
		t.Fatalf("failed creating tmpWriter error = %v", err)
	}

	fw, err := tr.WithDir(subFolder)
	if err != nil {
		t.Fatalf("failed creating tmpWriter with subdir error = %v", err)
	}

	gotPath, err := fw.Write("file.txt", []byte("file content"))
	if err != nil {
		t.Fatalf("tmpWriter.Write() error = %v", err)
	}

	wantPathPrefix := filepath.Join(rootFolder, subFolder)
	if !strings.HasPrefix(gotPath, wantPathPrefix) {
		t.Errorf("tmpWriter.Write() = %v, want to start with %v", gotPath, wantPathPrefix)
	}
}
