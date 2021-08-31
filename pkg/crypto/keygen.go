package crypto

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type keygenerator struct {
	writer filewriter.FileWriter
}

func NewKeyGenerator(writer filewriter.FileWriter) (KeyGenerator, error) {
	return &keygenerator{
		writer: writer,
	}, nil
}

func (kg *keygenerator) GenerateSSHKeyPair(privateKeyDir string, publicKeyDir string, privateKeyFileName string, publicKeyFileName string, clusterUsername string) (key []byte, err error) {
	bitSize := 4096

	privateKey, err := kg.generatePrivateKey(bitSize)
	if err != nil || privateKey == nil {
		return nil, fmt.Errorf("failed to generate private key: %s", err.Error())
	}

	publicKeyBytes, err := kg.generatePublicKey(&privateKey.PublicKey)
	if err != nil || publicKeyBytes == nil {
		return nil, fmt.Errorf("failed to generate public key: %s", err.Error())
	}

	privateKeyBytes := kg.encodePrivateKeyToPEM(privateKey)
	if privateKeyBytes == nil {
		return nil, errors.New("failed to encode private key")
	}
	filePath, err := kg.writeKeyToFile(privateKeyBytes, privateKeyDir, privateKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to write private key to %s: %s", filePath, err.Error())
	}
	logMsg := fmt.Sprintf("VSphereDatacenterConfig private key saved to %s. Use 'ssh -i %s %s@<VM-IP-Address>' to login to your cluster VM", filePath, filePath, clusterUsername)
	logger.Info(logMsg)

	filePath, err = kg.writeKeyToFile([]byte(publicKeyBytes), publicKeyDir, publicKeyFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to write public key to %s: %s", filePath, err.Error())
	}

	return publicKeyBytes, nil
}

// generatePrivateKey creates a RSA Private Key of specified byte size
func (kg *keygenerator) generatePrivateKey(bitSize int) (*rsa.PrivateKey, error) {
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

// generatePublicKey take a rsa.PublicKey and return bytes suitable for writing to .pub file
// returns in the format "ssh-rsa ..."
func (kg *keygenerator) generatePublicKey(privatekey *rsa.PublicKey) ([]byte, error) {
	publicRsaKey, err := ssh.NewPublicKey(privatekey)
	if err != nil {
		return nil, err
	}

	pubKeyBytes := ssh.MarshalAuthorizedKey(publicRsaKey)

	return pubKeyBytes, nil
}

// writePemToFile writes keys to a file
func (kg *keygenerator) writeKeyToFile(keyBytes []byte, dir string, saveFileTo string) (string, error) {
	keyFileWriter, err := kg.writer.WithDir(dir)
	if err != nil {
		return "", err
	}

	return keyFileWriter.Write(saveFileTo, keyBytes, filewriter.PersistentFile, filewriter.Permission0600)
}

// encodePrivateKeyToPEM encodes Private Key from RSA to PEM format
func (kg *keygenerator) encodePrivateKeyToPEM(privateKey *rsa.PrivateKey) []byte {
	// Get ASN.1 DER format
	privDER := x509.MarshalPKCS1PrivateKey(privateKey)

	// pem.Block
	privBlock := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   privDER,
	}

	// Private key in PEM format
	privatePEM := pem.EncodeToMemory(&privBlock)

	return privatePEM
}
