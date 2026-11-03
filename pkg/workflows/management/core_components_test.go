package management

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func TestRetryAdmissionConnection(t *testing.T) {
	for _, message := range []string{
		`failed calling webhook "mutation.cluster": failed to call webhook: connect: connection refused`,
		`failed calling webhook "mutation.cluster": no endpoints available for service "webhook"`,
	} {
		t.Run(message, func(t *testing.T) {
			attempts := 0
			err := retryAdmissionConnection(context.Background(), time.Millisecond, time.Second, func(context.Context) error {
				attempts++
				if attempts == 1 {
					return errors.New(message)
				}
				return nil
			})
			if err != nil || attempts != 2 {
				t.Fatalf("error = %v, attempts = %d", err, attempts)
			}
		})
	}
}

func TestRetryAdmissionConnectionPermanentErrors(t *testing.T) {
	for _, message := range []string{
		`failed calling webhook "mutation.cluster": x509: certificate signed by unknown authority`,
		`admission webhook "validation.cluster" denied the request: invalid cluster`,
		`clusters.anywhere.eks.amazonaws.com is forbidden`,
		`dial tcp api:443: connect: connection refused`,
		`failed calling webhook "mutation.cluster": TLS handshake timeout`,
		`failed calling webhook "mutation.cluster": x509: unknown authority; previous dial: connect: connection refused`,
		`failed calling webhook "mutation.cluster": tls: bad certificate; previous dial: connect: connection refused`,
	} {
		t.Run(message, func(t *testing.T) {
			expected := errors.New(message)
			attempts := 0
			err := retryAdmissionConnection(context.Background(), time.Millisecond, time.Second, func(context.Context) error {
				attempts++
				return expected
			})
			if !errors.Is(err, expected) || attempts != 1 {
				t.Fatalf("error = %v, attempts = %d", err, attempts)
			}
		})
	}
}

func TestRetryAdmissionConnectionTypedPermanentErrors(t *testing.T) {
	message := `failed calling webhook "mutation.cluster": connect: connection refused`
	for _, expected := range []error{
		apierrors.NewForbidden(schema.GroupResource{Group: "example.test", Resource: "clusters"}, "cluster", errors.New(message)),
		apierrors.NewInvalid(schema.GroupKind{Group: "example.test", Kind: "Cluster"}, "cluster",
			field.ErrorList{field.Invalid(field.NewPath("spec"), "dummy", message)}),
		apierrors.NewUnauthorized(message),
		fmt.Errorf("updating management version: %w",
			apierrors.NewForbidden(schema.GroupResource{Resource: "clusters"}, "cluster", errors.New(message))),
	} {
		t.Run(expected.Error(), func(t *testing.T) {
			attempts := 0
			err := retryAdmissionConnection(context.Background(), time.Millisecond, time.Second, func(context.Context) error {
				attempts++
				return expected
			})
			if !errors.Is(err, expected) || attempts != 1 {
				t.Fatalf("error = %v, attempts = %d", err, attempts)
			}
		})
	}
}

func TestRetryAdmissionConnectionTimeout(t *testing.T) {
	message := `failed calling webhook "mutation.cluster": connect: connection refused`
	attempts := 0
	err := retryAdmissionConnection(context.Background(), time.Millisecond, 10*time.Millisecond, func(context.Context) error {
		attempts++
		return errors.New(message)
	})
	if !errors.Is(err, context.DeadlineExceeded) || !strings.Contains(err.Error(), message) || attempts < 2 {
		t.Fatalf("error = %v, attempts = %d", err, attempts)
	}
}

func TestRetryAdmissionConnectionCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	err := retryAdmissionConnection(ctx, time.Hour, time.Hour, func(context.Context) error {
		attempts++
		cancel()
		return errors.New(`failed calling webhook "mutation.cluster": connect: connection refused`)
	})
	if !errors.Is(err, context.Canceled) || attempts != 1 {
		t.Fatalf("error = %v, attempts = %d", err, attempts)
	}
}
