package v1alpha1

import (
	"errors"
	"fmt"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const TinkerbellDatacenterKind = "TinkerbellDatacenterConfig"

// Used for generating yaml for generate clusterconfig command.
func NewTinkerbellDatacenterConfigGenerate(clusterName string) *TinkerbellDatacenterConfigGenerate {
	return &TinkerbellDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       TinkerbellDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: TinkerbellDatacenterConfigSpec{},
	}
}

func (c *TinkerbellDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *TinkerbellDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *TinkerbellDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetTinkerbellDatacenterConfig(fileName string) (*TinkerbellDatacenterConfig, error) {
	var clusterConfig TinkerbellDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}

func validateDatacenterConfig(config *TinkerbellDatacenterConfig) error {
	if config.Spec.OSImageURL != "" {
		if _, err := url.ParseRequestURI(config.Spec.OSImageURL); err != nil {
			return fmt.Errorf("parsing osImageOverride: %v", err)
		}
	}

	if config.Spec.HookImagesURLPath != "" {
		if _, err := url.ParseRequestURI(config.Spec.HookImagesURLPath); err != nil {
			return fmt.Errorf("parsing hookOverride: %v", err)
		}
	}

	if err := validateObjectMeta(config.ObjectMeta); err != nil {
		return fmt.Errorf("TinkerbellDatacenterConfig: %v", err)
	}

	if config.Spec.TinkerbellIP == "" {
		return errors.New("TinkerbellDatacenterConfig: missing spec.tinkerbellIP field")
	}

	if err := networkutils.ValidateIP(config.Spec.TinkerbellIP); err != nil {
		return fmt.Errorf("TinkerbellDatacenterConfig: invalid tinkerbell ip: %v", err)
	}

	return nil
}

func validateObjectMeta(meta metav1.ObjectMeta) error {
	if meta.Name == "" {
		return errors.New("missing name")
	}

	return nil
}
