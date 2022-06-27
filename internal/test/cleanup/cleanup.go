package cleanup

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/hashicorp/go-multierror"

	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/validations"
)

func CleanUpAwsTestResources(storageBucket string, maxAge string, tag string) error {
	session, err := session.NewSession()
	if err != nil {
		return fmt.Errorf("creating session: %v", err)
	}
	logger.V(1).Info("Fetching list of EC2 instances")
	key := "Integration-Test"
	value := tag
	maxAgeFloat, err := strconv.ParseFloat(maxAge, 64)
	if err != nil {
		return fmt.Errorf("parsing max age: %v", err)
	}
	results, err := ec2.ListInstances(session, key, value, maxAgeFloat)
	if err != nil {
		return fmt.Errorf("listing EC2 instances: %v", err)
	}
	logger.V(1).Info("Successfully listed EC2 instances for termination")
	if len(results) != 0 {
		logger.V(1).Info("Terminating EC2 instances")
		err = ec2.TerminateEc2Instances(session, results)
		if err != nil {
			return fmt.Errorf("terminating EC2 instacnes: %v", err)
		}
		logger.V(1).Info("Successfully terminated EC2 instances")
	} else {
		logger.V(1).Info("No EC2 instances available for termination")
	}
	logger.V(1).Info("Clean up s3 bucket objects")
	s3MaxAge := "604800" // one week
	s3MaxAgeFloat, err := strconv.ParseFloat(s3MaxAge, 64)
	if err != nil {
		return fmt.Errorf("parsing S3 max age: %v", err)
	}
	err = s3.CleanUpS3Bucket(session, storageBucket, s3MaxAgeFloat)
	if err != nil {
		return fmt.Errorf("clean up s3 bucket objects: %v", err)
	}
	logger.V(1).Info("Successfully cleaned up s3 bucket")

	return nil
}

func CleanUpVsphereTestResources(ctx context.Context, clusterName string) error {
	clusterName, err := validations.ValidateClusterNameArg([]string{clusterName})
	if err != nil {
		return fmt.Errorf("validating cluster name: %v", err)
	}
	err = VsphereRmVms(ctx, clusterName)
	if err != nil {
		return fmt.Errorf("removing vcenter vms: %v", err)
	}
	logger.V(1).Info("Vsphere vcenter vms cleanup complete")
	return nil
}

func VsphereRmVms(ctx context.Context, clusterName string, opts ...executables.GovcOpt) error {
	logger.V(1).Info("Deleting vsphere vcenter vms")
	executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, _ := filewriter.NewWriter("rmvms")
	govc := executableBuilder.BuildGovcExecutable(tmpWriter, opts...)
	defer govc.Close(ctx)

	return govc.CleanupVms(ctx, clusterName, false)
}

func CleanUpCloudstackTestResources(ctx context.Context, clusterName string, dryRun bool) (retErr error) {
	executableBuilder, close, err := executables.NewExecutableBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, err := filewriter.NewWriter("rmvms")
	if err != nil {
		return fmt.Errorf("creating filewriter for directory rmvms: %v", err)
	}
	execConfig, err := decoder.ParseCloudStackSecret()
	if err != nil {
		return fmt.Errorf("building cmk executable: %v", err)
	}
	for _, config := range execConfig.Profiles {
		cmk := executableBuilder.BuildCmkExecutable(tmpWriter, config)
		if err := cleanupCloudStackVms(ctx, cmk, clusterName, dryRun); err != nil {
			retErr = multierror.Append(retErr, err)
		}
		cmk.Close(ctx)
	}
	return retErr
}

func cleanupCloudStackVms(ctx context.Context, cmk *executables.Cmk, clusterName string, dryRun bool) error {
	if err := cmk.ValidateCloudStackConnection(ctx); err != nil {
		return fmt.Errorf("validating cloudstack connection with cloudmonkey: %v", err)
	}

	if err := cmk.CleanupVms(ctx, clusterName, dryRun); err != nil {
		return fmt.Errorf("cleaning up VMs with cloudmonkey: %v", err)
	}
	return nil
}
