package config

import (
	_ "embed"
	"os"
)

const (
	EksavSphereUsernameKey = "EKSA_VSPHERE_USERNAME"
	EksavSpherePasswordKey = "EKSA_VSPHERE_PASSWORD"
	// EksavSphereCPUsernameKey holds Username for cloud provider.
	EksavSphereCPUsernameKey = "EKSA_VSPHERE_CP_USERNAME"
	// EksavSphereCPPasswordKey holds Password for cloud provider.
	EksavSphereCPPasswordKey = "EKSA_VSPHERE_CP_PASSWORD"
	// EksavSphereCSIUsernameKey holds Username and password for the CSI driver.
	EksavSphereCSIUsernameKey = "EKSA_VSPHERE_CSI_USERNAME"
	EksavSphereCSIPasswordKey = "EKSA_VSPHERE_CSI_PASSWORD"
)

type VSphereUserConfig struct {
	EksaVsphereUsername    string
	EksaVspherePassword    string
	EksaVsphereCPUsername  string
	EksaVsphereCPPassword  string
	EksaVsphereCSIUsername string
	EksaVsphereCSIPassword string
}

//go:embed static/globalPrivs.json
var VSphereGlobalPrivsFile string

//go:embed static/eksUserPrivs.json
var VSphereUserPrivsFile string

//go:embed static/adminPrivs.json
var VSphereAdminPrivsFile string

//go:embed static/cnsDatastorePrivs.json
var VSphereCnsDatastorePrivsFile string

//go:embed static/cnsSearchAndSpbmPrivs.json
var VSphereCnsSearchAndSpbmPrivsFile string

//go:embed static/cnsVmPrivs.json
var VSphereCnsVmPrivsFile string

//go:embed static/cnsHostConfigStorage.json
var VSphereCnsHostConfigStorageFile string

//go:embed static/readOnlyPrivs.json
var VSphereReadOnlyPrivs string

func NewVsphereUserConfig() *VSphereUserConfig {
	eksaVsphereUsername := os.Getenv(EksavSphereUsernameKey)
	eksaVspherePassword := os.Getenv(EksavSpherePasswordKey)

	// Cloud provider credentials
	eksaCPUsername := os.Getenv(EksavSphereCPUsernameKey)
	eksaCPPassword := os.Getenv(EksavSphereCPPasswordKey)

	if eksaCPUsername == "" {
		eksaCPUsername = eksaVsphereUsername
		eksaCPPassword = eksaVspherePassword
	}
	// CSI driver credentials
	eksaCSIUsername := os.Getenv(EksavSphereCSIUsernameKey)
	eksaCSIPassword := os.Getenv(EksavSphereCSIPasswordKey)
	if eksaCSIUsername == "" {
		eksaCSIUsername = eksaVsphereUsername
		eksaCSIPassword = eksaVspherePassword
	}

	vuc := VSphereUserConfig{
		eksaVsphereUsername,
		eksaVspherePassword,
		eksaCPUsername,
		eksaCPPassword,
		eksaCSIUsername,
		eksaCSIPassword,
	}

	return &vuc
}
