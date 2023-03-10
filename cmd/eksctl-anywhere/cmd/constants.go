package cmd

const (
	imagesTarFile               = "images.tar"
	eksaToolsImageTarFile       = "tools-image.tar"
	cpWaitTimeoutFlag           = "control-plane-wait-timeout"
	externalEtcdWaitTimeoutFlag = "external-etcd-wait-timeout"
	perMachineWaitTimeoutFlag   = "per-machine-wait-timeout"
	unhealthyMachineTimeoutFlag = "unhealthy-machine-timeout"
	nodeStartupTimeoutFlag      = "node-startup-timeout"
	noTimeoutsFlag              = "no-timeouts"
)

type Operation int

const (
	Create  Operation = 0
	Upgrade Operation = 1
	Delete  Operation = 2
)
