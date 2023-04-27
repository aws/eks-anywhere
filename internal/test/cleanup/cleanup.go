package cleanup

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	cleanupRetries = 5
	retryBackoff   = 10 * time.Second
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
	executableBuilder, close, err := executables.InitInDockerExecutablesBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, _ := filewriter.NewWriter("rmvms")
	govc := executableBuilder.BuildGovcExecutable(tmpWriter, opts...)
	defer govc.Close(ctx)

	return govc.CleanupVms(ctx, clusterName, false)
}

func CleanUpCloudstackTestResources(ctx context.Context, clusterName string, dryRun bool) error {
	executableBuilder, close, err := executables.InitInDockerExecutablesBuilder(ctx, executables.DefaultEksaImage())
	if err != nil {
		return fmt.Errorf("unable to initialize executables: %v", err)
	}
	defer close.CheckErr(ctx)
	tmpWriter, err := filewriter.NewWriter("rmvms")
	if err != nil {
		return fmt.Errorf("creating filewriter for directory rmvms: %v", err)
	}
	execConfig, err := decoder.ParseCloudStackCredsFromEnv()
	if err != nil {
		return fmt.Errorf("parsing cloudstack credentials from environment: %v", err)
	}
	cmk, err := executableBuilder.BuildCmkExecutable(tmpWriter, execConfig)
	if err != nil {
		return fmt.Errorf("building cmk executable: %v", err)
	}
	defer cmk.Close(ctx)
	cleanupRetrier := retrier.NewWithMaxRetries(cleanupRetries, retryBackoff)

	errorsMap := map[string]error{}
	for _, profile := range execConfig.Profiles {
		if err := cleanupRetrier.Retry(func() error {
			return cmk.CleanupVms(ctx, profile.Name, clusterName, dryRun)
		}); err != nil {
			errorsMap[profile.Name] = err
		}
	}

	if len(errorsMap) > 0 {
		return fmt.Errorf("cleaning up VMs: %+v", errorsMap)
	}
	return nil
}

// NutanixTestResourcesCleanup cleans up any leftover VMs in Nutanix after a test run.
func NutanixTestResourcesCleanup(ctx context.Context, clusterName, endpoint, port string, insecure, ignoreErrors bool) error {
	creds := nutanix.GetCredsFromEnv()
	nutanixCreds := prismgoclient.Credentials{
		URL:      fmt.Sprintf("%s:%s", endpoint, port),
		Username: creds.PrismCentral.Username,
		Password: creds.PrismCentral.Password,
		Endpoint: endpoint,
		Port:     port,
		Insecure: insecure,
	}

	client, err := v3.NewV3Client(nutanixCreds)
	if err != nil {
		return fmt.Errorf("initailizing prism client: %v", err)
	}

	response, err := client.V3.ListAllVM(context.Background(), fmt.Sprintf("vm_name==%s.*", clusterName))
	if err != nil {
		return fmt.Errorf("getting ListVM response: %v", err)
	}

	for _, vm := range response.Entities {
		logger.V(4).Info("Deleting Nutanix VM", "Name", *vm.Spec.Name, "UUID:", *vm.Metadata.UUID)

		_, err = client.V3.DeleteVM(context.Background(), *vm.Metadata.UUID)
		if err != nil {
			if !ignoreErrors {
				return fmt.Errorf("deleting Nutanix VM %s: %v", *vm.Spec.Name, err)
			}
			logger.Info("Warning: Failed to delete Nutanix VM, skipping...", "Name", *vm.Spec.Name, "UUID:", *vm.Metadata.UUID)
		}
	}
	return nil
}
