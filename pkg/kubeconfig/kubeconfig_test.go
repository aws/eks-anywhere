package kubeconfig

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestValidate(t *testing.T) {
	t.Run("reports errors from validator", func(t *testing.T) {
		withValidationError(t, testError)

		if err := Validate(&bytes.Buffer{}); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("returns nil when valid", func(t *testing.T) {
		withValidationError(t, nil)

		if err := Validate(&bytes.Buffer{}); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
	})
}

func TestValidateFile(t *testing.T) {
	t.Run("reports errors for files that don't exist", func(t *testing.T) {
		doesntExist := filepath.Join(t.TempDir(), "does-not-exist")
		err := ValidateFile(doesntExist)
		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("expected fs.IsNotExist, got %s", err)
		}
	})

	t.Run("reports errors from validator", func(t *testing.T) {
		withValidationError(t, testError)

		if err := ValidateFile(withFakeFile(t).Name()); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("returns nil if valid", func(t *testing.T) {
		withValidationError(t, nil)

		if err := ValidateFile(withFakeFile(t).Name()); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
	})
}

// testError is... a test error!
var testError = fmt.Errorf("test error")

// withValidationError is a helper for setting the validator for a test, and
// automatically reverting the change afterward.
//
// This is *NOT* thread safe! (Because it manipulates package-level variables).
var withValidationError = func(t *testing.T, err error) {
	oldValidator := validator
	validator = func([]byte) error {
		return err
	}
	t.Cleanup(func() {
		validator = oldValidator
	})
}

// withFakeFile returns a throwaway file in a test-specific directory.
//
// The file is automatically closed and removed when the test ends.
func withFakeFile(t *testing.T) (f *os.File) {
	f, err := ioutil.TempFile(t.TempDir(), "fake-file")
	if err != nil {
		t.Fatalf("opening temp file: %s", err)
	}

	t.Cleanup(func() {
		f.Close()
	})

	return f
}
