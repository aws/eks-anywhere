package validations

type TlsValidator interface {
	ValidateCert(host, port, cert string) error
	HasSelfSignedCert(host, port string) (bool, error)
}
