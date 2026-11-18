package nutanix

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"syscall"
	"testing"

	"github.com/nutanix-cloud-native/prism-go-client/converged"
	v4converged "github.com/nutanix-cloud-native/prism-go-client/converged/v4"
	clusterClient "github.com/nutanix/ntnx-api-golang-clients/clustermgmt-go-client/v4/client"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"k8s.io/utils/ptr"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	mocknutanix "github.com/aws/eks-anywhere/pkg/providers/nutanix/mocks"
)

func TestIsRetryableValidationError(t *testing.T) {
	requestError := func(cause error) error {
		return &url.Error{Op: "Get", URL: "https://pc.example.test:9440/api/clustermgmt/v4.2/config/clusters", Err: cause}
	}
	sdkError := func(status string) error {
		return v4converged.CategoriseFromOpenAPI(clusterClient.GenericOpenAPIError{
			Status: status,
			Body:   []byte(`{"message":"test error"}`),
		})
	}
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil"},
		{name: "invalid configuration", err: errors.New("missing image")},
		{name: "not found", err: fmt.Errorf("image lookup: %w", converged.ErrNotFound)},
		{name: "not found dominates transport cause", err: &converged.APIError{Kind: converged.ErrNotFound, Cause: requestError(context.DeadlineExceeded)}},
		{name: "unauthorized", err: sdkError("401 Unauthorized")},
		{name: "forbidden", err: sdkError("403 Forbidden")},
		{name: "resource not found", err: sdkError("404 Not Found")},
		{name: "unauthorized dominates interruption", err: sdkStatusCauseError{Status: "401 Unauthorized", cause: context.DeadlineExceeded}},
		{name: "forbidden dominates interruption", err: sdkStatusCauseError{Status: "403 Forbidden", cause: io.ErrUnexpectedEOF}},
		{name: "HTTP not found dominates request", err: sdkStatusCauseError{Status: "404 Not Found", cause: requestError(io.EOF)}},
		{name: "invalid request", err: sdkError("400 Bad Request")},
		{name: "request timeout", err: sdkError("408 Request Timeout"), want: true},
		{name: "rate limited", err: sdkError("429 Too Many Requests"), want: true},
		{name: "server failure", err: sdkError("500 Internal Server Error"), want: true},
		{name: "unavailable", err: sdkError("503 Service Unavailable"), want: true},
		{name: "deadline", err: context.DeadlineExceeded, want: true},
		{name: "canceled request", err: context.Canceled, want: true},
		{name: "connection reset", err: requestError(syscall.ECONNRESET), want: true},
		{name: "connection refused", err: syscall.ECONNREFUSED, want: true},
		{name: "incomplete response", err: requestError(io.ErrUnexpectedEOF), want: true},
		{name: "SDK response body interrupted", err: v4converged.CategoriseFromOpenAPI(io.ErrUnexpectedEOF), want: true},
		{name: "SDK retry exhaustion", err: requestError(errors.New("GET request giving up after 6 attempt(s)")), want: true},
		{name: "SDK categorized retry exhaustion", err: v4converged.CategoriseFromOpenAPI(requestError(errors.New("GET request giving up after 6 attempt(s)"))), want: true},
		{name: "unclassified request cause", err: &converged.APIError{Cause: requestError(context.DeadlineExceeded)}, want: true},
		{name: "opaque validation text", err: errors.New("GET request giving up after 6 attempt(s)")},
		{name: "certificate trust", err: requestError(x509.UnknownAuthorityError{})},
		{name: "certificate hostname", err: requestError(x509.HostnameError{})},
		{name: "invalid certificate", err: requestError(x509.CertificateInvalidError{})},
		{name: "unknown hostname", err: requestError(&net.DNSError{IsNotFound: true})},
		{name: "DNS timeout", err: requestError(&net.DNSError{IsTimeout: true}), want: true},
		{name: "malformed URL", err: &url.Error{Op: "parse", URL: "%", Err: errors.New("invalid escape")}},
		{name: "unsupported protocol", err: &url.Error{Op: "Get", URL: "ftp://pc.example.test", Err: errors.New("unsupported protocol")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, IsRetryableValidationError(tt.err))
			if tt.err != nil {
				wrapped := fmt.Errorf("validate machine: %w", fmt.Errorf("Prism lookup: %w", tt.err))
				require.Equal(t, tt.want, IsRetryableValidationError(wrapped))
				require.Equal(t, tt.want, IsRetryableValidationError(wrapped), "classification must not consume or modify the error")
			}
		})
	}
}

type sdkStatusCauseError struct {
	Status string
	cause  error
}

func (e sdkStatusCauseError) Error() string { return e.Status }
func (e sdkStatusCauseError) Unwrap() error { return e.cause }

func TestValidationPreservesPrismRequestError(t *testing.T) {
	ctx := context.Background()
	ctrl := gomock.NewController(t)
	client := mocknutanix.NewMockClient(ctrl)
	cause := &url.Error{Op: "Get", URL: "https://pc.example.test:9440/api/clustermgmt/v4.2/config/clusters", Err: context.DeadlineExceeded}
	client.EXPECT().ListClusters(ctx, gomock.Any()).Return(nil, cause).Times(2)

	validator := &Validator{}
	credentialErr := validator.validateCredentials(ctx, client)
	require.ErrorIs(t, credentialErr, cause)
	require.True(t, IsRetryableValidationError(credentialErr))

	clusterErr := validator.validateClusterConfig(ctx, client, anywherev1.NutanixResourceIdentifier{
		Type: anywherev1.NutanixIdentifierName,
		Name: ptr.To("test-cluster"),
	})
	require.ErrorIs(t, clusterErr, cause)
	require.True(t, IsRetryableValidationError(clusterErr))
}
