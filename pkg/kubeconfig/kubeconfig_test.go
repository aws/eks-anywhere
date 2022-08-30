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

	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestValidate(t *testing.T) {
	t.Run("reports errors from validator", func(t *testing.T) {
		v := NewValidatorWithLoader(newTestLoader(testError))

		if err := v.Validate(&bytes.Buffer{}); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("returns nil when valid", func(t *testing.T) {
		v := NewValidatorWithLoader(newTestLoader(nil))

		if err := v.Validate(&bytes.Buffer{}); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
	})
}

func TestValidateFile(t *testing.T) {
	t.Run("reports errors for files that don't exist", func(t *testing.T) {
		doesntExist := filepath.Join(t.TempDir(), "does-not-exist")
		v := NewValidator()
		err := v.ValidateFile(doesntExist)
		if err == nil || !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("expected fs.IsNotExist, got %s", err)
		}
	})

	t.Run("reports errors from validator", func(t *testing.T) {
		v := NewValidatorWithLoader(newTestLoader(testError))

		if err := v.ValidateFile(withFakeFile(t).Name()); err == nil {
			t.Fatalf("expected error, got nil")
		}
	})

	t.Run("returns nil if valid", func(t *testing.T) {
		v := NewValidatorWithLoader(newTestLoader(nil))

		if err := v.ValidateFile(withFakeFile(t).Name()); err != nil {
			t.Fatalf("expected no error, got %s", err)
		}
	})
}

// testError is... a test error!
var testError = fmt.Errorf("test error")

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

type testLoader struct {
	testError error
}

var _ Loader = (*testLoader)(nil)

func newTestLoader(err error) *testLoader {
	return &testLoader{testError: err}
}

func (t *testLoader) Load([]byte) (*clientcmdapi.Config, error) {
	if t.testError != nil {
		return nil, t.testError
	}
	return &clientcmdapi.Config{}, nil
}
