package framework

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

const (
	webhookConnectionRefused = `Error: Error from server (InternalError): error when creating "STDIN": Internal error occurred: failed calling webhook "vpackage.kb.io": failed to call webhook: Post "https://eks-anywhere-packages-eks-anywhere-packages-webhook-service.eksa-packages.svc:443/validate-packages-eks-amazonaws-com-v1alpha1-package?timeout=10s": dial tcp 10.97.86.87:443: connect: connection refused`
	webhookNoEndpoints       = `Internal error occurred: failed calling webhook "vpackage.kb.io": failed to call webhook: Post "https://eks-anywhere-packages-eks-anywhere-packages-webhook-service.eksa-packages.svc:443/validate-packages-eks-amazonaws-com-v1alpha1-package?timeout=10s": no endpoints available for service "eks-anywhere-packages-eks-anywhere-packages-webhook-service"`
	webhookDenied            = `Error from server: error when creating "STDIN": admission webhook "vpackage.kb.io" denied the request: getting PackageBundle: no activeBundle set in PackageBundleController`
)

func TestIsPackagesWebhookUnavailable(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   bool
	}{
		{name: "connection refused", output: webhookConnectionRefused, want: true},
		{name: "no endpoints", output: webhookNoEndpoints, want: true},
		{name: "denied by the webhook", output: webhookDenied, want: false},
		{name: "other webhook", output: `failed calling webhook "validation.cluster.x-k8s.io": connection refused`, want: false},
		{name: "unrelated error", output: "Error: timed out waiting for the condition", want: false},
		{name: "empty", output: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isPackagesWebhookUnavailable(tt.output); got != tt.want {
				t.Errorf("isPackagesWebhookUnavailable() = %t, want %t", got, tt.want)
			}
		})
	}
}

type retryLogT struct {
	T
	logs []string
}

func (t *retryLogT) Logf(format string, args ...any) {
	t.logs = append(t.logs, fmt.Sprintf(format, args...))
}

func TestRetryWhilePackagesWebhookUnavailable(t *testing.T) {
	defer func(timeout, interval time.Duration) {
		packagesWebhookRetryTimeout, packagesWebhookRetryInterval = timeout, interval
	}(packagesWebhookRetryTimeout, packagesWebhookRetryInterval)
	packagesWebhookRetryInterval = time.Millisecond

	unavailable := errors.New(webhookConnectionRefused)
	denied := errors.New(webhookDenied)

	tests := []struct {
		name      string
		timeout   time.Duration
		results   []error
		wantErr   error
		wantCalls int
	}{
		{name: "succeeds first time", timeout: time.Minute, results: []error{nil}, wantCalls: 1},
		{name: "succeeds once the webhook is back", timeout: time.Minute, results: []error{unavailable, unavailable, nil}, wantCalls: 3},
		{name: "other error is not retried", timeout: time.Minute, results: []error{denied}, wantErr: denied, wantCalls: 1},
		{name: "webhook comes back then rejects", timeout: time.Minute, results: []error{unavailable, denied}, wantErr: denied, wantCalls: 2},
		{name: "gives up after the timeout", timeout: 0, results: []error{unavailable}, wantErr: unavailable, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			packagesWebhookRetryTimeout = tt.timeout
			logT := &retryLogT{}
			e := &ClusterE2ETest{T: logT}
			calls := 0
			err := e.retryWhilePackagesWebhookUnavailable(func() (string, error) {
				res := tt.results[calls]
				calls++
				if res == nil {
					return "", nil
				}
				return res.Error(), res
			})
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("error = %v, want %v", err, tt.wantErr)
			}
			if calls != tt.wantCalls {
				t.Errorf("calls = %d, want %d", calls, tt.wantCalls)
			}
			if len(logT.logs) != tt.wantCalls-1 && tt.wantErr != unavailable {
				t.Errorf("logged %d retries, want %d", len(logT.logs), tt.wantCalls-1)
			}
		})
	}
}
