package test

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"regexp"
	"testing"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

var updateGoldenFiles = flag.Bool("update", false, "update golden files")

func AssertFilesEquals(t *testing.T, gotPath, wantPath string) {
	gotFile := ReadFile(t, gotPath)
	processUpdate(t, wantPath, gotFile)
	wantFile := ReadFile(t, wantPath)

	if gotFile != wantFile {
		cmd := exec.Command("diff", wantPath, gotPath)
		result, err := cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					t.Fatalf("Results diff expected actual:\n%s", string(result))
				}
			}
		}
		t.Fatalf("Files are different got =\n  %s \n want =\n %s\n%s", gotFile, wantFile, err)
	}
}

func AssertContentToFile(t *testing.T, gotContent, wantFile string) {
	if wantFile == "" && gotContent == "" {
		return
	}

	processUpdate(t, wantFile, gotContent)

	fileContent := ReadFile(t, wantFile)

	if gotContent != fileContent {
		cmd := exec.Command("diff", wantFile, "-")
		cmd.Stdin = bytes.NewReader([]byte(gotContent))
		result, err := cmd.Output()
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				if exitError.ExitCode() == 1 {
					t.Fatalf("Results diff expected actual for %s:\n%s", wantFile, string(result))
				}
			}
		}
		t.Fatalf("Content doesn't match file got =\n%s\n\n\nwant =\n%s\n", gotContent, fileContent)
	}
}

func processUpdate(t *testing.T, filePath, content string) {
	if *updateGoldenFiles {
		if err := ioutil.WriteFile(filePath, []byte(content), 0o644); err != nil {
			t.Fatalf("failed to update golden file %s: %v", filePath, err)
		}
		log.Printf("Golden file updated: %s", filePath)
	}
}

func ReadFile(t *testing.T, file string) string {
	bytesRead, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatalf("File [%s] reading error in test: %v", file, err)
	}

	return string(bytesRead)
}

func NewWriter(t *testing.T) (dir string, writer filewriter.FileWriter) {
	dir, err := ioutil.TempDir(".", SanitizePath(t.Name())+"-")
	if err != nil {
		t.Fatalf("error setting up folder for test: %v", err)
	}

	t.Cleanup(cleanupDir(t, dir))
	writer, err = filewriter.NewWriter(dir)
	if err != nil {
		t.Fatalf("error creating writer with folder for test: %v", err)
	}
	return dir, writer
}

func cleanupDir(t *testing.T, dir string) func() {
	return func() {
		if !t.Failed() {
			os.RemoveAll(dir)
		}
	}
}

var sanitizePathChars = regexp.MustCompile(`[^\w-]`)

const sanitizePathReplacementChar = "_"

// SanitizePath sanitizes s so its usable as a path name. For safety, it assumes all characters that are not
// A-Z, a-z, 0-9, _ or - are illegal and replaces them with _.
func SanitizePath(s string) string {
	return sanitizePathChars.ReplaceAllString(s, sanitizePathReplacementChar)
}
