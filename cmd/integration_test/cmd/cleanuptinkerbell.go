package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/spf13/cobra"

	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/internal/test/cleanup"
	"github.com/aws/eks-anywhere/internal/test/e2e"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

var cleanUpTinkerbellCmd = &cobra.Command{
	Use:          "tinkerbell",
	Short:        "Clean up tinkerbell e2e resources",
	Long:         "Deletes vms created for e2e testing on vsphere and powers off metal machines",
	SilenceUsage: true,
	PreRun:       preRunCleanUpNutanixSetup,
	RunE: func(cmd *cobra.Command, _ []string) error {
		return cleanUpTinkerbellTestResources(cmd.Context())
	},
}

var (
	storageBucket  string
	instanceConfig string
	dryRun         bool
)

func init() {
	cleanUpInstancesCmd.AddCommand(cleanUpTinkerbellCmd)
	cleanUpTinkerbellCmd.Flags().StringVarP(&storageBucket, storageBucketFlagName, "s", "", "S3 bucket name where tinkerbell hardware inventory files are stored")
	cleanUpTinkerbellCmd.Flags().StringVar(&instanceConfig, instanceConfigFlagName, "", "File path to the instance-config.yml config")
	cleanUpTinkerbellCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Run command without deleting or powering off any resources")

	if err := cleanUpTinkerbellCmd.MarkFlagRequired(storageBucketFlagName); err != nil {
		log.Fatalf("Error marking flag %s as required: %v", storageBucketFlagName, err)
	}

	if err := cleanUpTinkerbellCmd.MarkFlagRequired(instanceConfigFlagName); err != nil {
		log.Fatalf("Error marking flag %s as required: %v", instanceConfigFlagName, err)
	}
}

// cleanUpTinkerbellTestResources deletes any test runner vm in vsphere and powers off all metal machines.
func cleanUpTinkerbellTestResources(ctx context.Context) error {
	session, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}

	deps, err := dependencies.NewFactory().WithGovc().Build(ctx)
	if err != nil {
		return err
	}
	defer deps.Close(ctx)
	govc := deps.Govc

	infraConfig, err := e2e.ReadRunnerConfig(instanceConfig)
	if err != nil {
		return fmt.Errorf("reading vms config for tests: %v", err)
	}

	govc.Configure(
		executables.GovcConfig{
			Username:   infraConfig.Username,
			Password:   infraConfig.Password,
			URL:        infraConfig.URL,
			Insecure:   infraConfig.Insecure,
			Datacenter: infraConfig.Datacenter,
		},
	)

	var errs []error

	if err := deleteSSMInstances(ctx, session); len(err) != 0 {
		errs = append(errs, err...)
	}

	if err := deleteRunners(ctx, govc, infraConfig.Folder); len(err) != 0 {
		errs = append(errs, err...)
	}

	if err := powerOffMachines(ctx, session); len(err) != 0 {
		errs = append(errs, err...)
	}

	return errors.NewAggregate(errs)
}

func deleteSSMInstances(ctx context.Context, session *session.Session) []error {
	var errs []error
	if ssmInstances, err := e2e.ListTinkerbellSSMInstances(ctx, session); err != nil {
		errs = append(errs, fmt.Errorf("listing ssm instances: %w", err))
	} else if dryRun {
		logger.Info("Found SSM instances", "instanceIDs", ssmInstances.InstanceIDs, "activationIDs", ssmInstances.ActivationIDs)
	} else {
		if _, err := ssm.DeregisterInstances(session, ssmInstances.InstanceIDs...); err != nil {
			errs = append(errs, fmt.Errorf("deleting ssm instances: %w", err))
		}
		if _, err := ssm.DeleteActivations(session, ssmInstances.ActivationIDs...); err != nil {
			errs = append(errs, fmt.Errorf("deleting ssm activations: %w", err))
		}
	}

	return errs
}

func deleteRunners(ctx context.Context, govc *executables.Govc, folder string) []error {
	var errs []error
	if runners, err := govc.ListVMs(ctx, folder); err != nil {
		errs = append(errs, fmt.Errorf("listing tinkerbell runners: %w", err))
	} else if dryRun {
		logger.Info("Found VM Runners", "vms", runners)
	} else {
		for _, vm := range runners {
			if err := govc.DeleteVM(ctx, vm.Path); err != nil {
				errs = append(errs, fmt.Errorf("deleting tinkerbell runner %s: %w", vm, err))
			}
		}
	}

	return errs
}

func powerOffMachines(_ context.Context, session *session.Session) []error {
	var errs []error
	if machines, err := e2e.ReadTinkerbellMachinePool(session, storageBucket); err != nil {
		errs = append(errs, fmt.Errorf("reading tinkerbell machine pool: %v", err))
	} else if dryRun {
		logger.Info("Metal machine pool", "machines", names(machines))
	} else {
		if err = cleanup.PowerOffTinkerbellMachines(machines, true); err != nil {
			errs = append(errs, fmt.Errorf("powering off tinkerbell machines: %v", err))
		}
	}

	return errs
}

func names(h []*hardware.Machine) []string {
	names := make([]string, 0, len(h))
	for _, m := range h {
		names = append(names, m.Hostname)
	}

	return names
}
