package curatedpackages_test

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/aws/eks-anywhere/pkg/curatedpackages"
)

func TestValidateNoKubeVersionWhenClusterSucceeds(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("", "morby")
	if err != nil {
		t.Errorf("empty kubeVersion allowed when cluster specified")
	}
}

func TestValidateKubeVersionWhenClusterFails(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("1.21", "morby")
	if err == nil {
		t.Errorf("not both kube-version and cluster")
	}
}

func TestValidateKubeVersionWhenNoClusterFails(t *testing.T) {
	err := curatedpackages.ValidateKubeVersion("", "")
	if err == nil {
		t.Errorf("must specify cluster or kubeversion")
	}
}

func TestValidateKubeVersionWhenRegistrySucceeds(t *testing.T) {
	kubeVersion := "1.21"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, "")
	if err != nil {
		t.Errorf("Registry with %s should succeed", kubeVersion)
	}
}

func TestValidateKubeVersionWhenInvalidVersionFails(t *testing.T) {
	kubeVersion := "1.2.3"
	err := curatedpackages.ValidateKubeVersion(kubeVersion, "")
	if err == nil {
		t.Errorf("Registry with %s should fail", kubeVersion)
	}
}

type mockHTTPClient struct {
	doFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if m.doFunc != nil {
		return m.doFunc(req)
	}
	return &http.Response{}, nil
}

func TestValidateECRAuthToken(t *testing.T) {
	tests := []struct {
		statusCode int
		returnErr  bool
	}{
		{
			statusCode: 200,
			returnErr:  false,
		},
		{
			statusCode: 400,
			returnErr:  true,
		},
		{
			statusCode: 401,
			returnErr:  true,
		},
		{
			statusCode: 403,
			returnErr:  true,
		},
	}
	for _, test := range tests {
		err := curatedpackages.ValidateECRAuthToken(&mockHTTPClient{
			doFunc: func(req *http.Request) (*http.Response, error) {
				return &http.Response{StatusCode: test.statusCode, Body: io.NopCloser(bytes.NewReader(nil))}, nil
			},
		}, "authToken", "registryURL")

		if test.returnErr && err == nil {
			t.Errorf("ValidateECRAuthToken should return err when http response is %s: %v", err, test.statusCode)
		}
	}
}
