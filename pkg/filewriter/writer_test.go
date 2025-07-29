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

func TestWriterDeleteSuccess(t *testing.T) {
	folder := "tmp_folder_delete"
	defer os.RemoveAll(folder)

	tr, err := filewriter.NewWriter(folder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	// Create a file first
	fileName := "test-file.txt"
	content := []byte("test content")
	filePath, err := tr.Write(fileName, content, filewriter.PersistentFile)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("test file should exist before deletion")
	}

	// Delete the file
	err = tr.Delete(fileName)
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	// Verify file is deleted
	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("file should be deleted after Delete()")
	}
}

func TestWriterDeleteFileNotExists(t *testing.T) {
	folder := "tmp_folder_delete_not_exists"
	defer os.RemoveAll(folder)

	tr, err := filewriter.NewWriter(folder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	// Try to delete non-existent file - should not error
	err = tr.Delete("non-existent-file.txt")
	if err != nil {
		t.Errorf("Delete() on non-existent file should not error, got: %v", err)
	}
}

func TestWriterDeleteInvalidPath(t *testing.T) {
	folder := "tmp_folder_delete_invalid"
	defer os.RemoveAll(folder)

	tr, err := filewriter.NewWriter(folder)
	if err != nil {
		t.Fatalf("failed creating writer error = %v", err)
	}

	// Try to delete with invalid characters (this should fail on most systems)
	err = tr.Delete("invalid\x00file.txt")
	if err == nil {
		t.Errorf("Delete() with invalid filename should error")
	}
}
