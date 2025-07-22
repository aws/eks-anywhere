package certificates

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/aws/eks-anywhere/pkg/logger"
)

// sshClient interface and sshDialer type remain the same.
type sshClient interface {
	Close() error
	NewSession() (*ssh.Session, error)
}

// sshDialer is a function type for dialing SSH connections.
type sshDialer func(network, addr string, config *ssh.ClientConfig) (sshClient, error)

// SSHRunner provides methods for running commands over SSH.
type SSHRunner interface {
	// RunCommand runs a command on the remote host and returns the output
	RunCommand(ctx context.Context, node string, cmd string, opts ...SSHOption) (string, error)
}

// DefaultSSHRunner is the default implementation of SSHRunner.
type DefaultSSHRunner struct {
	sshConfig  *ssh.ClientConfig
	sshDialer  sshDialer
	sshKeyPath string
	sshPasswd  string
}

type SSHOption func(*sshConfigOption)

type sshConfigOption struct {
	displayLogs bool
}

func defaultSSHConfig() *sshConfigOption {
	return &sshConfigOption{
		displayLogs: true,
	}
}

func WithSSHLogging(display bool) SSHOption {
	return func(c *sshConfigOption) {
		c.displayLogs = display
	}
}

// NewSSHRunner creates a new SSH runner with the given configuration.
func NewSSHRunner(cfg SSHConfig) (*DefaultSSHRunner, error) {
	r := &DefaultSSHRunner{
		sshDialer: func(network, addr string, config *ssh.ClientConfig) (sshClient, error) {
			return ssh.Dial(network, addr, config)
		},
	}

	r.sshKeyPath = cfg.KeyPath
	r.sshPasswd = cfg.Password

	key, err := os.ReadFile(cfg.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("reading SSH key: %v", err)
	}

	signer, err := r.parsePrivateKey(key)
	if err != nil {
		return nil, err
	}

	r.sshConfig = &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	return r, nil
}

// parsePrivateKey only get password from enviroment variables.
func (r *DefaultSSHRunner) parsePrivateKey(key []byte) (ssh.Signer, error) {
	signer, err := ssh.ParsePrivateKey(key)
	if err == nil {
		return signer, nil
	}

	if r.sshPasswd == "" {
		return nil, fmt.Errorf("SSH key is passphrase protected but no passphrase provided via environment variable")
	}

	signer, err = ssh.ParsePrivateKeyWithPassphrase(key, []byte(r.sshPasswd))
	if err != nil {
		return nil, fmt.Errorf("parsing SSH key with passphrase: %v", err)
	}
	return signer, nil
}

// RunCommand executes a command on the remote node via SSH and returns the output.
func (r *DefaultSSHRunner) RunCommand(ctx context.Context, node string, cmd string, opts ...SSHOption) (string, error) {
	cfg := defaultSSHConfig()
	for _, opt := range opts {
		opt(cfg)
	}
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return "", fmt.Errorf("connect to node %s: %v", node, err)
	}
	defer client.Close()

	cmdStr := cmd

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

		outputBytes, err := session.CombinedOutput(cmdStr)
		output := strings.TrimSpace(string(outputBytes))

		if cfg.displayLogs {
			logger.V(6).Info(cmdStr)
			logger.V(6).Info(output)
		}

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
