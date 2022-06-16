package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
)

type DefaultTlsValidator struct{}

type TlsValidator interface {
	ValidateCert(host, port, caCertContent string) error
	HasSelfSignedCert(host, port string) (bool, error)
}

func NewTlsValidator() TlsValidator {
	return &DefaultTlsValidator{}
}

// HasSelfSignedCert determines whether the url is using self-signed certs or not
func (tv *DefaultTlsValidator) HasSelfSignedCert(host, port string) (bool, error) {
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}

	_, err := tls.Dial("tcp", net.JoinHostPort(host, port), conf)
	if err != nil {
		// If the error is x509.UnknownAuthorityError, this means the url is using self-signed certs
		if err.Error() == (x509.UnknownAuthorityError{}).Error() {
			return true, nil
		}
		return false, err
	}
	return false, nil
}

// ValidateCert parses the cert, ensures that the cert format is valid and verifies that the cert is valid for the url
func (tv *DefaultTlsValidator) ValidateCert(host, port, caCertContent string) error {
	// Validates that the cert format is valid
	block, _ := pem.Decode([]byte(caCertContent))
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}
	providedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	roots := x509.NewCertPool()
	roots.AddCert(providedCert)
	conf := &tls.Config{
		InsecureSkipVerify: false,
		RootCAs:            roots,
	}
	// Verifies that the cert is valid by making a connection to the endpoint
	endpoint := net.JoinHostPort(host, port)
	conn, err := tls.Dial("tcp", endpoint, conf)
	if err != nil {
		return fmt.Errorf("verifying tls connection to host with custom CA: %v", err)
	}
	if err = conn.Close(); err != nil {
		return fmt.Errorf("closing tls connection to %v: %v", endpoint, err)
	}
	return nil
}
