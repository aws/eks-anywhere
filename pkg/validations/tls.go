package validations

type TlsValidator interface {
	ValidateCert(host, port, caCertContent string) error
	HasSelfSignedCert(host, port string) (bool, error)
}
