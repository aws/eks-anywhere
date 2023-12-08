package cmd

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os/exec"

	"github.com/spf13/cobra"
)

type expNetworkTestOptions struct {
	hostIp         string
	username       string
	sshKey         string
	enablePackages bool
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

	err := expNetworkTestCmd.MarkFlagRequired("host")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}

	err = expNetworkTestCmd.MarkFlagRequired("ssh-key")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}
	err = expNetworkTestCmd.MarkFlagRequired("username")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}
}

var endpoints = []string{
	"public.ecr.aws",
	"anywhere-assets.eks.amazonaws.com",
	"distro.eks.amazonaws.com",
	"d2glxqk2uabbnd.cloudfront.net",
	// "api.ecr.us-west-2.amazonaws.com", only required for packages
	"d5l0dvt14r5h8.cloudfront.net",
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

func PingPublicECR(ctx context.Context, tr NetworkTestRunner) error {
	return nil
}

/*
	network access:
    vCenter endpoint (must be accessible to EKS Anywhere clusters)
    public.ecr.aws
    anywhere-assets.eks.amazonaws.com (to download the EKS Anywhere binaries, manifests and OVAs)
    distro.eks.amazonaws.com (to download EKS Distro binaries and manifests)
    d2glxqk2uabbnd.cloudfront.net (for EKS Anywhere and EKS Distro ECR container images)
    api.ecr.us-west-2.amazonaws.com (for EKS Anywhere package authentication matching your region)
    d5l0dvt14r5h8.cloudfront.net (for EKS Anywhere package ECR container images)
    api.github.com (only if GitOps is enabled)

*/

// ssh panda@195.17.172.244 echo hello world

func runSSHCmd(ctx context.Context, host string, args string) error {
	fmt.Printf("running %s\n", args)
	cmd := exec.Command("ssh", host, args)
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

func (nt expNetworkTestOptions) Call(ctx context.Context) error {
	// host := "panda@195.17.172.244"
	// host := "ubuntu@195.17.172.208"
	username := nt.username
	hostIp := nt.hostIp
	host := fmt.Sprintf("%s@%s", username, hostIp)
	fmt.Println(host)
	tr := SSHTestRunner{Host: host}

	tr.Run(ctx, "echo hello world")

	for _, endpoint := range endpoints {
		fmt.Printf("nslookup on %s\n", endpoint)
		tr.Run(ctx, fmt.Sprintf("nslookup %s", endpoint))
	}

	for _, endpoint := range endpoints {
		fmt.Printf("ping on %s\n", endpoint)
		tr.Run(ctx, fmt.Sprintf("ping -c 1 %s", endpoint))
	}

	tr.Run(ctx, "sudo crictl rmi public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53")
	tr.Run(ctx, "sudo crictl pull public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53")
	tr.Run(ctx, "sudo crictl rmi public.ecr.aws/eks-anywhere/cli-tools:v0.18.2-eks-a-53")

	return nil
}
