package nutanix

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"syscall"

	"github.com/nutanix-cloud-native/prism-go-client/converged"
	v4converged "github.com/nutanix-cloud-native/prism-go-client/converged/v4"
)

// IsRetryableValidationError separates unavailable Prism APIs from invalid cluster configuration.
func IsRetryableValidationError(err error) bool {
	if isPermanentPrismError(err) {
		return false
	}
	if isRetryablePrismError(err) || isInterruptedRequest(err) {
		return true
	}
	if isPermanentConnectionError(err) {
		return false
	}

	var requestError *url.Error
	if errors.As(err, &requestError) {
		return isHTTPRequestFailure(requestError)
	}
	return isTransientNetworkError(err)
}

func isPermanentPrismError(err error) bool {
	if errors.Is(err, converged.ErrNotFound) {
		return true
	}
	for err != nil {
		status, _ := v4converged.GetStatusAndBody(err)
		fields := strings.Fields(status)
		if len(fields) > 0 {
			switch fields[0] {
			case "401", "403", "404":
				return true
			}
		}
		err = errors.Unwrap(err)
	}
	return false
}

func isRetryablePrismError(err error) bool {
	if errors.Is(err, converged.ErrRateLimit) || errors.Is(err, converged.ErrInternal) {
		return true
	}
	// The converged SDK leaves HTTP 408 unclassified, but retains the SDK status.
	for err != nil {
		status, _ := v4converged.GetStatusAndBody(err)
		fields := strings.Fields(status)
		if len(fields) > 0 && fields[0] == "408" {
			return true
		}
		err = errors.Unwrap(err)
	}
	return false
}

func isInterruptedRequest(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) ||
		errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF)
}

func isPermanentConnectionError(err error) bool {
	var unknownAuthority x509.UnknownAuthorityError
	var invalidCertificate x509.CertificateInvalidError
	var wrongHostname x509.HostnameError
	if errors.As(err, &unknownAuthority) || errors.As(err, &invalidCertificate) || errors.As(err, &wrongHostname) {
		return true
	}
	var dnsError *net.DNSError
	return errors.As(err, &dnsError) && dnsError.IsNotFound && !dnsError.IsTimeout && !dnsError.IsTemporary
}

func isHTTPRequestFailure(err *url.Error) bool {
	requestURL, parseErr := url.Parse(err.URL)
	if parseErr != nil || requestURL.Host == "" ||
		(requestURL.Scheme != "https" && requestURL.Scheme != "http") {
		return false
	}
	// The SDK's exhausted HTTP retries retain url.Error even when the inner cause is opaque.
	switch strings.ToUpper(err.Op) {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
		return true
	}
	return false
}

func isTransientNetworkError(err error) bool {
	var networkError net.Error
	return errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) ||
		errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ENETUNREACH) ||
		(errors.As(err, &networkError) && networkError.Timeout())
}
