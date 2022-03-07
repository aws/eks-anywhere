package providers

import (
	"bytes"
	"fmt"

	"github.com/go-logr/logr"
	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/crypto"
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

// type SshPublicKeyRetriever func() (string, error)

// func ConfigureUsersSshKeyIfNotPresent(user *v1alpha1.UserConfiguration, getSshPublicKey SshPublicKeyRetriever) error {
// 	// If we already have a key do nothing.
// 	if len(user.SshAuthorizedKeys) > 0 && len(user.SshAuthorizedKeys[0]) > 0 {
// 		return nil
// 	}

// 	// If the there aren't any key entries at all add an item to the slice so
// 	// we don't segfault when trying to set it.
// 	if len(user.SshAuthorizedKeys) == 0 {
// 		user.SshAuthorizedKeys = append(user.SshAuthorizedKeys, "")
// 	}

// 	// Retrieve and set the key.
// 	key, err := getSshPublicKey()
// 	if err != nil {
// 		return err
// 	}
// 	user.SshAuthorizedKeys[0] = key

// 	return nil
// }

// func NewLazyOneTimeGenerationSshPublicKeyRetriever(writer filewriter.FileWriter, privateKeyFilename, publicKeyFilename string) SshPublicKeyRetriever {
// 	var publicKey string
// 	return func() (string, error) {
// 		if publicKey == "" {
// 			key, err := crypto.NewSshKeyPairUsingFileWriter(writer, privateKeyFilename, publicKeyFilename)
// 			if err != nil {
// 				return "", err
// 			}
// 			publicKey = string(key)
// 		}
// 		return publicKey, nil
// 	}
// }

// func NormalizeSshAuthorizedKey(marshalledKey string) (string, error) {
// 	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(marshalledKey))
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(ssh.MarshalAuthorizedKey(key)), nil
// }

// Standard key file names.
const (
	PrivateSshKeyFileName = "eks-a-id_rsa"
	PublicSshKeyFileName  = "eks-a-id_rsa.pub"
)

func ValidateUserSshKeys(user v1alpha1.UserConfiguration) error {
	if len(user.SshAuthorizedKeys) == 0 {
		return fmt.Errorf("missing ssh authorized key for user %v", user.Name)
	}

	for _, key := range user.SshAuthorizedKeys {
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(key)); err != nil {
			return fmt.Errorf("%v has an invalid ssh key: %v", user.Name, err)
		}
	}

	return nil
}

func ConfigureUsersSshKeyIfNotPresent(writer filewriter.FileWriter, users []*v1alpha1.UserConfiguration, logger logr.Logger) error {
	usersWithoutSshKeys := filterUsersWithoutSshKey(users)
	if len(usersWithoutSshKeys) == 0 {
		return nil
	}

	privateKeyPath, publicKey, err := NewSshKeyPairUsingFileWriter(writer)
	if err != nil {
		return fmt.Errorf("shared machine config ssh key generation: %v", err)
	}

	logger.Info(
		"SSH key generated for users without keys. Use 'ssh -i %s %s@<VM-IP-Address>' to log in to your cluster machines.",
		privateKeyPath,
	)

	applyPublicSshKeyToUsers(usersWithoutSshKeys, string(publicKey))

	return nil
}

func applyPublicSshKeyToUsers(users []*v1alpha1.UserConfiguration, key string) {
	for _, user := range users {
		applyPublicSshKeyToUser(user, key)
	}
}

func applyPublicSshKeyToUser(user *v1alpha1.UserConfiguration, key string) {
	if len(user.SshAuthorizedKeys) == 0 {
		user.SshAuthorizedKeys = append(user.SshAuthorizedKeys, "")
	}
	user.SshAuthorizedKeys[0] = key
}

func filterUsersWithoutSshKey(users []*v1alpha1.UserConfiguration) []*v1alpha1.UserConfiguration {
	for i, user := range users {
		if len(user.SshAuthorizedKeys) > 0 && len(user.SshAuthorizedKeys[0]) > 0 {
			users = append(users[:i], users[i+1:]...)
		}
	}
	return users
}

// NewSshKeyPairUsingFileWriter provides a mechanism for generating SSH key pairs and writing them to the writer
// direcftory context. It exists to create compatibility with filewriter.FileWriter and compliment older code.
// The string returned is a path to the private key written to disk using writer.
// The bytes returned are the public key formated as specified in NewSshKeyPair().
func NewSshKeyPairUsingFileWriter(writer filewriter.FileWriter) (string, []byte, error) {
	var private, public bytes.Buffer

	if err := crypto.NewSshKeyPair(&private, &public); err != nil {
		return "", nil, fmt.Errorf("generating key pair: %v", err)
	}

	privateKeyPath, err := writer.Write(PrivateSshKeyFileName, private.Bytes(), filewriter.PersistentFile, filewriter.Permission0600)
	if err != nil {
		return "", nil, fmt.Errorf("writing private key: %v", err)
	}

	if _, err := writer.Write(PublicSshKeyFileName, public.Bytes(), filewriter.PersistentFile, filewriter.Permission0600); err != nil {
		return "", nil, fmt.Errorf("writing public key: %v", err)
	}

	return privateKeyPath, public.Bytes(), nil
}
