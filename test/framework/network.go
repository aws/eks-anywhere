package framework

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/networkutils"
)

const secondaryInterfaceName = "eksa-e2e-net0"

func PopIPFromEnv(ipPoolEnvVar string) (string, error) {
	ipPool, err := networkutils.NewIPPoolFromEnv(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("popping IP from environment: %v", err)
	}

	ip, popErr := ipPool.PopIP()
	if popErr != nil {
		return "", fmt.Errorf("failed to get an ip address from the cluster ip pool env var %s: %v", ipPoolEnvVar, popErr)
	}

	// PopIPFromEnv will remove the ip from the pool.
	// Therefore, we rewrite the envvar to the system so the next caller can pick from remaining ips in the pool
	err = ipPool.ToEnvVar(ipPoolEnvVar)
	if err != nil {
		return "", fmt.Errorf("popping IP from environment: %v", err)
	}

	return ip, nil
}

func GenerateUniqueIp(cidr string) (string, error) {
	ipgen := networkutils.NewIPGenerator(&networkutils.DefaultNetClient{})
	ip, err := ipgen.GenerateUniqueIP(cidr, nil)
	if err != nil {
		return "", fmt.Errorf("getting unique IP for cidr %s: %v", cidr, err)
	}
	return ip, nil
}

func GetIP(cidr, ipEnvVar string) (string, error) {
	value, ok := os.LookupEnv(ipEnvVar)
	var ip string
	var err error
	if ok && value != "" {
		ip, err = PopIPFromEnv(ipEnvVar)
		if err != nil {
			logger.V(2).Info("WARN: failed to pop ip from environment, attempting to generate unique ip")
			ip, err = GenerateUniqueIp(cidr)
			if err != nil {
				return "", fmt.Errorf("failed to generate ip for cidr %s: %v", cidr, err)
			}
		}
	} else {
		ip, err = GenerateUniqueIp(cidr)
		if err != nil {
			return "", fmt.Errorf("GenerateUniqueIp() failed to generate ip for cidr %s: %v", cidr, err)
		}
	}
	return ip, nil
}

// CreateSecondaryNetworkInterface creates a secondary network interface on the
// admin (test runner) machine holding a free IP from cidr and returns that IP.
// It is used to simulate a multi-NIC admin machine: the interface is an ipvlan
// (L2) child of the NIC already attached to cidr, so its IP is reachable from
// the bare metal hardware on that network, while remaining a distinct interface
// that does not carry the machine's default route. Passing its IP as
// --tinkerbell-bootstrap-ip exercises the advertise-vs-bind split in the
// Tinkerbell stack (listeners must bind the bootstrap IP explicitly instead of
// auto-detecting the default-route NIC).
// The interface is deleted via t.Cleanup at the end of the test.
func CreateSecondaryNetworkInterface(t *testing.T, cidr string) string {
	parent, prefixLen, err := interfaceForCIDR(cidr)
	if err != nil {
		t.Fatalf("finding parent interface for secondary network interface: %v", err)
	}

	ip, err := GetIP(cidr, ClusterIPPoolEnvVar)
	if err != nil {
		t.Fatalf("getting IP for secondary network interface: %v", err)
	}

	// Remove any leftover interface from a previous run before creating a new one.
	if err := DeleteSecondaryNetworkInterfaceIfExists(); err != nil {
		t.Fatalf("deleting leftover secondary network interface: %v", err)
	}

	t.Logf("Creating secondary network interface %s (%s/%d) on parent %s", secondaryInterfaceName, ip, prefixLen, parent)
	// ipvlan in L2 mode reuses the parent's MAC address, which keeps the new
	// interface functional on networks that filter unknown MACs (unlike macvlan).
	if err := runIPCommand("link", "add", secondaryInterfaceName, "link", parent, "type", "ipvlan", "mode", "l2"); err != nil {
		t.Fatalf("creating secondary network interface: %v", err)
	}
	t.Cleanup(func() {
		if err := runIPCommand("link", "del", secondaryInterfaceName); err != nil {
			t.Logf("WARN: failed deleting secondary network interface %s: %v", secondaryInterfaceName, err)
		}
	})

	// Prevent NetworkManager (if present) from taking over the new interface
	// and flushing its statically assigned address while it attempts DHCP.
	setInterfaceUnmanagedBestEffort(t, secondaryInterfaceName)

	addr := fmt.Sprintf("%s/%d", ip, prefixLen)
	if err := runIPCommand("addr", "add", addr, "dev", secondaryInterfaceName); err != nil {
		t.Fatalf("assigning IP to secondary network interface: %v", err)
	}
	if err := runIPCommand("link", "set", secondaryInterfaceName, "up"); err != nil {
		t.Fatalf("bringing up secondary network interface: %v", err)
	}

	verifyInterfaceIPv4Persists(t, ip, addr)

	return ip
}

// verifyInterfaceIPv4Persists checks that the address assigned to the
// secondary interface sticks. A host network manager (NetworkManager or a
// catch-all systemd-networkd .network file) can asynchronously reset the
// interface and remove the address; re-add once before giving up.
func verifyInterfaceIPv4Persists(t *testing.T, ip, addr string) {
	if interfaceHasIPv4(secondaryInterfaceName, ip, 3*time.Second) {
		return
	}
	t.Logf("WARN: address %s disappeared from %s, re-adding", addr, secondaryInterfaceName)
	if err := runIPCommand("addr", "replace", addr, "dev", secondaryInterfaceName); err != nil {
		t.Fatalf("re-assigning IP to secondary network interface: %v", err)
	}
	if !interfaceHasIPv4(secondaryInterfaceName, ip, 5*time.Second) {
		t.Fatalf(
			"address %s does not persist on %s: a host network manager is likely resetting the interface. "+
				"Mark it unmanaged (e.g. 'nmcli device set %s managed no' or a systemd-networkd Unmanaged=yes drop-in) and retry",
			addr, secondaryInterfaceName, secondaryInterfaceName,
		)
	}
}

// interfaceHasIPv4 waits up to timeout for the named interface to hold ip,
// rechecking briefly to catch asynchronous removal by host network managers.
func interfaceHasIPv4(name, ip string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	seen := false
	for time.Now().Before(deadline) {
		seen = false
		if iface, err := net.InterfaceByName(name); err == nil {
			if addrs, err := iface.Addrs(); err == nil {
				for _, a := range addrs {
					if ipNet, ok := a.(*net.IPNet); ok && ipNet.IP.String() == ip {
						seen = true
					}
				}
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	return seen
}

// setInterfaceUnmanagedBestEffort tells host network managers to leave the
// interface alone, so they do not flush its statically assigned address while
// attempting to configure it (e.g. DHCP):
//   - NetworkManager: mark the device unmanaged via nmcli.
//   - systemd-networkd: install a .network drop-in with Unmanaged=yes. The
//     drop-in in /etc/systemd/network sorts before netplan's generated
//     catch-all in /run/systemd/network, so it wins the first-match.
//
// All steps are best-effort: failures are logged and non-fatal since the
// address verification in CreateSecondaryNetworkInterface catches actual
// interference and fails with instructions.
func setInterfaceUnmanagedBestEffort(t *testing.T, name string) {
	if _, err := exec.LookPath("nmcli"); err == nil {
		if err := runPrivilegedCommand("nmcli", "device", "set", name, "managed", "no"); err != nil {
			t.Logf("WARN: could not mark %s unmanaged in NetworkManager: %v", name, err)
		}
	}

	if err := exec.Command("systemctl", "is-active", "--quiet", "systemd-networkd").Run(); err == nil {
		dropin := fmt.Sprintf("/etc/systemd/network/05-%s.network", name)
		content := fmt.Sprintf("[Match]\nName=%s\n\n[Link]\nUnmanaged=yes\n", name)
		if err := writeFileAsRoot(dropin, content); err != nil {
			t.Logf("WARN: could not install systemd-networkd unmanaged drop-in %s: %v", dropin, err)
			return
		}
		if err := runPrivilegedCommand("networkctl", "reload"); err != nil {
			t.Logf("WARN: could not reload systemd-networkd config: %v", err)
		}
	}
}

// writeFileAsRoot writes content to path, falling back to passwordless sudo
// tee when the test is not running as root.
func writeFileAsRoot(path, content string) error {
	if err := os.WriteFile(path, []byte(content), 0o644); err == nil {
		return nil
	} else if !os.IsPermission(err) {
		return err
	}
	cmd := exec.Command("sudo", "-n", "tee", path)
	cmd.Stdin = strings.NewReader(content)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("sudo tee %s: %v: %s (CI runners execute tests as root; for local runs create the file manually)", path, err, string(out))
	}
	return nil
}

// DeleteSecondaryNetworkInterfaceIfExists removes the secondary network
// interface created by CreateSecondaryNetworkInterface if it is present.
// It is a no-op when the interface does not exist, so it is safe to call as a
// defensive cleanup from any test: a leaked interface (e.g. after a hard-killed
// test run) holds an IP from the provisioning subnet pool that could otherwise
// conflict with IPs allocated to later tests on the same admin machine.
func DeleteSecondaryNetworkInterfaceIfExists() error {
	if _, err := net.InterfaceByName(secondaryInterfaceName); err != nil {
		// Interface not present; nothing to clean up.
		return nil
	}
	return runIPCommand("link", "del", secondaryInterfaceName)
}

// interfaceForCIDR returns the name and IPv4 prefix length of the first
// network interface attached to the network containing cidr. The provisioning
// CIDR (T_TINKERBELL_CP_NETWORK_CIDR) is typically an IP reservation range
// carved out of a larger L2 network (e.g. a 10.80.8.128/25 pool on a
// 10.80.8.0/22 network), so the interface is matched by checking whether its
// own connected network contains the CIDR's base address, not whether the
// interface's IP falls inside the CIDR.
func interfaceForCIDR(cidr string) (string, int, error) {
	_, ipNet, err := net.ParseCIDR(cidr)
	if err != nil {
		return "", 0, fmt.Errorf("parsing cidr %s: %v", cidr, err)
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		return "", 0, fmt.Errorf("listing network interfaces: %v", err)
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			if a, ok := addr.(*net.IPNet); ok && a.IP.To4() != nil && a.Contains(ipNet.IP) {
				prefixLen, _ := a.Mask.Size()
				return iface.Name, prefixLen, nil
			}
		}
	}

	return "", 0, fmt.Errorf("no interface attached to a network containing %s", cidr)
}

// runIPCommand runs an iproute2 command, retrying with sudo when the test is
// not running as root.
func runIPCommand(args ...string) error {
	return runPrivilegedCommand("ip", args...)
}

// runPrivilegedCommand runs a command, retrying with passwordless sudo when it
// fails due to insufficient privileges.
func runPrivilegedCommand(name string, args ...string) error {
	out, err := exec.Command(name, args...).CombinedOutput()
	if err == nil {
		return nil
	}
	lowerOut := strings.ToLower(string(out))
	if strings.Contains(lowerOut, "not permitted") || strings.Contains(lowerOut, "permission denied") || strings.Contains(lowerOut, "insufficient privileges") {
		if sudoOut, sudoErr := exec.Command("sudo", append([]string{"-n", name}, args...)...).CombinedOutput(); sudoErr != nil {
			return fmt.Errorf(
				"%s %s: %v: %s (CI runners execute tests as root; for local runs configure passwordless sudo for %s or run the test binary as root)",
				name, strings.Join(args, " "), sudoErr, string(sudoOut), name,
			)
		}
		return nil
	}
	return fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, string(out))
}
