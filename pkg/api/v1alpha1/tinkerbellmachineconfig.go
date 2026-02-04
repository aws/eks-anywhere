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
			OSFamily:         Ubuntu,
			OSImageURL:       "",
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

	// Validate hardware selection (HardwareSelector vs HardwareAffinity)
	if err := validateHardwareSelection(config); err != nil {
		return err
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

// validateHardwareSelection validates the hardware selection configuration.
// HardwareSelector and HardwareAffinity are mutually exclusive.
func validateHardwareSelection(config *TinkerbellMachineConfig) error {
	hasSelector := !config.Spec.HardwareSelector.IsEmpty()
	hasAffinity := config.Spec.HardwareAffinity != nil

	// Check mutual exclusivity
	if hasSelector && hasAffinity {
		return fmt.Errorf("TinkerbellMachineConfig: hardwareSelector and hardwareAffinity are mutually exclusive: %s", config.Name)
	}

	// At least one must be specified
	if !hasSelector && !hasAffinity {
		return fmt.Errorf("TinkerbellMachineConfig: either hardwareSelector or hardwareAffinity must be specified: %s", config.Name)
	}

	// Validate HardwareSelector if present
	if hasSelector {
		if len(config.Spec.HardwareSelector) != 1 {
			return fmt.Errorf(
				"TinkerbellMachineConfig: spec.hardwareSelector must contain only 1 key-value pair: %s",
				config.Name,
			)
		}
		return nil
	}

	// Validate HardwareAffinity if present
	return validateHardwareAffinity(config.Spec.HardwareAffinity, config.Name)
}

// validateHardwareAffinity validates the HardwareAffinity configuration.
func validateHardwareAffinity(affinity *HardwareAffinity, configName string) error {
	// Required terms must have at least one entry
	if len(affinity.Required) == 0 {
		return fmt.Errorf("TinkerbellMachineConfig: hardwareAffinity.required must have at least one term: %s", configName)
	}

	// Validate each required term
	for i, term := range affinity.Required {
		if err := validateLabelSelector(&term.LabelSelector, configName, fmt.Sprintf("required[%d]", i)); err != nil {
			return err
		}
	}

	// Validate each preferred term
	for i, weightedTerm := range affinity.Preferred {
		// Validate weight range (1-100)
		if weightedTerm.Weight < 1 || weightedTerm.Weight > 100 {
			return fmt.Errorf("TinkerbellMachineConfig: preferred term weight must be in range [1, 100], got %d: %s", weightedTerm.Weight, configName)
		}

		if err := validateLabelSelector(&weightedTerm.HardwareAffinityTerm.LabelSelector, configName, fmt.Sprintf("preferred[%d]", i)); err != nil {
			return err
		}
	}

	return nil
}

// validateLabelSelector validates a Kubernetes LabelSelector.
func validateLabelSelector(selector *metav1.LabelSelector, configName, path string) error {
	// Validate matchExpressions
	for j, expr := range selector.MatchExpressions {
		// Validate operator
		if !isValidLabelSelectorOperator(expr.Operator) {
			return fmt.Errorf("TinkerbellMachineConfig: invalid matchExpression operator '%s' in %s.labelSelector.matchExpressions[%d], must be one of: In, NotIn, Exists, DoesNotExist: %s",
				expr.Operator, path, j, configName)
		}

		// Validate values for In/NotIn operators
		if expr.Operator == metav1.LabelSelectorOpIn || expr.Operator == metav1.LabelSelectorOpNotIn {
			if len(expr.Values) == 0 {
				return fmt.Errorf("TinkerbellMachineConfig: matchExpression with operator %s must have non-empty values in %s.labelSelector.matchExpressions[%d]: %s",
					expr.Operator, path, j, configName)
			}
		}

		// Validate that Exists/DoesNotExist don't have values
		if expr.Operator == metav1.LabelSelectorOpExists || expr.Operator == metav1.LabelSelectorOpDoesNotExist {
			if len(expr.Values) > 0 {
				return fmt.Errorf("TinkerbellMachineConfig: matchExpression with operator %s must not have values in %s.labelSelector.matchExpressions[%d]: %s",
					expr.Operator, path, j, configName)
			}
		}
	}

	return nil
}

// isValidLabelSelectorOperator checks if the operator is a valid LabelSelector operator.
func isValidLabelSelectorOperator(op metav1.LabelSelectorOperator) bool {
	switch op {
	case metav1.LabelSelectorOpIn, metav1.LabelSelectorOpNotIn,
		metav1.LabelSelectorOpExists, metav1.LabelSelectorOpDoesNotExist:
		return true
	default:
		return false
	}
}

// ValidateHardwareAffinityOperator validates a single operator string.
func ValidateHardwareAffinityOperator(op string) bool {
	return isValidLabelSelectorOperator(metav1.LabelSelectorOperator(op))
}

// ValidateHardwareAffinityWeight validates a weight value.
func ValidateHardwareAffinityWeight(weight int32) bool {
	return weight >= 1 && weight <= 100
}

// ValidateLabelSelectorRequirement validates a single LabelSelectorRequirement.
func ValidateLabelSelectorRequirement(req metav1.LabelSelectorRequirement) error {
	if !isValidLabelSelectorOperator(req.Operator) {
		return fmt.Errorf("invalid operator '%s'", req.Operator)
	}

	switch req.Operator {
	case metav1.LabelSelectorOpIn, metav1.LabelSelectorOpNotIn:
		if len(req.Values) == 0 {
			return fmt.Errorf("operator %s requires non-empty values", req.Operator)
		}
	case metav1.LabelSelectorOpExists, metav1.LabelSelectorOpDoesNotExist:
		if len(req.Values) > 0 {
			return fmt.Errorf("operator %s must not have values", req.Operator)
		}
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
