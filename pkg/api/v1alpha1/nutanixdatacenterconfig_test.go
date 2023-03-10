package v1alpha1_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/constants"
)

func TestGetNutanixDatacenterConfigInvalidConfig(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		expectedErr string
	}{
		{
			name:        "non-existent-file",
			fileName:    "testdata/nutanix/non-existent-file.yaml",
			expectedErr: "open testdata/nutanix/non-existent-file.yaml: no such file or directory",
		},
		{
			name:        "invalid-file",
			fileName:    "testdata/invalid_format.yaml",
			expectedErr: "unable to parse testdata/invalid_format.yaml",
		},
		{
			name:        "invalid-cluster-extraneous-field",
			fileName:    "testdata/nutanix/invalid-cluster.yaml",
			expectedErr: "unknown field \"idont\"",
		},
		{
			name:        "invalid-kind",
			fileName:    "testdata/nutanix/invalid-kind.yaml",
			expectedErr: "does not contain kind NutanixDatacenterConfig",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf, err := v1alpha1.GetNutanixDatacenterConfig(test.fileName)
			assert.Error(t, err)
			assert.Nil(t, conf)
			assert.Contains(t, err.Error(), test.expectedErr, "expected error", test.expectedErr, "got error", err)
		})
	}
}

func TestGetNutanixDatacenterConfigValidConfig(t *testing.T) {
	expectedDCConf := &v1alpha1.NutanixDatacenterConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.NutanixDatacenterKind,
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-unit-test",
			Namespace: constants.DefaultNamespace,
		},
		Spec: v1alpha1.NutanixDatacenterConfigSpec{
			Endpoint: "prism.nutanix.com",
			Port:     9440,
			CredentialRef: &v1alpha1.Ref{
				Name: "eksa-unit-test",
				Kind: constants.SecretKind,
			},
		},
	}

	tests := []struct {
		name       string
		fileName   string
		assertions func(*testing.T, *v1alpha1.NutanixDatacenterConfig)
	}{
		{
			name:     "valid-cluster",
			fileName: "testdata/nutanix/valid-cluster.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				assert.NoError(t, dcConf.Validate())
				assert.Equal(t, expectedDCConf, dcConf)
			},
		},
		{
			name:     "valid-cluster-extra-delimiter",
			fileName: "testdata/nutanix/valid-cluster-extra-delimiter.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				assert.NoError(t, dcConf.Validate())
			},
		},
		{
			name:     "valid-cluster-setters-getters",
			fileName: "testdata/nutanix/valid-cluster.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				assert.Equal(t, dcConf.ExpectedKind(), dcConf.Kind())

				assert.False(t, dcConf.IsReconcilePaused())
				dcConf.PauseReconcile()
				assert.True(t, dcConf.IsReconcilePaused())
				dcConf.ClearPauseAnnotation()
				assert.False(t, dcConf.IsReconcilePaused())
			},
		},
		{
			name:     "valid-cluster-marshal",
			fileName: "testdata/nutanix/valid-cluster.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				m := dcConf.Marshallable()
				require.NotNil(t, m)
				y, err := yaml.Marshal(m)
				assert.NoError(t, err)
				assert.NotNil(t, y)
			},
		},
		{
			name:     "datacenterconfig-valid-trust-bundle",
			fileName: "testdata/nutanix/datacenterconfig-valid-trustbundle.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				assert.NoError(t, dcConf.Validate())
			},
		},
		{
			name:     "datacenterconfig-invalid-trust-bundle",
			fileName: "testdata/nutanix/datacenterconfig-invalid-trustbundle.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NutanixDatacenterConfig additionalTrustBundle is not valid")
			},
		},
		{
			name:     "datacenterconfig-non-pem-trust-bundle",
			fileName: "testdata/nutanix/datacenterconfig-non-pem-trustbundle.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "could not find a PEM block in the certificate")
			},
		},
		{
			name:     "datacenterconfig-empty-endpoint",
			fileName: "testdata/nutanix/datacenterconfig-empty-endpoint.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NutanixDatacenterConfig endpoint is not set or is empty")
			},
		},
		{
			name:     "datacenterconfig-invalid-port",
			fileName: "testdata/nutanix/datacenterconfig-invalid-port.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NutanixDatacenterConfig port is not set or is empty")
			},
		},
		{
			name:     "datecenterconfig-credentialref-invalid-kind",
			fileName: "testdata/nutanix/invalid-credentialref-kind.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NutanixDatacenterConfig credentialRef Kind (ConfigMap) is not a secret")
			},
		},
		{
			name:     "datecenterconfig-credentialref-invalid-kind",
			fileName: "testdata/nutanix/empty-credentialref-name.yaml",
			assertions: func(t *testing.T, dcConf *v1alpha1.NutanixDatacenterConfig) {
				err := dcConf.Validate()
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "NutanixDatacenterConfig credentialRef name is not set or is empty")
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			conf, err := v1alpha1.GetNutanixDatacenterConfig(test.fileName)
			assert.NoError(t, err)
			require.NotNil(t, conf)
			test.assertions(t, conf)
		})
	}
}

func TestNewNutanixDatacenterConfigGenerate(t *testing.T) {
	dcConfGen := v1alpha1.NewNutanixDatacenterConfigGenerate("eksa-unit-test")
	require.NotNil(t, dcConfGen)
	assert.Equal(t, "eksa-unit-test", dcConfGen.Name())
	assert.Equal(t, v1alpha1.NutanixDatacenterKind, dcConfGen.Kind())
	assert.Equal(t, v1alpha1.SchemeBuilder.GroupVersion.String(), dcConfGen.APIVersion())
	assert.Equal(t, constants.NutanixCredentialsName, dcConfGen.Spec.CredentialRef.Name)
}

func TestNutanixDatacenterConfigSetDefaults(t *testing.T) {
	dcConf := &v1alpha1.NutanixDatacenterConfig{}
	dcConf.SetDefaults()
	assert.Equal(t, constants.NutanixCredentialsName, dcConf.Spec.CredentialRef.Name)
	assert.Equal(t, constants.SecretKind, dcConf.Spec.CredentialRef.Kind)
}
