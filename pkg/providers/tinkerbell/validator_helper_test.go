package tinkerbell

import "github.com/aws/eks-anywhere/pkg/hardware"

// SetValidatorHardwareConfig sets v internal hardwareConfig member to cfg.
func SetValidatorHardwareConfig(v *Validator, cfg hardware.HardwareConfig) {
	v.hardwareConfig = cfg
}
