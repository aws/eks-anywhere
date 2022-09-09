package executables

import "regexp"

const (
	generatedDir = "generated"
	overridesDir = "overrides"
)

// Cannot be const because regexp type can't be, but intended to be unchanging.
var (
	connectionRefusedRegex = regexp.MustCompile("The connection to the server .* was refused")
	ioTimeoutRegex         = regexp.MustCompile("Unable to connect to the server.*i/o timeout.*")
)
