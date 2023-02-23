package executables

import (
	"context"
	"fmt"
)

// SSH is an executable for running SSH commands.
type SSH struct {
	Executable
}

const (
	sshPath             = "ssh"
	strictHostCheckFlag = "StrictHostKeyChecking=no"
)

// NewSSH returns a new instance of SSH client.
func NewSSH(executable Executable) *SSH {
	return &SSH{
		Executable: executable,
	}
}

// RunCommand runs a command on the host using SSH.
func (s *SSH) RunCommand(ctx context.Context, privateKeyPath, username, IP string, command ...string) (string, error) {
	params := []string{
		"-i", privateKeyPath,
		"-o", strictHostCheckFlag,
		fmt.Sprintf("%s@%s", username, IP),
	}
	params = append(params, command...)

	out, err := s.Executable.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("running SSH command: %v", err)
	}

	return out.String(), nil
}
