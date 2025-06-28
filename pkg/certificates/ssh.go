package certificates

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
	"golang.org/x/crypto/ssh"
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
	// RunCommand runs a command on the remote host
	RunCommand(ctx context.Context, node string, cmds []string) error
	// RunCommandWithOutput runs a command on the remote host and returns the output
	RunCommandWithOutput(ctx context.Context, node string, cmds []string) (string, error)
}

// DefaultSSHRunner is the default implementation of SSHRunner.
type DefaultSSHRunner struct {
	sshConfig  *ssh.ClientConfig
	sshDialer  sshDialer
	sshKeyPath string
	sshPasswd  string
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

// RunCommand runs a command on the remote host.
func (r *DefaultSSHRunner) RunCommand(ctx context.Context, node string, cmds []string) error {
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return fmt.Errorf("connect to node %s: %v", node, err)
	}
	defer client.Close()

	cmdStr := r.buildCommandString(cmds)

	done := make(chan error, 1)
	go func() {
		session, err := client.NewSession()
		if err != nil {
			done <- fmt.Errorf("creating session: %v", err)
			return
		}
		defer session.Close()

		done <- r.executeCommand(session, cmdStr, node)
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
	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err := session.Run(cmd)

	if stdout.Len() > 0 {
		logger.V(2).Info(fmt.Sprintf("Command executed on node %s:\n%s", node, cmd))
		logger.V(2).Info(fmt.Sprintf("Command stdout for node %s:\n%s", node, stdout.String()))
	}
	if stderr.Len() > 0 {
		logger.V(2).Info(fmt.Sprintf("Command stderr for node %s:\n%s", node, stderr.String()))
	}

	return err
}

// RunCommandWithOutput runs a command on the remote host and returns the output.
func (r *DefaultSSHRunner) RunCommandWithOutput(ctx context.Context, node string, cmds []string) (string, error) {
	client, err := r.sshDialer("tcp", fmt.Sprintf("%s:22", node), r.sshConfig)
	if err != nil {
		return "", fmt.Errorf("connect to node %s: %v", node, err)
	}
	defer client.Close()

	cmdStr := r.buildCommandString(cmds)

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

func (r *DefaultSSHRunner) buildCommandString(cmds []string) string {
	if len(cmds) >= 3 && cmds[0] == "sudo" && cmds[1] == "sh" && cmds[2] == "-c" {
		if len(cmds) == 4 {
			return fmt.Sprintf("%s %s %s '%s'", cmds[0], cmds[1], cmds[2], cmds[3])
		}
	}

	if len(cmds) == 1 && strings.Contains(cmds[0], "sudo sheltie") {
		return cmds[0]
	}
	return strings.Join(cmds, " ")
}
