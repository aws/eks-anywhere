package v1alpha1

import (
	"fmt"
	"net/url"
	"strings"

	"golang.org/x/crypto/ssh"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const TinkerbellMachineConfigKind = "TinkerbellMachineConfig"

// +kubebuilder:object:generate=false
type TinkerbellMachineConfigGenerateOpt func(config *TinkerbellMachineConfigGenerate)

// Used for generating yaml for generate clusterconfig command.
func NewTinkerbellMachineConfigGenerate(name string, opts ...TinkerbellMachineConfigGenerateOpt) *TinkerbellMachineConfigGenerate {
	machineConfig := &TinkerbellMachineConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellMachineConfigKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: name,
		},
		Spec: TinkerbellMachineConfigSpec{
			HardwareSelector: HardwareSelector{},
			OSFamily:         Bottlerocket,
			Users: []UserConfiguration{
				{
					Name:              "ec2-user",
					SshAuthorizedKeys: []string{"ssh-rsa AAAA..."},
				},
			},
		},
	}

	for _, opt := range opts {
		opt(machineConfig)
	}

	return machineConfig
}

func (c *TinkerbellMachineConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellMachineConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellMachineConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func WithTemplateRef(ref ProviderRefAccessor) TinkerbellMachineConfigGenerateOpt {
	return func(c *TinkerbellMachineConfigGenerate) {
		c.Spec.TemplateRef = Ref{
			Kind: ref.Kind(),
			Name: ref.Name(),
		}
	}
}

func validateTinkerbellMachineConfig(config *TinkerbellMachineConfig) error {
	if err := validateObjectMeta(config.ObjectMeta); err != nil {
		return fmt.Errorf("TinkerbellMachineConfig: %v", err)
	}

	if len(config.Spec.HardwareSelector) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.hardwareSelector: %s", config.Name)
	}

	if len(config.Spec.HardwareSelector) != 1 {
		return fmt.Errorf(
			"TinkerbellMachineConfig: spec.hardwareSelector must contain only 1 key-value pair: %s",
			config.Name,
		)
	}

	if config.Spec.OSFamily == "" {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.osFamily: %s", config.Name)
	}

	if config.Spec.OSFamily != Ubuntu && config.Spec.OSFamily != Bottlerocket && config.Spec.OSFamily != RedHat {
		return fmt.Errorf(
			"TinkerbellMachineConfig: unsupported spec.osFamily (%v); Please use one of the following: %s, %s, %s",
			config.Spec.OSFamily,
			Ubuntu,
			RedHat,
			Bottlerocket,
		)
	}

	if config.Spec.OSImageURL != "" {
		if _, err := url.ParseRequestURI(config.Spec.OSImageURL); err != nil {
			return fmt.Errorf("parsing osImageOverride: %v", err)
		}
	}

	if len(config.Spec.Users) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: missing spec.Users: %s", config.Name)
	}

	if err := validateHostOSConfig(config.Spec.HostOSConfiguration, config.Spec.OSFamily); err != nil {
		return fmt.Errorf("HostOSConfiguration is invalid for TinkerbellMachineConfig %s: %v", config.Name, err)
	}

	return nil
}

func setTinkerbellMachineConfigDefaults(machineConfig *TinkerbellMachineConfig) {
	if machineConfig.Spec.OSFamily == "" {
		machineConfig.Spec.OSFamily = Bottlerocket
	}
}

func normalizeSSHKeys(machineConfig *TinkerbellMachineConfig) {
	_ = stripCommentsFromSSHKeys(machineConfig)
}

func stripCommentsFromSSHKeys(machine *TinkerbellMachineConfig) error {
	public, _, _, _, err := ssh.ParseAuthorizedKey([]byte(machine.Spec.Users[0].SshAuthorizedKeys[0]))
	if err != nil {
		return err
	}

	machine.Spec.Users[0].SshAuthorizedKeys[0] = strings.TrimSpace(string(ssh.MarshalAuthorizedKey(public)))

	return nil
}
