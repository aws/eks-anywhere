package validations

type TlsValidator interface {
	ValidateCert(host, port, caCertContent string) error
	IsSignedByUnknownAuthority(host, port string) (bool, error)
}
