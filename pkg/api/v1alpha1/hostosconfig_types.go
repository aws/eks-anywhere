package v1alpha1

// HostOSConfiguration defines the configuration settings on the host OS.
type HostOSConfiguration struct {
	NTPConfiguration *NTPConfiguration `json:"ntpConfiguration"`
}

// NTPConfiguration defines the NTP configuration on the host OS.
type NTPConfiguration struct {
	Servers []string `json:"servers"`
}
