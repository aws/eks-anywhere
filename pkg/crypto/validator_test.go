package crypto_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/aws/eks-anywhere/pkg/crypto"
)

const (
	endpoint        = "127.0.0.1"
	invalidEndpoint = "invalid-endpoint.local"
	/*
		This certificate was generated using the following commands and is valid only for `127.0.0.1`
		openssl genrsa -out ca.key 2048
		openssl req -new -x509 -days 3650 -key ca.key -out ca.crt
		openssl req -newkey rsa:2048 -nodes -keyout server.key -out server.csr
		openssl x509 -req -extfile <(printf "subjectAltName=IP:127.0.0.1") -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
	*/
	caCert = `
-----BEGIN CERTIFICATE-----
MIICrjCCAZYCCQD6stMfKEVAGTANBgkqhkiG9w0BAQsFADAZMQswCQYDVQQGEwJV
UzEKMAgGA1UEAwwBKjAeFw0yMjA2MTUwNzE4MDRaFw0zMjA2MTIwNzE4MDRaMBkx
CzAJBgNVBAYTAlVTMQowCAYDVQQDDAEqMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A
MIIBCgKCAQEAv25FFUfMGKdF9wRb3GBBmE0F0ifMz8nUb52nEX0+BGFg+x6Q8gRD
xTCvCwZoMOo2uRZpaz6KtGQywFuuQ5h8VrYB3WcuDPcpg1OJueWrITRY0NHUBMjh
JbvvqZGs8I/DfKoQOyxcXoeIMfbSW2IHnMeHzTfE9lkGiZatQVqlRQAOkshFIed6
YZ2p3hgxrzkPmTTiDNoQQotluTN0FcNgUwiKYGINVlmH4LaDobranBaMLztulWsK
0tQiDw2e7aaAZmxqaUN6QRMzwc1g2mvj+PXNf/V3/WGymsXUShy9zCx4PiMDuINW
fA/EhAdkZk0MrpYbfk221gSfxxvnd1xFLwIDAQABMA0GCSqGSIb3DQEBCwUAA4IB
AQCwRcvdYlSPYypWh7B1gCvqoq4XNk8YiRPs8X4JLly41uXlgQLfDw9oJUnR0h8c
TNObdEZEdsneinC8Gwg9hscooB2R5iYyqDjST0NAS7jPDDx9x42iRY6naUW7zuVZ
2IC/9ss4BZYlS0zztUOBDax9YOJxV3dwROwnfFqGiiiYqWuDfhlfUaHj6q+NAX9/
xISi7YOoym7wHgMlMIEwGKfMLPOgvZBnU8vMqVqhbxMjAx42VdWyyV8AOcIZxasq
jXcaCczmsh5VMZHNocnTmivmY3KwmUXgC+/4w7E/fRGPblh+2CKxBO6Wlz+ZyN1n
CNgC8QQ1opc4yanhaAZgn/Yd
-----END CERTIFICATE-----
`
	serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDZzMrH+mOavXSn
XHV9lMywwmZARE0TLpuyUTBDuYwVHG13Yx5X7br3uwEtyp2tTUBMHRh+2uVf/7lm
fQfy0AoE5O2qa4EvuOc4zS8eHr1gW4rFC++EzA2FrTMLGczIwKzPPa4VAMteQyCj
LYmfdrOSLSM0Fn7TPFJ1c3DBjj1pY/DdIs/GPItPnL7B2m+gNSQ9E5lRJ/DmP1j8
2+hnScbB5GiLqdZBASd7T+Dwp7gUqZInMKdW6JJjbUY9xNXcHrCS7GINuXHbo5CK
z9Z19nl7+v0fzQbcVP6fH2jWAo7vfFd9IR+Tz/7zJ76UescRS1ONEu6071KfTA7j
fW1dhgqLAgMBAAECggEAdbH0NsK5Boqwuiv9laJORoqWpM4D9ISwQFkdQsvGxjW5
ddWLNSrTaUGV7n/aWycuwrLKZOq3Hvxa3OZd4DnJ4EExqXE0u2wpDwnaF2W3IpX1
VGwRv+pguEcTGUGU5zsvZ0JGizUFsOeHgIaAIzsK6MgZiPFLEa08Rhne6cmKqCMI
tueV7Rqwbbd6+pvnvHFqSrA9KHaJax9+GY0rDr3KlzuJkjK0Wy1D13j/1fkDzalN
opsxF3rShEuTHAVEiwjS+nzD5g6KZ3q2WDTRSKVAbh50rBOEPE9ERwY0vlEQRgKP
4YX/UH1DLDRiDguM2Fr4nhIiTlL98nF6EVLnD7dugQKBgQD+8ZWrvgRlvcqVJmB8
BloKWsoyJp3BgBO77ys0j1fgmQe+TPLpsnVetGmjIK75EQmv/30/lFoLXIM9sfcY
4WgOsnndXFS8gKLfvG8+yMVoSwxNIR/2dNnqZuohqp5LCMlamb7Xvb2rkeSMOx4P
Q0OkTgd0YSl2l7VUwCmBs0VBSwKBgQDas89DAkBVN6q0NRd3iqrjcAN5kKchFO3W
IyDVwrrqhodnx7c7GncPULacMjOs1FPDrhiBoE7x5pDO4v30uVQKrKRZbmN4EHJC
lAUVGKBJFG5RFGzqWrCWf4Hdzynlv5na2SYtGfcmEV5F5ja9ZoDxsOdnjvy6yh0p
E44mtB3TwQKBgGVxASn2EM/e5eXVAF05Ncia+Ytc/DaLXM7RyrI+Oyw+F+uruJgu
jy8gwEvNbHHkSqOCGHcc83tD02DQGE8JGZuHfqAK5hifYq99zhIAVzQ5cGqcPJiX
REJVsuG0fwnCNERdmqdDc136TiNSPpK6JAcTmTnAk3wBv4A6egmGqI7jAoGAG4Vb
BIyo+dBKe+jebh2WCY7T8R1B2sjecP70p9GcYdzR9z5LkXVwHA5FHHy4wfvqGoqy
7MT2ijxAZrhryrrzl3BIMjTQ8Y/oQPaNeS0jJm8avrs6RXdqF1YuSnJCTHYC72Y6
Bpzo2/J9kYA5zTWz7jYbuI1mwj6i0sNyNO6ffkECgYBJ3xZ/h5zDsrbxPDiTsXCs
3liEvmzCibkoCexflz/sN524E1d3F5jj+gsTvcXTvYTaSKLkjRgvtGT9deSCJt76
wKprVZnZJeY+y21aOW1+HM22zfrGJ1GLt1Qjx33GLDzxPJDL9gsHzyRM0S2wLa8u
z9TAYr37qFPTWOJJtmRitg==
-----END PRIVATE KEY-----
`
	serverCert = `
-----BEGIN CERTIFICATE-----
MIICyDCCAbCgAwIBAgIJAJA97n2wo1wBMA0GCSqGSIb3DQEBBQUAMBkxCzAJBgNV
BAYTAlVTMQowCAYDVQQDDAEqMB4XDTIyMDYxNTA3MTkwOVoXDTIzMDYxNTA3MTkw
OVowGTELMAkGA1UEBhMCVVMxCjAIBgNVBAMMASowggEiMA0GCSqGSIb3DQEBAQUA
A4IBDwAwggEKAoIBAQDZzMrH+mOavXSnXHV9lMywwmZARE0TLpuyUTBDuYwVHG13
Yx5X7br3uwEtyp2tTUBMHRh+2uVf/7lmfQfy0AoE5O2qa4EvuOc4zS8eHr1gW4rF
C++EzA2FrTMLGczIwKzPPa4VAMteQyCjLYmfdrOSLSM0Fn7TPFJ1c3DBjj1pY/Dd
Is/GPItPnL7B2m+gNSQ9E5lRJ/DmP1j82+hnScbB5GiLqdZBASd7T+Dwp7gUqZIn
MKdW6JJjbUY9xNXcHrCS7GINuXHbo5CKz9Z19nl7+v0fzQbcVP6fH2jWAo7vfFd9
IR+Tz/7zJ76UescRS1ONEu6071KfTA7jfW1dhgqLAgMBAAGjEzARMA8GA1UdEQQI
MAaHBH8AAAEwDQYJKoZIhvcNAQEFBQADggEBAJTUv9eqrYJnF9ugKpC6QjsxQzJm
6A9Xwc13ORcno+NlYip1t0ITgW2ZebUboqvScy2B+IHM9HyQGkmVu8bXFJQ1clxy
HHA7Xp3xM+zB1KnvQ0ZPbaHnquOl39nULOCTxqSWCN8fqCuA/X3y9UiYL6P1cYRz
GFuoSgMoU+CrhkMQbF8c4dvl1Wr52Re9SYoX0o74N45pXhOMjeqUq2Mauapq7pFE
+21Y+KHEXojBWEufU8F6ihmWDMqu2vbbfODT63qp+6+VAiQOvMrBvsjW7Iq86VLE
8jddCYLRLwQjqaSn6APSSEocuzzP2J6CeGA4mxXHZz5gD22pugKbKLjHxwQ=
-----END CERTIFICATE-----
`
	incorrectCert = `
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
	invalidCert = `
-----BEGIN CERTIFICATE-----
invalidCert
-----END CERTIFICATE-----
`
)

func TestHasSelfSignedCert(t *testing.T) {
	certSvr, err := runTestServerWithCert(serverCert, serverKey)
	if err != nil {
		t.Fatalf("starting test server with certs: %v", err)
	}
	defer certSvr.Close()
	certServerPort := strings.Split(certSvr.URL, ":")[2]
	svr, err := runTestServer()
	if err != nil {
		t.Fatalf("starting test server: %v", err)
	}
	defer svr.Close()
	serverPort := strings.Split(svr.URL, ":")[2]
	tests := []struct {
		testName  string
		endpoint  string
		port      string
		wantCert  bool
		wantError bool
	}{
		{
			testName:  "valid cert",
			endpoint:  endpoint,
			port:      certServerPort,
			wantCert:  true,
			wantError: false,
		},
		{
			testName:  "invalid endpoint",
			endpoint:  invalidEndpoint,
			port:      serverPort,
			wantCert:  false,
			wantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			tv := crypto.NewTlsValidator()
			hasCert, err := tv.HasSelfSignedCert(tc.endpoint, tc.port)
			if (err != nil) != tc.wantError {
				t.Fatalf("HasSelfSignedCert() error = %v, wantError %v", err, tc.wantError)
			}
			if hasCert != tc.wantCert {
				t.Fatalf("HasSelfSignedCert() returned %v, want %v", hasCert, tc.wantCert)
			}
		})
	}
}

func TestValidateCert(t *testing.T) {
	svr, err := runTestServerWithCert(serverCert, serverKey)
	if err != nil {
		t.Fatalf("starting test server with certs: %v", err)
	}
	defer svr.Close()
	serverPort := strings.Split(svr.URL, ":")[2]
	tests := []struct {
		testName  string
		endpoint  string
		port      string
		caCert    string
		wantError bool
	}{
		{
			testName:  "valid cert",
			endpoint:  endpoint,
			port:      serverPort,
			caCert:    caCert,
			wantError: false,
		},
		{
			testName:  "invalid endpoint",
			endpoint:  invalidEndpoint,
			port:      serverPort,
			caCert:    caCert,
			wantError: true,
		},
		{
			testName:  "incorrect cert",
			endpoint:  endpoint,
			port:      serverPort,
			caCert:    incorrectCert,
			wantError: true,
		},
		{
			testName:  "invalid cert format",
			endpoint:  endpoint,
			port:      serverPort,
			caCert:    invalidCert,
			wantError: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			tv := crypto.NewTlsValidator()
			err := tv.ValidateCert(tc.endpoint, tc.port, tc.caCert)
			if (err != nil) != tc.wantError {
				t.Fatalf("ValidateCert() error = %v, wantError %v", err, tc.wantError)
			}
		})
	}
}

func runTestServerWithCert(serverCert, serverKey string) (*httptest.Server, error) {
	mux := http.NewServeMux()
	svr := httptest.NewUnstartedServer(mux)
	certificate, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		return nil, fmt.Errorf("creating key pair: %v", err)
	}
	svr.TLS = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}
	svr.StartTLS()
	return svr, nil
}

func runTestServer() (*httptest.Server, error) {
	mux := http.NewServeMux()
	svr := httptest.NewServer(mux)
	return svr, nil
}
