package snow

import (
	"context"

	"github.com/aws/eks-anywhere/pkg/cluster"
)

type ConfigManager struct {
	validator  *AwsClientValidator
	defaulters *Defaulters
}

func NewConfigManager(defaulters *Defaulters, validators *AwsClientValidator) *ConfigManager {
	return &ConfigManager{
		validator:  validators,
		defaulters: defaulters,
	}
}

func (cm *ConfigManager) SetDefaultsAndValidate(ctx context.Context, config *cluster.Config) error {
	configManager := cluster.NewConfigManager()
	if err := configManager.Register(cm.snowEntry(ctx)); err != nil {
		return err
	}

	if err := configManager.SetDefaults(config); err != nil {
		return err
	}

	if err := configManager.Validate(config); err != nil {
		return err
	}

	return nil
}

func (cm *ConfigManager) snowEntry(ctx context.Context) *cluster.ConfigManagerEntry {
	return &cluster.ConfigManagerEntry{
		Defaulters: []cluster.Defaulter{
			func(c *cluster.Config) error {
				return cm.defaulters.GenerateDefaultSshKeys(ctx, c.SnowMachineConfigs)
			},
		},
		Validations: []cluster.Validation{
			func(c *cluster.Config) error {
				for _, m := range c.SnowMachineConfigs {
					if err := cm.validator.ValidateMachineDeviceIPs(ctx, m); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *cluster.Config) error {
				for _, m := range c.SnowMachineConfigs {
					if err := cm.validator.ValidateEC2ImageExistsOnDevice(ctx, m); err != nil {
						return err
					}
				}
				return nil
			},
			func(c *cluster.Config) error {
				for _, m := range c.SnowMachineConfigs {
					if err := cm.validator.ValidateEC2SshKeyNameExists(ctx, m); err != nil {
						return err
					}
				}
				return nil
			},
		},
	}
}
