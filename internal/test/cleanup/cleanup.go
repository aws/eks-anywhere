package cleanup

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"
	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/ec2"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/filewriter"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
	"github.com/aws/eks-anywhere/pkg/providers/nutanix"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
	"github.com/aws/eks-anywhere/pkg/retrier"
	"github.com/aws/eks-anywhere/pkg/validations"
)

const (
	cleanupRetries       = 5
	retryBackoff         = 10 * time.Second
	cloudstackNetworkVar = "T_CLOUDSTACK_NETWORK"
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
	logger.V(1).Info("Deleting vsphere vcenter vms", "clusterName", clusterName)
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

// CloudstackTestResources cleans up resources on the CloudStack environment.
// This can include VMs as well as duplicate networks.
func CloudstackTestResources(ctx context.Context, clusterName string, dryRun bool, deleteDuplicateNetworks bool) error {
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

	return cleanupCloudstackDuplicateNetworks(ctx, cmk, execConfig, deleteDuplicateNetworks)
}

func cleanupCloudstackDuplicateNetworks(ctx context.Context, cmk *executables.Cmk, execConfig *decoder.CloudStackExecConfig, deleteDuplicateNetworks bool) error {
	if !deleteDuplicateNetworks {
		return nil
	}

	networkName, set := os.LookupEnv(cloudstackNetworkVar)
	if !set {
		return fmt.Errorf("ensuring no duplicate networks, %s is not set", cloudstackNetworkVar)
	}

	for _, profile := range execConfig.Profiles {
		if err := cmk.EnsureNoDuplicateNetwork(ctx, profile.Name, networkName); err != nil {
			return err
		}
	}
	return nil
}

// NutanixTestResources cleans up any leftover VMs in Nutanix after a test run.
func NutanixTestResources(clusterName, endpoint, port string, insecure, ignoreErrors bool) error {
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

// TinkerbellTestResources cleans up machines by powering them down.
func TinkerbellTestResources(inventoryCSVFilePath string, ignoreErrors bool) error {
	hardwarePool, err := api.NewHardwareMapFromFile(inventoryCSVFilePath)
	if err != nil {
		return fmt.Errorf("failed to create hardware map from inventory csv: %v", err)
	}

	logger.Info("Powering off hardware: %+v", hardwarePool)
	return powerOffHardwarePool(hardwarePool, ignoreErrors)
}

func powerOffHardwarePool(hardware map[string]*hardware.Machine, ignoreErrors bool) error {
	errList := []error{}
	for _, h := range hardware {
		if err := powerOffHardware(h, ignoreErrors); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("failed to power off %d hardware: %+v", len(errList), errors.NewAggregate(errList))
	}

	return nil
}

func powerOffHardware(h *hardware.Machine, ignoreErrors bool) (reterror error) {
	ctx, done := context.WithTimeout(context.Background(), 2*time.Minute)
	defer done()
	bmcClient := newBmclibClient(logr.Discard(), h.BMCIPAddress, h.BMCUsername, h.BMCPassword)

	if err := bmcClient.Open(ctx); err != nil {
		md := bmcClient.GetMetadata()
		logger.Info("Warning: Failed to open connection to BMC: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)
		return handlePowerOffHardwareError(err, ignoreErrors)
	}

	md := bmcClient.GetMetadata()
	logger.Info("Connected to BMC: hardware: %v, providersAttempted: %v, successfulProvider: %v", h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)

	defer func() {
		if err := bmcClient.Close(ctx); err != nil {
			md := bmcClient.GetMetadata()
			logger.Info("Warning: BMC close connection failed: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.FailedProviderDetail)
			reterror = handlePowerOffHardwareError(err, ignoreErrors)
		}
	}()

	state, err := bmcClient.GetPowerState(ctx)
	if err != nil {
		state = "unknown"
	}
	if strings.Contains(strings.ToLower(state), "off") {
		return nil
	}

	if _, err := bmcClient.SetPowerState(ctx, "off"); err != nil {
		md := bmcClient.GetMetadata()
		logger.Info("Warning: failed to power off hardware: %v, hardware: %v, providersAttempted: %v, failedProviderDetail: %v", err, h.BMCIPAddress, md.ProvidersAttempted, md.SuccessfulOpenConns)
		return handlePowerOffHardwareError(err, ignoreErrors)
	}

	return nil
}

func handlePowerOffHardwareError(err error, ignoreErrors bool) error {
	if err != nil && !ignoreErrors {
		return err
	}
	return nil
}

// newBmclibClient creates a new BMClib client.
func newBmclibClient(log logr.Logger, hostIP, username, password string) *bmclib.Client {
	o := []bmclib.Option{}
	log = log.WithValues("host", hostIP, "username", username)
	o = append(o, bmclib.WithLogger(log))
	client := bmclib.NewClient(hostIP, username, password, o...)
	client.Registry.Drivers = client.Registry.PreferProtocol("redfish")

	return client
}
