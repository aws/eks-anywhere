package test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func NewHTTPServerForFile(t *testing.T, filePath string) *httptest.Server {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed reading file [%s] to serve from http: %s", filePath, err)
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(fileContent); err != nil {
			t.Errorf("Failed writing response to http request: %s", err)
		}
	}))
	t.Cleanup(func() { ts.Close() })
	return ts
}

// NewHTTPSServerForFile creates an HTTPS server that always serves the content of the
// given file.
func NewHTTPSServerForFile(t *testing.T, filePath string) *httptest.Server {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("Failed reading file [%s] to serve from http: %s", filePath, err)
	}
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, err := w.Write(fileContent); err != nil {
			t.Errorf("Failed writing response to http request: %s", err)
		}
	}))
	t.Cleanup(func() { ts.Close() })
	return ts
}
