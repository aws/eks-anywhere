package tinkerbell

import (
	"net"
	"net/url"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

const (
	// SmeeHTTPPort serves boot requests from Docker during bootstrap.
	SmeeHTTPPort = "7171"
	// The consolidated Kubernetes stack serves boot and metadata requests together.
	tinkerbellHTTPPort = "7172"
)

// TinkerbellISOURL builds the media URL for the provisioning phase.
// IsoBootURL, when configured, is preserved across bootstrap and move.
// A nonempty bootstrapIP selects Docker Smee; otherwise the Kubernetes stack
// serves the default URL. CAPT inserts the Hardware MAC before the ISO filename.
func TinkerbellISOURL(dc v1alpha1.TinkerbellDatacenterConfigSpec, bootstrapIP string) string {
	if dc.IsoBootURL != "" {
		return dc.IsoBootURL
	}

	tinkerbellIP := dc.TinkerbellIP
	port := tinkerbellHTTPPort
	if bootstrapIP != "" {
		tinkerbellIP = bootstrapIP
		port = SmeeHTTPPort
	}
	isoPath := "/iso/hook.iso"
	if addressStringIsIPv6(tinkerbellIP) {
		isoPath = "/iso6/hook.iso"
	}

	return (&url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(tinkerbellIP, port),
		Path:   isoPath,
	}).String()
}
