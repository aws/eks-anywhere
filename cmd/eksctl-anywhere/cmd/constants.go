package cmd

const (
	timeoutWarningTemplate      = "Warning: Invalid %s value: %s, Using the default timeout instead: %s"
	imagesTarFile               = "images.tar"
	eksaToolsImageTarFile       = "tools-image.tar"
	cpWaitTimeoutFlag           = "control-plane-wait-timeout"
	externalEtcdWaitTimeoutFlag = "external-etcd-wait-timeout"
	perMachineWaitTimeoutFlag   = "per-machine-wait-timeout"
)
