package cmd

import (
	"github.com/spf13/pflag"
)

// TinkerbellHardwareCSVFlagName is the flag used for providing hardware CSV used by the Tinkerbell
// provider.
const TinkerbellHardwareCSVFlagName = "hardware-csv"

// PopulateTinkerbellHardwareCSVFlag populates s with the TinkerbellHardwareCSVFlagName.
// The value provided to the CLI will be written to field.
func PopulateTinkerbelHardwareCSVFlag(field *string, s *pflag.FlagSet) {
	s.StringVar(field, TinkerbellHardwareCSVFlagName, "",
		"A file path to a CSV file containing hardware data to be submitted to the cluster for provisioning.")
}
