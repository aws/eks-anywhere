package cmd

import (
	"fmt"
	"log"

	"github.com/cheynewallace/tabby"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest"
	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"
	"github.com/spf13/cobra"
)

type networkTestOptions struct {
	hostIP              string
	username            string
	sshKey              string
	enablePackages      bool
	packagesRegion      string
	additionalEndpoints []string
	enableGitOps        bool
}

func NewNetworkTestCmd() *cobra.Command {
	var nt networkTestOptions

	cmd := &cobra.Command{
		Use:          "network test -h <host-ip> [flags]",
		Short:        "Experimental Network Test command",
		Long:         "This command is used to test the correct network access",
		SilenceUsage: true,
		RunE:         nt.RunE,
	}

	flgs := cmd.Flags()
	flgs.StringVar(&nt.hostIP, "host", "", "IP of VM")
	flgs.StringVar(&nt.sshKey, "ssh-key", "", "SSH key of admin user")
	flgs.StringVar(&nt.username, "username", "", "Username of admin user")
	flgs.BoolVar(&nt.enablePackages, "enable-packages", false, "Enable packages network check")
	flgs.StringVar(&nt.packagesRegion, "packages-region", "", "Packages region")
	flgs.StringArrayVar(&nt.additionalEndpoints, "additional-endpoints", []string{}, "Additional endpoints to run DNS resolution and connectivity tests against")
	flgs.BoolVar(&nt.enableGitOps, "enable-gitops", false, "Enable gitops")

	err := cmd.MarkFlagRequired("host")
	if err != nil {
		log.Fatalf("marking flag as required: %v", err)
	}

	return cmd
}

func (nt *networkTestOptions) RunE(cmd *cobra.Command, _ []string) error {
	username := nt.username
	hostIP := nt.hostIP
	host := fmt.Sprintf("%s@%s", username, hostIP)

	if nt.enablePackages && nt.packagesRegion == "" {
		return fmt.Errorf("please include packages-region if packages are enabled, e.g. us-west-2")
	}

	var additionalEndpoints []string

	if nt.enablePackages {
		additionalEndpoints = append(additionalEndpoints, fmt.Sprintf("api.ecr.%s.amazonaws.com", nt.packagesRegion))
	}

	if nt.enableGitOps {
		additionalEndpoints = append(additionalEndpoints, netest.GitOptsEndpoints...)
	}

	fmt.Println("Starting test run...")

	report := netest.ExecVSphereTests(cmd.Context(), invoker.SSH{Host: host}, netest.VSphereOptions{AdditionalEndpoints: additionalEndpoints})

	var passed int
	for _, r := range report {
		if r.Outcome == netest.Pass {
			passed++
		}
	}

	fmt.Printf("%v of %v passed\n", passed, len(report))

	fmt.Println()
	fmt.Println("FAILURES")
	for i, r := range report {
		if r.Outcome != netest.Fail {
			continue
		}
		t := tabby.New()
		t.AddLine(fmt.Sprintf("%v.", i+1), "Cmd:", r.Cmd)
		t.AddLine("", "Err:", r.Error)
		t.Print()
	}

	return nil
}
