package filewriter_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

func TestWriterWriteValid(t *testing.T) {
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
			testName: "test 1",
			dir:      folder,
			fileName: "TestWriterWriteValid-success.yaml",
			content: []byte(`
			fake content
			blablab
			`),
		},
		{
			testName: "test 2",
			dir:      folder2,
			fileName: "TestWriterWriteValid-success.yaml",
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
				t.Fatalf("failed creating writer error = %v", err)
			}

			gotPath, err := tr.Write(tt.fileName, tt.content)
			if err != nil {
				t.Fatalf("writer.Write() error = %v", err)
			}

			wantPath := filepath.Join(tt.dir, filewriter.DefaultTmpFolder, tt.fileName)
			if strings.Compare(gotPath, wantPath) != 0 {
				t.Errorf("writer.Write() = %v, want %v", gotPath, wantPath)
			}

			test.AssertFilesEquals(t, gotPath, wantPath)
		})
	}
}

func TestEmptyFileName(t *testing.T) {
	folder := "tmp_folder"
	defer os.RemoveAll(folder)
	tr, err := filewriter.NewWriter(folder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}
	_, err = tr.Write("", []byte("content"))
	if err == nil {
		t.Fatalf("writer.Write() error is nil")
	}
}

func TestWriterWithDir(t *testing.T) {
	rootFolder := "folder_root"
	subFolder := "subFolder"
	defer os.RemoveAll(rootFolder)

	tr, err := filewriter.NewWriter(rootFolder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	tr, err = tr.WithDir(subFolder)
	if err != nil {
		t.Fatalf("failed creating writer with subdir error = %v", err)
	}

	gotPath, err := tr.Write("file.txt", []byte("file content"))
	if err != nil {
		t.Fatalf("writer.Write() error = %v", err)
	}

	wantPathPrefix := filepath.Join(rootFolder, subFolder)
	if !strings.HasPrefix(gotPath, wantPathPrefix) {
		t.Errorf("writer.Write() = %v, want to start with %v", gotPath, wantPathPrefix)
	}
}

func TestWriterWritePersistent(t *testing.T) {
	folder := "tmp_folder_opt"
	folder2 := "tmp_folder_2_opt"
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
		options  []filewriter.FileOptionsFunc
	}{
		{
			testName: "Write persistent file",
			dir:      folder,
			fileName: "TestWriterWriteValid-success.yaml",
			content: []byte(`
			fake content
			blablab
			`),
			options: []filewriter.FileOptionsFunc{filewriter.PersistentFile},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			tr, err := filewriter.NewWriter(tt.dir)
			if err != nil {
				t.Fatalf("failed creating writer error = %v", err)
			}

			gotPath, err := tr.Write(tt.fileName, tt.content, tt.options...)
			if err != nil {
				t.Fatalf("writer.Write() error = %v", err)
			}

			wantPath := filepath.Join(tt.dir, tt.fileName)
			if strings.Compare(gotPath, wantPath) != 0 {
				t.Errorf("writer.Write() = %v, want %v", gotPath, wantPath)
			}

			test.AssertFilesEquals(t, gotPath, wantPath)
		})
	}
}

func TestWriterDir(t *testing.T) {
	rootFolder := "folder_root"
	defer os.RemoveAll(rootFolder)

	tr, err := filewriter.NewWriter(rootFolder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	if strings.Compare(tr.Dir(), rootFolder) != 0 {
		t.Errorf("writer.Dir() = %v, want %v", tr.Dir(), rootFolder)
	}
}

func TestWriterTempDir(t *testing.T) {
	rootFolder := "folder_root"
	tempFolder := fmt.Sprintf("%s/generated", rootFolder)
	defer os.RemoveAll(rootFolder)

	tr, err := filewriter.NewWriter(rootFolder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	if strings.Compare(tr.TempDir(), tempFolder) != 0 {
		t.Errorf("writer.TempDir() = %v, want %v", tr.TempDir(), tempFolder)
	}
}

func TestWriterCleanUpTempDir(t *testing.T) {
	rootFolder := "folder_root"
	defer os.RemoveAll(rootFolder)

	tr, err := filewriter.NewWriter(rootFolder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	tr.CleanUpTemp()

	if _, err := os.Stat(tr.TempDir()); err == nil {
		t.Errorf("writer.CleanUp(), want err, got nil")
	}
}
