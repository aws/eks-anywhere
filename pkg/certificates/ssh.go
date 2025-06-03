package certificates

import (
	"golang.org/x/crypto/ssh"
)

// sshClient interface defines the methods we need from ssh.Client
type sshClient interface {
	Close() error
	NewSession() (*ssh.Session, error)
}
