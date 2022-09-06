package crypto_test

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
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
		openssl req -new -x509 -sha256 -days 3650 -key ca.key -out ca.crt
		openssl req -newkey rsa:2048 -nodes -keyout server.key -out server.csr
		openssl x509 -req -extfile <(printf "subjectAltName=IP:127.0.0.1") -sha256 -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
	*/
	caCert = `
-----BEGIN CERTIFICATE-----
MIIDazCCAlOgAwIBAgIUWn739ioGaXBxeHg8FNAHHfag37IwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjA4MTgyMTM0MTRaFw0zMjA4
MTUyMTM0MTRaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDhDQ1KESLi3DtHTQllnLZ7wasKgcz6bDF5QmI6hQL2
2CRLF1GWw8xg3qTTDPy0FwEYq+8dAdRqE6Lft/ZzpNtXyFa8iPZdH5egNqxS2rrd
xXKicu5ce4MDj/hmpDsfEKJKKOVl8u0vUUccmcsGaS6bqVrXJvenNbeYOXOKjuIG
Z8jDRx906G//uMsUn+ISfB91aFyHRvYfmRp1aQY1i5qxr0oCMUiG6VOBY9mvYZB+
CQbJVv0Tldmtpx0phGRZycIvAGHkxMvylyepZG3NaiYABJnV5ZtpXEmcHJnXrkeU
seLa1HQt9uyO9phw7jJl6uhmXmNIjSI7E2PacnknyDpnAgMBAAGjUzBRMB0GA1Ud
DgQWBBRcM9kTVIvbh3LriH71BhVQwU5EZDAfBgNVHSMEGDAWgBRcM9kTVIvbh3Lr
iH71BhVQwU5EZDAPBgNVHRMBAf8EBTADAQH/MA0GCSqGSIb3DQEBCwUAA4IBAQBa
pRUG9O9oPd2VorOEOiQa/kUHXT8i3zyljOv9TUGaWB7KvyJPy8V0UMCXQdLfuNp9
Jj8nhQpVN0LpDxq0jMkhWC21RdcCQ8uBpIZb87qvUNxBcHRnWHvVFFrb1l6d2TzT
EAdbMOOj8rWHhuq0rGZkhYz7hUUK873YZP9FMuhppiGcapDmfUJpR4956AYtkv8f
rvMLWhytaYxZJQrN2r8uNsklhQytJc9ZjfgGOmHkSvxUPkG6e4bts2leFVBK/g8m
NlyAQFLn7C06paTuNQkjtXypFT1ndHy4+hYewW+Yz9KvpmdmIZ4UqjEspX8vA3Lr
JvkUkvQfzDkQWnyL7D6D
-----END CERTIFICATE-----
`
	serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDOAsPAhhGBid9s
fabdjKPkqt8gl9Kp52CXTkgVGR1t0FQy1oYo5vTAOb2gliUTIs3uoxr3T7xf3O6v
Sxu/TjUV+/6G1NNjid0Nohzt37riauwPOIUNQ8eZ8DThhddS23kopQg3iGSO7GVI
jLhD9KqAA8BYa6+AFodnJfS4xC5MUFTSYQYWReI5UPPvMAwb/Qi2MDh42I3Y5m1Y
Nq+07+uCrqt1gFdR6sapB86LTF0/KRO5BBA8LKU1LKDnibS9ydBCThD9rJyYrXTp
IIa8gHZvDpgidHe3Eq3IdKeIgrULIgqP5Gt0u4uTage1AnAKzq2C/wwheSZJ+oKX
aiwnQgljAgMBAAECggEAcfsNm2OSI/3IZBpRTJsXwtjXgwz9PYwK4SwqraGrqqpt
K4ONzuuZ1xEbXdI0yRWkorgTn2m6a7+tP8dqnroCnRtDhA4Utyn76CMdpm7203cd
DH7U/rXVpdJFL9IBhJJxwo8ssK0CFiGtGCrjeJXSD+oDbeiGvXO5jtRtRh0lEIsS
oAbRC2ZM41WlCfUIXHvrPmTk0OuLAYO+q0GQnyWtaAbhYyRtk8XuuM+RuD5hOoQu
yRJJK/F0y2BhJ4jCGVUIPnGuepW5er+Wb5nCK+U05EYRvCbZGHo1rxk1yuev65KT
88k+tbUFvrfuERpNdi8GrjVzFu2XCmjpi9kFomtYoQKBgQD5sC/8kwGIYIZXSrxX
P4IKNsWwAlYwdHaE8AD9F6vw1rJ1DATmMp4pKFnJc2NAVCEfBg5LiVWuDGlJ4Ymg
Sa74rkhRBqMK2d9Oma7n+j5ILfJRANCdvRbdD6d83sal4u/RIhQZAx82YBrcASaE
6iv6S0Ya6SbtdF866Jnc7qyrbQKBgQDTN+0yKdWBSYmvMv7NgSYTsKNqO5jYQlwU
688LXn+7bekcq7F9NTBp406/MxPGIPXfbNTS+Ph0FtmspjHw6YjwzPhNBjq1EiKq
QW9npmO5jeAch3FgfZ4R5EV2/wnl9ZmPQ2qVCtEz71nP1DhU2z7HdRaM4P8uW3BF
Isd86wc2DwKBgCJuDh/o8YQpszykPJZXVoosBVSA7fueg51PLwO3WOlL4a3MK3zG
rBKG0uK5e40qTKrnfd8in+LxKS+b3wtwPaVi+uvZW3AqnOVMwdaRJjdzxn8u+pVV
tqpi9zh7y66iPWl8JoNQb+RimjGOIw6e79OCv7cEQW7q5hrMajMR4lN9AoGBAMgU
hVNsf3xOLefRlb8T5P7n55TdSacqDVIgImvxo2vn7Nek6Kfjt63GjjTebI/VbzOr
Q1tqTuihMKe0c0Bz6K26bEeCbCBUQpQnEiIMYxFFjRNZVhQCSrdGFmtnone8lC86
vH7c1VmuFNSjgo0Xdru4dZkUFYZTReGn1XLGrHkPAoGAHpmv09bLs51SpSSo1w2u
az3O3LMNWXsdeDYRu3KYUMSEkZiowiHSdRSp194OBm4NQ+8qBqpb3e07VxMeS8Ue
oWvhek8oBOpk7yCoi4zB/wi0ceKgjq0t2jJ/KMuiCqgm2EkUbT6MVp1lV0uoT13J
VQ9QJpWnSd6LTuoWuzCbY80=
-----END PRIVATE KEY-----
`
	serverCert = `
-----BEGIN CERTIFICATE-----
MIIDKzCCAhOgAwIBAgIUQRi1UwgnimZ++kgEcfbQE5GZ8d8wDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCQVUxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAeFw0yMjA4MTgyMTM1MDdaFw0yMzA4
MTgyMTM1MDdaMEUxCzAJBgNVBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEw
HwYDVQQKDBhJbnRlcm5ldCBXaWRnaXRzIFB0eSBMdGQwggEiMA0GCSqGSIb3DQEB
AQUAA4IBDwAwggEKAoIBAQDOAsPAhhGBid9sfabdjKPkqt8gl9Kp52CXTkgVGR1t
0FQy1oYo5vTAOb2gliUTIs3uoxr3T7xf3O6vSxu/TjUV+/6G1NNjid0Nohzt37ri
auwPOIUNQ8eZ8DThhddS23kopQg3iGSO7GVIjLhD9KqAA8BYa6+AFodnJfS4xC5M
UFTSYQYWReI5UPPvMAwb/Qi2MDh42I3Y5m1YNq+07+uCrqt1gFdR6sapB86LTF0/
KRO5BBA8LKU1LKDnibS9ydBCThD9rJyYrXTpIIa8gHZvDpgidHe3Eq3IdKeIgrUL
IgqP5Gt0u4uTage1AnAKzq2C/wwheSZJ+oKXaiwnQgljAgMBAAGjEzARMA8GA1Ud
EQQIMAaHBH8AAAEwDQYJKoZIhvcNAQELBQADggEBADiqHRze1eQvXbdItJOLppOl
b2YDpXTeoXtyjPVDii1ut29uGoWzuoHjb8XzY1wbKThPz6Pw3iIL26v8i7y8KLzH
LW64pz8CYxchELCuqv+a1K07an82uPnGynrEXz5rP9yOdN3+g1GDGEVdw0ziBDPc
++pGmKe0Wi6V4FOexNSJHOHkIEnxk6rhYi/450grNIkDki3f4saJcT9mB+nMgGl7
F8Wd/nMAlxKt39q4PTaNz+KohZByCJZ9BRx412B6H1hqtrUXv6sdJJrAE8IrPUmM
obFNEP4CAqPBBGeml9PF+9V9sW1HXHd095LerFJFZ0B6bNnwRLA6E9cSzo5RgIY=
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

func TestIsSignedByUnknownAuthority(t *testing.T) {
	// We need to add this check as newer version of macos require the adding trusted certs manually to test for this,
	// through `security add-trusted-cert`, so we are only able to run this test on linux for now.
	// This is a known issue here: https://github.com/golang/go/issues/52010
	// Refer to https://go-review.googlesource.com/c/go/+/353132 and https://github.com/helm/helm/pull/11160
	// Follow up tracking issue for us to fix: https://github.com/aws/eks-anywhere/issues/3267
	if runtime.GOOS == "darwin" {
		t.Skipf("Skipping as this test will fail on darwin because newer versions require this cert " +
			"to be added to the system trust store")
	}
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
			hasCert, err := tv.IsSignedByUnknownAuthority(tc.endpoint, tc.port)
			if (err != nil) != tc.wantError {
				t.Fatalf("IsSignedByUnknownAuthority() error = %v, wantError %v", err, tc.wantError)
			}
			if hasCert != tc.wantCert {
				t.Fatalf("IsSignedByUnknownAuthority() returned %v, want %v", hasCert, tc.wantCert)
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
