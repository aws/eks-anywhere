package test

import (
	"errors"
	"io"
	"io/fs"
	"os"
	"testing"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

// WithFakeFile returns a throwaway file in a test-specific directory.
//
// The file is automatically closed and removed when the test ends.
func WithFakeFile(t *testing.T) (f *os.File) {
	f, err := os.CreateTemp(t.TempDir(), "fake-file")
	if err != nil {
		t.Fatalf("opening throwaway file: %s", err)
	}

	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Logf("closing throwaway file %q: %s", f.Name(), err)
		}
	})

	return f
}

// WithFakeFileContents returns a throwaway file containing some data.
//
// The file is automatically closed and removed when the test ends.
func WithFakeFileContents(t *testing.T, r io.Reader) (f *os.File) {
	f = WithFakeFile(t)
	_, err := io.Copy(f, r)
	if err != nil {
		t.Fatalf("copying contents into fake file %q: %s", f.Name(), err)
	}

	return f
}

// RemoveFileIfExists is a helper for ValidateFilename tests.
func RemoveFileIfExists(t *testing.T, filename string) {
	if err := os.Remove(filename); err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			t.Fatalf("removing file %q: %s", filename, err)
		}
	}
}

// UseEnvTest sets up the controller-runtime EnvTest framework.
//
// The test will be skipped if EnvTest framework isn't detected.
//
// EnvTest provides fake k8s control plane components for testing
// purposes. The process of bringing up and tearing down the EnvTest framework
// involves running a few binaries, and is not integrated into a vanilla "go
// test" run, but rather can be run via the unit-test target in Makefile. See
// https://book.kubebuilder.io/reference/envtest.html for details.
//
// TODO: What could be done to integrate EnvTest with go test, so that "go
// test" would work?
func UseEnvTest(t *testing.T) *rest.Config {
	// Detect if EnvTest has been set up.
	if os.Getenv("KUBEBUILDER_ASSETS") == "" {
		// By skipping this test, we allow traditional runs of go test from
		// the command-line (or your editor) to work.
		t.Skip("no EnvTest assets found in KUBEBUILDER_ASSETS")
	}

	testEnv := &envtest.Environment{}
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatalf("setting up EnvTest framework: %s", err)
	}
	t.Cleanup(func() {
		if err := testEnv.Stop(); err != nil {
			t.Logf("stopping EnvTest framework: %s", err)
		}
	})

	return cfg
}
