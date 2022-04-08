package common

import (
	"github.com/aws/eks-anywhere/pkg/filewriter"
)

// sshAuthKeyGenerator satisfies SSHAuthKeyGenerator. It exists to wrap the common key generation function so we can
// isolate the RNG in testing.
type SshAuthKeyGenerator struct{}

func (SshAuthKeyGenerator) GenerateSSHAuthKey(w filewriter.FileWriter) (string, error) {
	return GenerateSSHAuthKey(w)
}
