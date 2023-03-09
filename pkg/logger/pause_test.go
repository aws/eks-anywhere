package logger_test

import (
	"bytes"
	"testing"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func TestPausibleWriter(t *testing.T) {
	var dst bytes.Buffer
	sink := logger.NewPausableWriter(&dst)

	_, err := sink.Write([]byte("hello"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dst.String() != "hello" {
		t.Fatalf("Expected 'hello', got '%s'", dst.String())
	}

	resume := sink.Pause()

	_, err = sink.Write([]byte("world"))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if dst.String() != "hello" {
		t.Fatalf("Expected 'hello', got '%s'", dst.String())
	}

	err = resume()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if dst.String() != "helloworld" {
		t.Fatalf("Expected 'helloworld'; got '%s'", dst.String())
	}
}
