package v1alpha1

import (
	"errors"
	"fmt"
	"net/netip"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const TinkerbellDatacenterKind = "TinkerbellDatacenterConfig"

var unusableDHCPv6DerivedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("::/128"),
	netip.MustParsePrefix("::1/128"),
	netip.MustParsePrefix("::ffff:0:0/96"),
	netip.MustParsePrefix("fe80::/10"),
	netip.MustParsePrefix("ff00::/8"),
}

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
		Spec: TinkerbellDatacenterConfigSpec{
			TinkerbellIP: "",
		},
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

	if err := validateDHCPv6Config(config.Spec.DHCPv6Config); err != nil {
		return err
	}

	if config.Spec.IsoBoot {
		if config.Spec.HookIsoURL != "" {
			if _, err := url.ParseRequestURI(config.Spec.HookIsoURL); err != nil {
				return fmt.Errorf("parsing hookIsoURL: %v, please provide a valid URL", err)
			}
		}
	} else {
		if config.Spec.HookIsoURL != "" {
			return fmt.Errorf("isoURL can be set, only when isoBoot is set to true")
		}
	}

	return nil
}

func validateDHCPv6Config(config *DHCPv6Config) error {
	if config == nil {
		return nil
	}

	prefix, err := netip.ParsePrefix(config.DerivedAddressPool)
	if err != nil {
		return fmt.Errorf("TinkerbellDatacenterConfig: invalid DHCPv6 derivedAddressPool: %v", err)
	}

	if !usableDHCPv6DerivedPrefix(prefix) {
		return errors.New("TinkerbellDatacenterConfig: DHCPv6 derivedAddressPool must be a usable IPv6 unicast prefix between /1 and /64")
	}

	return nil
}

func usableDHCPv6DerivedPrefix(prefix netip.Prefix) bool {
	address := prefix.Addr()
	if !address.Is6() || address.Is4In6() || prefix.Bits() < 1 || prefix.Bits() > 64 {
		return false
	}

	prefix = prefix.Masked()
	for _, unusable := range unusableDHCPv6DerivedPrefixes {
		if prefix.Contains(unusable.Addr()) || unusable.Contains(prefix.Addr()) {
			return false
		}
	}

	return true
}

func validateObjectMeta(meta metav1.ObjectMeta) error {
	if meta.Name == "" {
		return errors.New("missing name")
	}

	return nil
}
