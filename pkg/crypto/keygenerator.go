package crypto

type KeyGenerator interface {
	GenerateSSHKeyPair(privateKeyDir string, publicKeyDir string, privateKeyFileName string, publicKeyFileName string, clusterUsername string) (key []byte, err error)
}
