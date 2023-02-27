package v1alpha1

import (
	"fmt"
	"net/url"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const CloudStackDatacenterKind = "CloudStackDatacenterConfig"

// Used for generating yaml for generate clusterconfig command.
func NewCloudStackDatacenterConfigGenerate(clusterName string) *CloudStackDatacenterConfigGenerate {
	return &CloudStackDatacenterConfigGenerate{
		TypeMeta: metav1.TypeMeta{
			Kind:       CloudStackDatacenterKind,
			APIVersion: SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: ObjectMeta{
			Name: clusterName,
		},
		Spec: CloudStackDatacenterConfigSpec{
			AvailabilityZones: []CloudStackAvailabilityZone{
				{
					Name: "az-1",
					Zone: CloudStackZone{
						Network: CloudStackResourceIdentifier{},
					},
					CredentialsRef: "global",
					Account:        "admin",
					Domain:         "domain1",
				},
			},
		},
	}
}

func (c *CloudStackDatacenterConfigGenerate) APIVersion() string {
	return c.TypeMeta.APIVersion
}

func (c *CloudStackDatacenterConfigGenerate) Kind() string {
	return c.TypeMeta.Kind
}

func (c *CloudStackDatacenterConfigGenerate) Name() string {
	return c.ObjectMeta.Name
}

func GetCloudStackDatacenterConfig(fileName string) (*CloudStackDatacenterConfig, error) {
	var clusterConfig CloudStackDatacenterConfig
	err := ParseClusterConfig(fileName, &clusterConfig)
	if err != nil {
		return nil, err
	}
	return &clusterConfig, nil
}

// GetCloudStackManagementAPIEndpointHostname parses the CloudStackAvailabilityZone's ManagementApiEndpoint URL and returns the hostname.
func GetCloudStackManagementAPIEndpointHostname(az CloudStackAvailabilityZone) (string, error) {
	return getHostnameFromURL(az.ManagementApiEndpoint)
}

func getHostnameFromURL(rawurl string) (string, error) {
	url, err := url.Parse(rawurl)
	if err != nil {
		return "", fmt.Errorf("%s is not a valid url", rawurl)
	}
	return url.Hostname(), nil
}
