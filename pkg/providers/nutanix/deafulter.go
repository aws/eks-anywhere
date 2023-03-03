package nutanix

import (
	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

type Defaulter struct{}

// NewDefaulter returns a new Defaulter.
func NewDefaulter() *Defaulter {
	return &Defaulter{}
}

func (d *Defaulter) SetDefaultsForDatacenterConfig(dcConf anywherev1.NutanixDatacenterConfig) {
	dcConf.SetDefaults()
}

func (d *Defaulter) SetDefaultsForMachineConfig(machineConf anywherev1.NutanixMachineConfig) {
	machineConf.SetDefaults()
}
