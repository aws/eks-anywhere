package crypto_test

import (
	"testing"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/crypto"
)

const (
	endpoint         = "unit-test.local"
	invalid_endpoint = "invalid-endpoint.local"
	port             = constants.DefaultHttpsPort
	/*
		This certificate was generated using the following commands and is valid only for `unit-test.local`
		openssl genrsa -out ca.key 2048
		openssl req -new -x509 -days 3650 -key ca.key -out ca.crt
		openssl req -newkey rsa:2048 -nodes -keyout server.key -out server.csr
		openssl x509 -req -extfile <(printf "subjectAltName=DNS:unit-test.local") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
	*/
	cert = `
-----BEGIN CERTIFICATE-----
MIID/jCCAuagAwIBAgIUO/ncrEaWxLUqZ8IioBVCRl1P2R4wDQYJKoZIhvcNAQEL
BQAwgagxCzAJBgNVBAYTAlVTMRMwEQYDVQQIDApXYXNoaW5ndG9uMRAwDgYDVQQH
DAdTZWF0dGxlMRwwGgYDVQQKDBNBbWF6b24gV2ViIFNlcnZpY2VzMRUwEwYDVQQL
DAxFS1MgQW55d2hlcmUxGDAWBgNVBAMMD3VuaXQtdGVzdC5sb2NhbDEjMCEGCSqG
SIb3DQEJARYUdGVzdEB1bml0LXRlc3QubG9jYWwwHhcNMjExMTE2MjMwNzUzWhcN
MjIxMTE2MjMwNzUzWjCBqDELMAkGA1UEBhMCVVMxEzARBgNVBAgMCldhc2hpbmd0
b24xEDAOBgNVBAcMB1NlYXR0bGUxHDAaBgNVBAoME0FtYXpvbiBXZWIgU2Vydmlj
ZXMxFTATBgNVBAsMDEVLUyBBbnl3aGVyZTEYMBYGA1UEAwwPdW5pdC10ZXN0Lmxv
Y2FsMSMwIQYJKoZIhvcNAQkBFhR0ZXN0QHVuaXQtdGVzdC5sb2NhbDCCASIwDQYJ
KoZIhvcNAQEBBQADggEPADCCAQoCggEBALC+5yZrxn8sy7WilquxsjRjqCzaUoio
i31TlU1lRCI1HhgCjAE5xvMMS1vd1lhXxx7VdOS4b5UO+S+IAOjWWTQDfX7H+hOm
AIAU45ejaUtQDZ7hjdHXyfIOhi5Qb+D4ZLiMAQEe/EHLpB0dxBu+KD0nlBxHKUQY
HT1s41u9J4gOjhB+oVAQZmWvoTt0v5iPrljOfjsHPV4HqDUxPh9ngeL3a7AkMxIG
Nf4nh7hqFKKwGMgifAXG3k4aec/gOIKBEt9Ns43uTn45zKHkL2C4NHTGjFGWlnT8
ixxUW3bXFTI6LjKzllprYimGaMiyMPSOEtXOFV2xHedv39Qaq6yp4/sCAwEAAaMe
MBwwGgYDVR0RBBMwEYIPdW5pdC10ZXN0LmxvY2FsMA0GCSqGSIb3DQEBCwUAA4IB
AQCUHw792stgHCPJ6qYD6ij1ehp4yAdXKdaOykTwOxvPvcjdKxnwBYJ7feU+Wh6B
fauk1tpUp6tEF/FzFXdGoYfMvMJtHeS57dFf/mx9uUgHXnhgIYFDe4bM+4W2LHHC
mbvwOYSLAyPfjhihryCRSmIk2X+JYhTSnTXUwKkacEMn3BiEyTfZG9pzr/aIpsIE
e/rwWa9a4BdrhqTBK6VWtTvNYbRaDVP8cbQPVl0qdIyqbTPI/QrITGchY2Qk/eoS
zwaAnAW1ZiriAbeFx+xOaO1fETVSm+5Poyl97r5Mmu97+3IpoWHFPO2z4Os9vn3q
XsKvL2lz2uQY+ZbrfvrL20p2
-----END CERTIFICATE-----
`
	invalid_cert = `
-----BEGIN CERTIFICATE-----
invalidCert
-----END CERTIFICATE-----
`
)

func TestValidateCertValidCert(t *testing.T) {
	tv := crypto.NewTlsValidator(endpoint, port)
	if err := tv.ValidateCert(cert); err != nil {
		t.Fatalf("Failed to validate cert: %v", err)
	}
}

func TestValidateCertInvalidEndpoint(t *testing.T) {
	tv := crypto.NewTlsValidator(invalid_endpoint, port)
	if err := tv.ValidateCert(cert); err == nil {
		t.Fatalf("Certificate validation passed for invalid endpoint")
	}
}

func TestValidateCertInvalidCert(t *testing.T) {
	tv := crypto.NewTlsValidator(endpoint, port)
	if err := tv.ValidateCert(invalid_cert); err == nil {
		t.Fatalf("Certificate validation passed for invalid cert")
	}
}
