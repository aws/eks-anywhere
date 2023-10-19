package bmc

import (
	"context"
	"fmt"
	"time"

	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"
)

// ClientFunc defines a func that returns a bmclib.Client.
type ClientFunc func(ctx context.Context, log logr.Logger, hostIP, username, password string) (*bmclib.Client, error)

// NewClientFunc returns a new BMCClientFactoryFunc. The timeout parameter determines the
// maximum time to probe for compatible interfaces.
func NewClientFunc(timeout time.Duration) ClientFunc {
	// Initializes a bmclib client based on input host and credentials
	// Establishes a connection with the bmc with client.Open
	// Returns a bmclib.Client.
	return func(ctx context.Context, log logr.Logger, hostIP, username, password string) (*bmclib.Client, error) {
		o := []bmclib.Option{}
		log = log.WithValues("host", hostIP, "username", username)
		o = append(o, bmclib.WithLogger(log))
		client := bmclib.NewClient(hostIP, username, password, o...)

		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// TODO (pokearu): Make an option
		client.Registry.Drivers = client.Registry.PreferProtocol("redfish")
		if err := client.Open(ctx); err != nil {
			md := client.GetMetadata()
			log.Info("Failed to open connection to BMC", "error", err, "providersAttempted", md.ProvidersAttempted, "successfulProvider", md.SuccessfulOpenConns)

			return nil, fmt.Errorf("failed to open connection to BMC: %w", err)
		}
		md := client.GetMetadata()
		log.Info("Connected to BMC", "providersAttempted", md.ProvidersAttempted, "successfulProvider", md.SuccessfulOpenConns)

		return client, nil
	}
}
