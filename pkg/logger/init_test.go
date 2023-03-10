package logger_test

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func TestInit(t *testing.T) {
	tdir := t.TempDir()
	defer os.RemoveAll(tdir)

	logFile := filepath.Join(tdir, "test.log")

	err := logger.Init(logger.Options{
		OutputFilePath: logFile,
	})
	if err != nil {
		t.Fatal(err)
	}

	message := "log me"
	logger.Info(message)

	// Opening the file validates it exists.
	f, err := os.Open(logFile)
	if err != nil {
		t.Fatalf("Error opening log file: %v", err)
	}
	defer f.Close()

	// Ensure we're writing data to the logger from package functions.
	buf, err := io.ReadAll(f)
	if err != nil {
		t.Fatalf("Reading log file: %v", err)
	}

	if !bytes.Contains(buf, []byte(message)) {
		t.Fatalf("Log file does not contain expected message: %s", message)
	}
}
