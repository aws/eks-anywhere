package cleanup

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/pkg/errors"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/providers/tinkerbell/hardware"
)

// PowerOffTinkerbellMachinesFromFile cleans up machines by powering them down.
func PowerOffTinkerbellMachinesFromFile(inventoryCSVFilePath string, ignoreErrors bool) error {
	hardwarePool, err := api.ReadTinkerbellHardwareFromFile(inventoryCSVFilePath)
	if err != nil {
		return fmt.Errorf("failed to create hardware map from inventory csv: %v", err)
	}

	logger.Info("Powering off hardware: %+v", hardwarePool)
	return PowerOffTinkerbellMachines(hardwarePool, ignoreErrors)
}

// PowerOffTinkerbellMachines powers off machines.
func PowerOffTinkerbellMachines(hardware []*hardware.Machine, ignoreErrors bool) error {
	errList := []error{}
	for _, h := range hardware {
		if err := powerOffTinkerbellMachine(h, ignoreErrors); err != nil {
			errList = append(errList, err)
		}
	}

	if len(errList) > 0 {
		return fmt.Errorf("failed to power off %d hardware: %+v", len(errList), errors.NewAggregate(errList))
	}

	return nil
}

func powerOffTinkerbellMachine(h *hardware.Machine, ignoreErrors bool) (reterror error) {
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
