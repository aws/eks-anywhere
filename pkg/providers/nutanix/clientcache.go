package nutanix

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"strconv"

	prismgoclient "github.com/nutanix-cloud-native/prism-go-client"
	"github.com/nutanix-cloud-native/prism-go-client/environment/credentials"
	v3 "github.com/nutanix-cloud-native/prism-go-client/v3"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
)

// ClientCache is a map of NutanixDatacenterConfig name to Nutanix client.
type ClientCache struct {
	clients map[string]Client
}

// NewClientCache returns a new ClientCache.
func NewClientCache() *ClientCache {
	return &ClientCache{
		clients: make(map[string]Client),
	}
}

// GetNutanixClient returns a Nutanix client for the given NutanixDatacenterConfig.
func (cb *ClientCache) GetNutanixClient(datacenterConfig *anywherev1.NutanixDatacenterConfig, creds credentials.BasicAuthCredential) (Client, error) {
	if client, ok := cb.clients[datacenterConfig.Name]; ok {
		return client, nil
	}

	clientOpts := make([]v3.ClientOption, 0)
	if datacenterConfig.Spec.AdditionalTrustBundle != "" {
		block, _ := pem.Decode([]byte(datacenterConfig.Spec.AdditionalTrustBundle))
		certs, err := x509.ParseCertificates(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("unable to parse additional trust bundle %s: %v", datacenterConfig.Spec.AdditionalTrustBundle, err)
		}
		if len(certs) == 0 {
			return nil, fmt.Errorf("unable to extract certs from the addtional trust bundle %s", datacenterConfig.Spec.AdditionalTrustBundle)
		}
		clientOpts = append(clientOpts, v3.WithCertificate(certs[0]))
	}

	endpoint := datacenterConfig.Spec.Endpoint
	port := datacenterConfig.Spec.Port
	url := net.JoinHostPort(endpoint, strconv.Itoa(port))
	nutanixCreds := prismgoclient.Credentials{
		URL:      url,
		Username: creds.PrismCentral.Username,
		Password: creds.PrismCentral.Password,
		Endpoint: endpoint,
		Port:     fmt.Sprintf("%d", port),
		Insecure: datacenterConfig.Spec.Insecure,
	}

	client, err := v3.NewV3Client(nutanixCreds, clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("error creating nutanix client: %v", err)
	}

	cb.clients[datacenterConfig.Name] = client.V3
	return client.V3, nil
}
