package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"
)

type certificategenerator struct{}

type CertificateGenerator interface {
	GenerateSelfSignCertKeyPair() ([]byte, []byte, error)
}

func NewCertificateGenerator() CertificateGenerator {
	return &certificategenerator{}
}

func (cg *certificategenerator) GenerateSelfSignCertKeyPair() ([]byte, []byte, error) {
	bitSize := 2048

	privateKey, err := cg.generatePrivateKey(bitSize)
	if err != nil || privateKey == nil {
		return nil, nil, fmt.Errorf("failed to generate private key for self sign cert: %v", err)
	}

	notBefore, notAfter := cg.getCertLifeTime()

	serialNumber, err := cg.generateCertSerialNumber()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate serial number for self sign cert: %v", err)
	}
	certBytes, err := cg.generateCertificate(serialNumber, privateKey, notBefore, notAfter)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate certificate for self sign cert: %v", err)
	}
	keyBytes := cg.encodePrivateKey(privateKey)

	return cg.encodeToPEM(certBytes, "CERTIFICATE"), cg.encodeToPEM(keyBytes, "RSA PRIVATE KEY"), nil
}

func (cg *certificategenerator) generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
	// Private Key generation
	privateKey, err := rsa.GenerateKey(rand.Reader, bitSize)
	if err != nil {
		return nil, err
	}

	// Validate Private Key
	err = privateKey.Validate()
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func (cg *certificategenerator) getCertLifeTime() (time.Time, time.Time) {
	// lifetime of the CA certificate
	lifeTime := time.Hour * 24 * 365 * 100
	lifeTimeStart := time.Now()
	lifeTimeEnd := lifeTimeStart.Add(lifeTime)

	return lifeTimeStart, lifeTimeEnd
}

func (cg *certificategenerator) generateCertSerialNumber() (*big.Int, error) {
	// choose a random 128 bit serial number
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return nil, err
	}
	return serialNumber, nil
}

func (cg *certificategenerator) generateCertificate(serialNumber *big.Int, privateKey *rsa.PrivateKey, notBefore, notAfter time.Time) ([]byte, error) {
	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			CommonName: "aws-iam-authenticator",
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
		DNSNames:              []string{"localhost"},
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
	if err != nil {
		return nil, err
	}

	return certBytes, nil
}

func (cg *certificategenerator) encodePrivateKey(privateKey *rsa.PrivateKey) []byte {
	// ASN.1 DER format
	return x509.MarshalPKCS1PrivateKey(privateKey)
}

func (cg *certificategenerator) encodeToPEM(bytes []byte, blockType string) []byte {
	block := pem.Block{
		Type:    blockType,
		Headers: nil,
		Bytes:   bytes,
	}
	return pem.EncodeToMemory(&block)
}
