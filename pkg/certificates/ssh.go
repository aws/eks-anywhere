package certificates

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/term"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// sshClient interface defines the methods we need from ssh.Client.
type sshClient interface {
	Close() error
	NewSession() (*ssh.Session, error)
}

// sshDialer is a function type for dialing SSH connections.
type sshDialer func(network, addr string, config *ssh.ClientConfig) (sshClient, error)

// SSHRunner provides methods for running commands over SSH.
type SSHRunner interface {
	// RunCommand runs a command on the remote host
	RunCommand(ctx context.Context, node string, cmd string) error

	// RunCommandWithOutput runs a command on the remote host and returns the output
	RunCommandWithOutput(ctx context.Context, node string, cmd string) (string, error)

	// InitSSHConfig initializes the SSH configuration
	InitSSHConfig(user, keyPath, passwd string) error

	DownloadFile(ctx context.Context, node, remote, local string) error
}

// DefaultSSHRunner is the default implementation of SSHRunner.
type DefaultSSHRunner struct {
	sshConfig  *ssh.ClientConfig
	sshDialer  sshDialer
	sshKeyPath string
	sshPasswd  string
}

// NewSSHRunner creates a new DefaultSSHRunner.
func NewSSHRunner() *DefaultSSHRunner {
	return &DefaultSSHRunner{
		sshDialer: func(network, addr string, config *ssh.ClientConfig) (sshClient, error) {
			return ssh.Dial(network, addr, config)
		},
	}
}

// InitSSHConfig initializes the SSH configuration.
func (r *DefaultSSHRunner) InitSSHConfig(user, keyPath, passwd string) error {
	if r.sshConfig != nil && r.sshKeyPath == keyPath {
		return nil
	}

	r.sshKeyPath = keyPath // Store SSH key path.
	key, err := os.ReadFile(keyPath)
	if err != nil {
		return fmt.Errorf("reading SSH key: %v", err)
	}

	signer, err := r.parsePrivateKey(key, keyPath, passwd)
	if err != nil {
		return err
	}

	r.sshConfig = &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	return nil
}

// parsePrivateKey parses an SSH private key, handling passphrase protection if needed.
func (r *DefaultSSHRunner) parsePrivateKey(key []byte, keyPath, passwd string) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err == nil {
		return signer, nil
	}

	if err.Error() == "ssh: this private key is passphrase protected" {
		if passwd == "" && r.sshPasswd == "" {
			logger.Info("Enter passphrase for SSH key", "path", keyPath)
			passphrase, err := term.ReadPassword(int(os.Stdin.Fd()))
			if err != nil {
				return nil, fmt.Errorf("reading passphrase: %v", err)
			}
			logger.Info("")
			r.sshPasswd = string(passphrase)
		} else if passwd != "" {
			r.sshPasswd = passwd
		}

		signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(r.sshPasswd))
		if err != nil {
			return nil, fmt.Errorf("parsing SSH key with passphrase: %v", err)
		}
		return signer, nil
	}

	return nil, fmt.Errorf("parsing SSH key: %v", err)
}

// RunCommand runs a command on the remote host.
func (r *DefaultSSHRunner) RunCommand(ctx context.Context, node string, cmd string) error {
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to node %s: %v", node, err)
	}
	defer client.Close()

	done := make(chan error, 1)
	go func() {
		session, err := client.NewSession()
		if err != nil {
			done <- fmt.Errorf("creating session: %v", err)
			return
		}
		defer session.Close()

		done <- r.executeCommand(session, cmd, node)
	}()

	select {
	case <-ctx.Done():
		return fmt.Errorf("cancelling command: %v", ctx.Err())
	case err := <-done:
		if err != nil {
			return fmt.Errorf("executing command: %v", err)
		}
		return nil
	}
}

// executeCommand executes a command on an SSH session and handles its output.
func (r *DefaultSSHRunner) executeCommand(session *ssh.Session, cmd string, node string) error {
	if VerbosityLevel >= 2 {
		session.Stdout = os.Stdout
		session.Stderr = os.Stderr
		return session.Run(cmd)
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err := session.Run(cmd)

	// Special handling for certificate check commands
	if strings.Contains(cmd, "kubeadm certs check-expiration") && VerbosityLevel >= 1 {
		lines := strings.Split(stdout.String(), "\n")
		logger.Info("Certificate check results", "node", node)
		for _, line := range lines {
			if line != "" {
				logger.Info(line)
			}
		}
	}

	return err
}

// RunCommandWithOutput runs a command on the remote host and returns the output.
func (r *DefaultSSHRunner) RunCommandWithOutput(ctx context.Context, node string, cmd string) (string, error) {
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return "", fmt.Errorf("failed to connect to node %s: %v", node, err)
	}
	defer client.Close()

	type result struct {
		output string
		err    error
	}
	done := make(chan result, 1)

	go func() {
		session, err := client.NewSession()
		if err != nil {
			done <- result{"", fmt.Errorf("creating session: %v", err)}
			return
		}
		defer session.Close()

		outputBytes, err := session.CombinedOutput(cmd)
		output := strings.TrimSpace(string(outputBytes))

		if err != nil {
			if output != "" {
				done <- result{output, fmt.Errorf("executing command: %v, output: %s", err, output)}
			} else {
				done <- result{"", fmt.Errorf("executing command: %v", err)}
			}
			return
		}

		done <- result{output, nil}
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("cancelling command: %v", ctx.Err())
	case res := <-done:
		return res.output, res.err
	}
}

// DownloadFile copies a remote file to the local host via an SSH cat pipe.
func (r *DefaultSSHRunner) DownloadFile(ctx context.Context, node, remote, local string) error {
	output, err := r.RunCommandWithOutput(ctx, node, fmt.Sprintf("sudo cat %s", remote))
	if err != nil {
		return err
	}
	return os.WriteFile(local, []byte(output), 0o600)
}
