package nutanix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	anywherev1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestNewClientCache(t *testing.T) {
	cc := NewClientCache()
	dcConf := &anywherev1.NutanixDatacenterConfig{}
	err := yaml.Unmarshal([]byte(nutanixDatacenterConfigSpecWithTrustBundle), dcConf)
	require.NoError(t, err)
	t.Setenv(constants.EksaNutanixUsernameKey, "admin")
	t.Setenv(constants.EksaNutanixPasswordKey, "password")
	c, err := cc.GetNutanixClient(dcConf, GetCredsFromEnv())
	assert.NoError(t, err)
	assert.NotNil(t, c)
}
