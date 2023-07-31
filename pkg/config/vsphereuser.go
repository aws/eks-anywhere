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
)

type VSphereUserConfig struct {
	EksaVsphereUsername   string
	EksaVspherePassword   string
	EksaVsphereCPUsername string
	EksaVsphereCPPassword string
}

//go:embed static/globalPrivs.json
var VSphereGlobalPrivsFile string

//go:embed static/eksUserPrivs.json
var VSphereUserPrivsFile string

//go:embed static/adminPrivs.json
var VSphereAdminPrivsFile string

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

	vuc := VSphereUserConfig{
		eksaVsphereUsername,
		eksaVspherePassword,
		eksaCPUsername,
		eksaCPPassword,
	}

	return &vuc
}
