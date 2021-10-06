package crypto

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type DefaultTlsValidator struct {
	cert string
	url  string
}

type TlsValidator interface {
	ValidateCert() error
	HasSelfSignedCert() (bool, error)
}

func NewTlsValidator(cert, url string) TlsValidator {
	return &DefaultTlsValidator{
		cert: cert,
		url:  url,
	}
}

// getBaseCertFromUrl connects to the url and retrieves the base certificate
func (tv *DefaultTlsValidator) getBaseCertFromUrl() (*x509.Certificate, error) {
	conf := &tls.Config{
		// This is needed so to make a connection to the url even if it is using self-signed certs
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", tv.url+":443", conf)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// ConnectionState().PeerCertificates returns the entire certificate chain but we only need the base
	return conn.ConnectionState().PeerCertificates[0], nil
}

// HasSelfSignedCert determines whether the url is using self-signed certs or not
func (tv *DefaultTlsValidator) HasSelfSignedCert() (bool, error) {
	conf := &tls.Config{
		InsecureSkipVerify: false,
	}

	_, err := tls.Dial("tcp", tv.url+":443", conf)
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
func (tv *DefaultTlsValidator) ValidateCert() error {
	cert, err := tv.getBaseCertFromUrl()
	if err != nil {
		return err
	}

	// Validates that the cert format is valid
	block, _ := pem.Decode([]byte(tv.cert))
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}
	providedCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err)
	}

	// Validates that the provided cert matches the one returned by the url
	if !cert.Equal(providedCert) {
		return fmt.Errorf("certificate mismatch error: provided certificate does not match certificate received from %s", tv.url)
	}

	roots := x509.NewCertPool()
	roots.AddCert(providedCert)
	opts := x509.VerifyOptions{
		DNSName: tv.url,
		Roots:   roots,
	}

	// Verifies that the cert is valid
	_, err = providedCert.Verify(opts)
	if err != nil {
		return fmt.Errorf("failed to verify certificate: %v", err)
	}
	return nil
}
