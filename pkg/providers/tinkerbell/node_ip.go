package tinkerbell

import (
	_ "embed"
	"fmt"
	"net/netip"
	"strings"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

//go:embed config/ipv6-node-ip.py
var ipv6NodeIPScript string

const ipv6NodeIPScriptPath = "/usr/local/bin/eks-a-set-node-ip"

func addIPv6NodeIPBootstrap(values map[string]interface{}, spec *cluster.Spec, dc v1alpha1.TinkerbellDatacenterConfigSpec, osFamily v1alpha1.OSFamily) error {
	if osFamily == v1alpha1.Bottlerocket {
		return nil
	}

	tinkerbellIP, err := netip.ParseAddr(dc.TinkerbellIP)
	if err != nil || !tinkerbellIP.Is6() || tinkerbellIP.Is4In6() {
		if dc.DHCPv6Config == nil {
			return nil
		}
		return fmt.Errorf("invalid Tinkerbell IPv6 address %q", dc.TinkerbellIP)
	}
	args := []string{tinkerbellIP.String()}
	if address, err := netip.ParseAddr(spec.Cluster.Spec.ControlPlaneConfiguration.Endpoint.Host); err == nil && address.Is6() {
		args = append(args, address.String())
	}
	values["ipv6NodeIPScript"] = strings.TrimSuffix(ipv6NodeIPScript, "\n")
	values["ipv6NodeIPCommand"] = ipv6NodeIPScriptPath + " " + strings.Join(args, " ") + " || exit 1"
	return nil
}
