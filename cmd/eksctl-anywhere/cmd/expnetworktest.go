package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

const (
	testImage = "public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53"
)

type expNetworkTestOptions struct {
	hostIp         string
	username       string
	sshKey         string
	enablePackages bool
	packagesRegion string
	additionalEndpoints []string
	enableGitOps bool
}

var nt = &expNetworkTestOptions{}

var expNetworkTestCmd = &cobra.Command{
	Use:          "network test -h <host-ip> [flags]",
	Short:        "Experimental Network Test command",
	Long:         "This command is used to test the correct network access",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		return nt.Call(ctx)
	},
}

func init() {
	expCmd.AddCommand(expNetworkTestCmd)
	expNetworkTestCmd.Flags().StringVar(&nt.hostIp, "host", "", "IP of VM")
	expNetworkTestCmd.Flags().StringVar(&nt.sshKey, "ssh-key", "", "ssh-key of admin user")
	expNetworkTestCmd.Flags().StringVar(&nt.username, "username", "", "username of admin user")
	expNetworkTestCmd.Flags().BoolVar(&nt.enablePackages, "enable-packages", false, "enable packages network check")
	expNetworkTestCmd.Flags().StringVar(&nt.packagesRegion, "packages-region", "", "packages region")
	expNetworkTestCmd.Flags().StringArrayVar(&nt.additionalEndpoints, "additional-endpoints", []string{}, "additional endpoints")
	expNetworkTestCmd.Flags().BoolVar(&nt.enablePackages, "enable-git-ops", false, "enable gitops")

	err := expNetworkTestCmd.MarkFlagRequired("host")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}
}

var endpoints = []string{
	"public.ecr.aws",
	"anywhere-assets.eks.amazonaws.com",
	"distro.eks.amazonaws.com",
	"d2glxqk2uabbnd.cloudfront.net",
	"d5l0dvt14r5h8.cloudfront.net",
	"api.github.com",
}

var gitOpsEndpoints = []string{
	"api.github.com",
}

type NetworkTestRunner interface {
	Run(ctx context.Context, args []string) error
}

type SSHTestRunner struct {
	Host string
}

func (str *SSHTestRunner) Run(ctx context.Context, args string) error {
	fmt.Printf("running %s\n", args)
	cmd := exec.Command("ssh", str.Host, args)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	fmt.Println(stdout.String())
	fmt.Println(stderr.String())
	if err != nil {
		return err
	}

	return nil
}

// ssh panda@195.17.172.244 echo hello world
// host := "ubuntu@195.17.172.208"

func (nt expNetworkTestOptions) Call(ctx context.Context) error {

	username := nt.username
	hostIp := nt.hostIp
	host := fmt.Sprintf("%s@%s", username, hostIp)
	fmt.Println(host)
	tr := SSHTestRunner{Host: host}

	if nt.enablePackages && nt.packagesRegion == "" {
		return fmt.Errorf("please include packages-region if packages are enabled, e.g. us-west-2")
	}

	if nt.enablePackages {
		endpoints = append(endpoints, fmt.Sprintf("api.ecr.%s.amazonaws.com", nt.packagesRegion))
	}

	if nt.enableGitOps {
		endpoints = append(endpoints, gitOpsEndpoints...)
	}

	endpoints = append(endpoints, nt.additionalEndpoints...)

	tr.Run(ctx, "echo hello world")

	for _, endpoint := range endpoints {
		fmt.Printf("nslookup on %s\n", endpoint)
		tr.Run(ctx, fmt.Sprintf("nslookup %s", endpoint))
	}

	for _, endpoint := range endpoints {
		fmt.Printf("ping on %s\n", endpoint)
		tr.Run(ctx, fmt.Sprintf("ping -c 1 %s", endpoint))
	}

	tr.Run(ctx, fmt.Sprintf("sudo crictl rmi %s", testImage))
	tr.Run(ctx, fmt.Sprintf("sudo crictl pull %s", testImage))
	tr.Run(ctx, fmt.Sprintf("sudo crictl rmi %s", testImage))

	return nil
}
