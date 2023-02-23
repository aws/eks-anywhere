package executables_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

var (
	privateKeyPath = "id_rsa"
	username       = "eksa-test"
	ip             = "1.2.3.4"
	command        = []string{"some", "random", "test", "command"}
)

func TestSSHRunCommandNoError(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	ssh := executables.NewSSH(executable)

	executable.EXPECT().Execute(ctx, "-i", privateKeyPath, "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", username, ip), "some", "random", "test", "command")

	_, err := ssh.RunCommand(ctx, privateKeyPath, username, ip, command...)
	g.Expect(err).To(Not(HaveOccurred()))
}

func TestSSHRunCommandError(t *testing.T) {
	ctx := context.Background()
	g := NewWithT(t)
	mockCtrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(mockCtrl)
	ssh := executables.NewSSH(executable)
	errMsg := "sshKey invalid"

	executable.EXPECT().Execute(ctx, "-i", privateKeyPath, "-o", "StrictHostKeyChecking=no", fmt.Sprintf("%s@%s", username, ip), "some", "random", "test", "command").Return(bytes.Buffer{}, errors.New(errMsg))

	_, err := ssh.RunCommand(ctx, privateKeyPath, username, ip, command...)
	g.Expect(err).To(MatchError(fmt.Sprintf("running SSH command: %s", errMsg)))
}
