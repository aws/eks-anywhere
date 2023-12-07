package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

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

func main() {

	ctx := context.Background()
	// host := "panda@195.17.172.244"
	host := "ubuntu@195.17.172.208"
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
}
