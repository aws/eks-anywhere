package registry

import (
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
)

// GetCertificates get X509 certificates.
func GetCertificates(certFile string) (certificates *x509.CertPool, err error) {
	if len(certFile) < 1 {
		return nil, nil
	}
	fileContents, err := os.ReadFile(filepath.Clean(certFile))
	if err != nil {
		return nil, fmt.Errorf("error reading certificate file <%s>: %v", certFile, err)
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(fileContents)

	return certPool, nil
}
