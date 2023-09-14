package curatedpackages_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestTestRegistry(t *testing.T) {
	err := curatedpackages.TestRegistryWithAuthToken("authToken", "registry_url", func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	})
	if err != nil {
		t.Errorf("Registry is good, but error has been returned %v\n", err)
	}

	err = curatedpackages.TestRegistryWithAuthToken("authToken", "registry_url", func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 404,
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	})
	if err != nil {
		t.Errorf("Registry is good, but error has been returned %v\n", err)
	}

	err = curatedpackages.TestRegistryWithAuthToken("authToken", "registry_url", func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewReader(nil)),
		}, nil
	})
	if err == nil {
		t.Errorf("Error should have been returned")
	}
}
