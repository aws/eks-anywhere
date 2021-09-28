package crypto

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

type tlsValidator struct{}

type TlsValidator interface {
	ValidateCert(cert string, url string) error
}

func NewTlsValidator() TlsValidator {
	return &tlsValidator{}
}

func (tv *tlsValidator) ValidateCert(certPEM string, url string) error {
	roots := x509.NewCertPool()
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return fmt.Errorf("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %v", err.Error())
	}
	roots.AddCert(cert)

	opts := x509.VerifyOptions{
		DNSName: url,
		Roots:   roots,
	}

	if _, err := cert.Verify(opts); err != nil {
		return fmt.Errorf("failed to verify certificate: %v", err.Error())
	}
	return nil
}
