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
		openssl req -new -x509 -sha256 -days 36500 -key ca.key -out ca.crt
		openssl req -newkey rsa:2048 -nodes -keyout server.key -out server.csr
		openssl x509 -req -extfile <(printf "subjectAltName=IP:127.0.0.1") -sha256 -days 36500 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
	*/
	caCert = `
-----BEGIN CERTIFICATE-----
MIIDbTCCAlWgAwIBAgIULI93MO3Gkw42mhimlmai7VqqwlAwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yMzA4MjExNzE3MjdaGA8yMTIz
MDcyODE3MTcyN1owRTELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAOZnOwNSWT64j/xrBKsDZ1QcfeW093lA2oN9LkqD
FXUZmb4KQWW/0XWHgqopfz0ndHZbbzTsB9+OwCa6T7mTMxzpa5GyVy47wdHhoyXl
l2xDnCCGYEOOwft72H+BHgPMeDT3XXYn/WGTRwVM4HyLg6hZWTIsIGOriRC16saf
1bpL2jXX07cBooPA7IH3xNpRNzBbqsgKPLPXo/sWIsYjYsCcVrqTMX+p/r8h8lcb
vKin7dhFK2xPyCLfujh2ge7w5nyn7+PCzA9hLYokI/2t6w0d6FWbUzx3v3Ztpt7m
+VqcOqgZcKOazQLfpwL5k4TAaggVDUUY9RXMCfqX/7FkI80CAwEAAaNTMFEwHQYD
VR0OBBYEFDi6QiXh0xHE5EMUO8LQopPX3cuzMB8GA1UdIwQYMBaAFDi6QiXh0xHE
5EMUO8LQopPX3cuzMA8GA1UdEwEB/wQFMAMBAf8wDQYJKoZIhvcNAQELBQADggEB
ADgvICvqDab0MeQh9HSkWJCoKf899W3kFR2lBxZEnjWJKqPj8oy0MkHGZA0vW+1Z
j8D8RsZFaodLCTlYr/FxO2vVB5iQoxaVaVXIqSdi1XhLGxMEmD21fnmfs3ETQw+C
jHr83KQNYKAstmaO2r3uNJ3fphbzBGa2G/MB4zHSkA8QlAO8bmUFlr6jXeWhVH24
nnRnEzGMVNcgl9ymJOv3dbk8Gg60LXtgypa2iOOuA4T26a90CH7Rn4iq+eYOpZ36
IU7igvQUMse0GscZXoTgDSs1rZMWdK6W8WIZOHa1/n9YV+UF7kmx3zq7+thT879d
JLgMkqC6LJL/7gta8J7P3/w=
-----END CERTIFICATE-----
`
	serverKey = `
-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQCh9Js646xioQVM
K/vG/mZwJ2tHIH56MdggIhKqdXQ/+/q+gBG9p4F4L51SDnroZEOQCk+mEIjaChTg
uilxnQkGBPIZVCkhs1mSjJaKcSTUS7IPbMOAUkGoPx/JZ1MFVAYyQyEJ5T1bMC8M
u2h1a73h5Q/WhcDieq79AF04Sj7vd9qPqCjlGAlZZMISbnjBzT0g54X+c+3EJPcU
vRsqiJ6114wS+tNCLDi+7V2MpdWtYiOMf25WVvQAckoJWBK2vVOdtcHLa34OEjpg
xrbSE5+NaIQ4t+SHM7U/fyZwAvG+yVfAsAabZYtwdiqeUUVsN/WQn0mBVU+/ac+g
Vkpy6vKPAgMBAAECggEAbaCv+vrmU5T/iwIy2g6WtNBaE7lxI7HdxFKMJAqBeFZO
8uiqGaDrSLmiWksm82v7N+Ni6txCZqVwYHZjW16bHxH19yM6/G708PhtZqUT8wiS
LVLIO3JyszqXf7OLRQ4Na7R5BsO/9pv1HT82hFU8IU72m5XKbQPys5utfk7zrxB0
LqklQh9l1tBnHQB2LerkDwFvHBQnWyz0G3oH7XhcgTjf2V0jxQP68sBB0lNog+/n
WCK/uRrq3MQNJ5/aTo4EA1AbFpnKuUh40G4R3WUd0FHjNG3AGCRetG9aiT3wanXu
GMchwlWIlHvM76KSCgwqBoKfRe9tTSsJ/jKZBCN8EQKBgQDXYciAqk7qiMWNhLTO
6GZMMCo2TobVzobeQRPORcGJqyEd53aso5bRYyQ+1ZTSWn//dsIG5U/RvriLFGhi
kUKZsW/ysw07iXntMJ6cgCfaDZVnPZ1lU8q6u8iCm/n8YxpeO3xvrBlqsZHS/Gcv
qA7zrtMIjMdlsVpXki2SlL7cEwKBgQDAf3+WQJlFGLg9MKVvpMpKLd2gZitUAHEO
FErcRoUrWpI69Zm/9S4kmUtcigBvRjtHh0K4PmdO/jnkAa8IkwdSB/LEJlzj6xlM
ngyBbx12OYEsBHstHnK82wwEh8FMv1R8oGutgFSvtYGwj/KAJsHcgc3h37rHtz51
eXTE5NQnFQKBgDzU41pjv7JCOvnDd4XQ4cO2/tUjzLVqhXEUlFF4JjeJ2+qkS0Jr
Wdd91ujgTg4zfV9b3WUxMyAca+bsNqdQO61JDkNQCva713IEf2fYUmkl2QK1xlSu
G94t1238O9jq7LGcv4KS6wLVcEhU/ZfaMY7l27jYeiDgvJFrOe/ijx19AoGBAKvb
qH7uYWgs75/BpJGOIAQl+q3PSXOPiV+2gjcyxoW0MMVuq7uTG4UVTxDpLAYsP5RW
kByJqhX+JpUHY8tV6L0112mDjn71T/r1R9ju6PC52jcAMTBQ9MLjFVGeGdd5Iea6
GigmYHUWqRiHC0uaTo2dXcAAzHKtiJe8vaFjYn0ZAoGBAKF2XCB2IzqKX+BSFoqM
YmOVIkn6uGkaz1sn6P2yNWVTzeziVSo2+5wD3N9kDG3m5F5XzJL98J1LKX/55i7w
AVEgC7pz5U4ElvmiTdG28FAgFGdCRQVed1iO+Kk9psLx50o7MKLNA2zaE6+LGXHr
F9FXKvwFsJ7Nx3pdYbU0vKuM
-----END PRIVATE KEY-----
`
	serverCert = `
-----BEGIN CERTIFICATE-----
MIIDLTCCAhWgAwIBAgIUbVS5oLpHLZ/nY81Qj20RxFLofvIwDQYJKoZIhvcNAQEL
BQAwRTELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUxITAfBgNVBAoM
GEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDAgFw0yMzA4MjExNzE4MDNaGA8yMTIz
MDcyODE3MTgwM1owRTELMAkGA1UEBhMCVVMxEzARBgNVBAgMClNvbWUtU3RhdGUx
ITAfBgNVBAoMGEludGVybmV0IFdpZGdpdHMgUHR5IEx0ZDCCASIwDQYJKoZIhvcN
AQEBBQADggEPADCCAQoCggEBAKH0mzrjrGKhBUwr+8b+ZnAna0cgfnox2CAiEqp1
dD/7+r6AEb2ngXgvnVIOeuhkQ5AKT6YQiNoKFOC6KXGdCQYE8hlUKSGzWZKMlopx
JNRLsg9sw4BSQag/H8lnUwVUBjJDIQnlPVswLwy7aHVrveHlD9aFwOJ6rv0AXThK
Pu932o+oKOUYCVlkwhJueMHNPSDnhf5z7cQk9xS9GyqInrXXjBL600IsOL7tXYyl
1a1iI4x/blZW9ABySglYEra9U521wctrfg4SOmDGttITn41ohDi35IcztT9/JnAC
8b7JV8CwBptli3B2Kp5RRWw39ZCfSYFVT79pz6BWSnLq8o8CAwEAAaMTMBEwDwYD
VR0RBAgwBocEfwAAATANBgkqhkiG9w0BAQsFAAOCAQEAXGIfc53xz+KD2/qJaJsc
d/h9ZYTuJPWCuVuModgLeq2X56OYutwD6gTM6G/YgMnIAZQehpq69gDSBwXn4Iok
jeNDi1RFVaXXkS99cC0VehAuNSMU56iPwHN1lquBMkKe5WiAMSPm1AonoQWhC5ns
r6hi55qmpOSBp9WVsbOzo0jYPlAuRvYjUKbR8MpOnZ8ekSM3SFDWh3xp5mVERPX6
S1w0Y4z78MhEcPV1rGJQjgV4J2rHKLCaPbnuK5auPhRj8usU3S3PDVfFBbmDHKRR
mC+vJiNFtZJX5c9bIs7tx5HjV6ExHvMbjVW2LV8KcsECVa4ciHf+jvRs+CmUQNit
LQ==
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
