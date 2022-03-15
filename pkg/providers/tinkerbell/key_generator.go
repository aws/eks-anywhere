package tinkerbell

import (
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/providers/common"
)

// sshAuthKeyGenerator satisfies SSHAuthKeyGenerator. It exists to wrap the common key generation function so we can
// isolate the RNG in testing.
type sshAuthKeyGenerator struct{}

func (sshAuthKeyGenerator) GenerateSSHAuthKey(w filewriter.FileWriter) (string, error) {
	return common.GenerateSSHAuthKey(w)
}
