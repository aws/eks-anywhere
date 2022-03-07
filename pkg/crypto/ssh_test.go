package crypto_test

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/onsi/gomega"
	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/crypto"
)

func TestNewSshKeyPair(t *testing.T) {
	g := gomega.NewWithT(t)
	var priv, pub bytes.Buffer

	err := crypto.NewSshKeyPair(&priv, &pub)

	g.Expect(err).ToNot(gomega.HaveOccurred())

	block, _ := pem.Decode(priv.Bytes())
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	g.Expect(err).ToNot(gomega.HaveOccurred())

	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pub.Bytes())

	g.Expect(err).ToNot(gomega.HaveOccurred())
	g.Expect(&privKey.PublicKey).To(gomega.BeEquivalentTo(pubKey))
}
