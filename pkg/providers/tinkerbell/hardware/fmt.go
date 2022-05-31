package hardware

import "fmt"

func formatBMCRef(m Machine) string {
	return fmt.Sprintf("bmc-%s", m.Hostname)
}

func formatBMCSecretRef(m Machine) string {
	return fmt.Sprintf("%s-auth", formatBMCRef(m))
}
