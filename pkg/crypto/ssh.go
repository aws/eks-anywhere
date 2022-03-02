package crypto

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/filewriter"
)

// SshKeysize is the key size used when calling NewSshKeyPair().
const SshKeySize = 4096

// NewSshKeyPair creates an RSA public key pair and writes each part to privateOut and publicOut. The output
// written to privateOut and pulicKeyOut is formatted as ssh-keygen would format keys.
// The private key part is PEM encoded with the key data formatted in PKCS1, ASN.1 DER as typically done by
// the ssh-keygen GNU tool. The public key is formatted as an SSH Authorized Key suitable for storing on servers.
func NewSshKeyPair(privateOut, publicOut io.Writer) error {
	private, err := rsa.GenerateKey(cryptorand.Reader, SshKeySize)
	if err != nil {
		return fmt.Errorf("generate key: %v", err)
	}

	privateEncoded := pem.EncodeToMemory(&pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(private),
	})

	public, err := ssh.NewPublicKey(&private.PublicKey)
	if err != nil {
		return err
	}
	publicEncoded := ssh.MarshalAuthorizedKey(public)

	if _, err := privateOut.Write(privateEncoded); err != nil {
		return err
	}

	if _, err := publicOut.Write(publicEncoded); err != nil {
		return err
	}

	return nil
}

// NewSshKeyPairUsingFileWriter provides a mechanism for generating SSH key pairs and writing them to the writer
// direcftory context. It exists to create compatibility with filewriter.FileWriter and compliment older code.
// The bytes returned by NewSshKeyPairUsingFileWriter are the public key formated as specified in NewSshKeyPair().
func NewSshKeyPairUsingFileWriter(writer filewriter.FileWriter, privateKeyFilename, publicKeyFilename string) ([]byte, error) {
	var private, public bytes.Buffer

	if err := NewSshKeyPair(&private, &public); err != nil {
		return nil, fmt.Errorf("generating key pair: %v", err)
	}

	if _, err := writer.Write(privateKeyFilename, private.Bytes(), filewriter.Permission0600); err != nil {
		return nil, fmt.Errorf("writing private key: %v", err)
	}

	if _, err := writer.Write(publicKeyFilename, public.Bytes(), filewriter.Permission0600); err != nil {
		return nil, fmt.Errorf("writing public key: %v", err)
	}

	return public.Bytes(), nil
}
