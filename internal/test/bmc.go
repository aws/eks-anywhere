package test

import (
	"github.com/bmc-toolbox/bmclib/v2"
	"github.com/go-logr/logr"
)

// NewBmclibClient creates a new BMClib client.
func NewBmclibClient(log logr.Logger, hostIP, username, password string) *bmclib.Client {
	o := []bmclib.Option{}
	log = log.WithValues("host", hostIP, "username", username)
	o = append(o, bmclib.WithLogger(log))
	client := bmclib.NewClient(hostIP, username, password, o...)
	client.Registry.Drivers = client.Registry.PreferProtocol("redfish")

	return client
}
