package flags

// TinkerbellBootstrapIP is used to override the Tinkerbell IP for serving a Tinkerbell stack
// from an admin machine.
var TinkerbellBootstrapIP = Flag[string]{
	Name:  "tinkerbell-bootstrap-ip",
	Usage: "The IP used to expose the Tinkerbell stack from the bootstrap cluster",
}
