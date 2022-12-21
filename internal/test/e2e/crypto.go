package e2e

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"golang.org/x/crypto/ssh"
)

func generateKeyPairEcdsa() (*ecdsa.PrivateKey, error) {
	k, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generating key pair: %v", err)
	}
	return k, nil
}

func pubFromPrivateKeyEcdsa(k *ecdsa.PrivateKey) ([]byte, error) {
	pub, err := ssh.NewPublicKey(k.Public())
	if err != nil {
		return nil, err
	}
	return ssh.MarshalAuthorizedKey(pub), nil
}

func pemFromPrivateKeyEcdsa(k *ecdsa.PrivateKey) ([]byte, error) {
	pk, err := x509.MarshalPKCS8PrivateKey(k)
	if err != nil {
		return nil, fmt.Errorf("marshalling private key to x509 PKCS 8 from private key: %v", err)
	}

	p := pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: pk,
	}

	return pem.EncodeToMemory(&p), nil
}
