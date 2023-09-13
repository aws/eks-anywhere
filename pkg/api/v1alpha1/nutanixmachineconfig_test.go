package v1alpha1_test

import (
	"fmt"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/cluster"
)

func TestGetNutanixMachineConfigsValidConfig(t *testing.T) {
	expectedMachineConfig := &v1alpha1.NutanixMachineConfig{
		TypeMeta: metav1.TypeMeta{
			Kind:       v1alpha1.NutanixMachineConfigKind,
			APIVersion: v1alpha1.SchemeBuilder.GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eksa-unit-test",
			Namespace: "default",
		},
		Spec: v1alpha1.NutanixMachineConfigSpec{
			SystemDiskSize: resource.MustParse("40Gi"),
			MemorySize:     resource.MustParse("8Gi"),
			VCPUSockets:    4,
			VCPUsPerSocket: 1,
			OSFamily:       v1alpha1.Ubuntu,
			Image: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("prism-image"),
			},
			Cluster: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("prism-element"),
			},
			Subnet: v1alpha1.NutanixResourceIdentifier{
				Type: v1alpha1.NutanixIdentifierName,
				Name: ptr.String("prism-subnet"),
			},
			Users: []v1alpha1.UserConfiguration{{
				Name:              "mySshUsername",
				SshAuthorizedKeys: []string{"mySshAuthorizedKey"},
			}},
		},
	}
	const machineConfName = "eksa-unit-test"

	tests := []struct {
		name        string
		fileName    string
		machineConf map[string]*v1alpha1.NutanixMachineConfig
		assertions  func(t *testing.T, machineConf *v1alpha1.NutanixMachineConfig)
	}{
		{
			name:     "valid-cluster-setters-getters",
			fileName: "testdata/nutanix/valid-cluster.yaml",
			machineConf: map[string]*v1alpha1.NutanixMachineConfig{
				machineConfName: expectedMachineConfig,
			},
			assertions: func(t *testing.T, machineConf *v1alpha1.NutanixMachineConfig) {
				assert.False(t, machineConf.IsReconcilePaused())
				machineConf.PauseReconcile()
				assert.True(t, machineConf.IsReconcilePaused())

				assert.False(t, machineConf.IsEtcd())
				machineConf.SetEtcd()
				assert.True(t, machineConf.IsEtcd())

				assert.False(t, machineConf.IsManaged())
				machineConf.Annotations = nil
				machineConf.SetManagedBy(machineConfName)
				assert.True(t, machineConf.IsManaged())

				assert.False(t, machineConf.IsControlPlane())
				machineConf.SetControlPlane()
				assert.True(t, machineConf.IsControlPlane())

				assert.Equal(t, v1alpha1.Ubuntu, machineConf.OSFamily())
				assert.Equal(t, "default", machineConf.GetNamespace())
				assert.Equal(t, machineConfName, machineConf.GetName())
			},
		},
		{
			name:     "valid-cluster-marshal",
			fileName: "testdata/nutanix/valid-cluster.yaml",
			machineConf: map[string]*v1alpha1.NutanixMachineConfig{
				machineConfName: expectedMachineConfig,
			},
			assertions: func(t *testing.T, machineConf *v1alpha1.NutanixMachineConfig) {
				m := machineConf.Marshallable()
				require.NotNil(t, m)
				y, err := yaml.Marshal(m)
				assert.NoError(t, err)
				assert.NotNil(t, y)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := cluster.ParseConfigFromFile(test.fileName)
			if err != nil {
				t.Fatal(err)
			}
			machineConfigs := config.NutanixMachineConfigs

			if test.machineConf != nil {
				require.NoError(t, err)
				require.NotNil(t, machineConfigs)
				assert.Equal(t, test.machineConf, machineConfigs)
				test.assertions(t, machineConfigs[machineConfName])
			} else {
				assert.Error(t, err)
			}
		})
	}
}

func TestNewNutanixMachineConfigGenerate(t *testing.T) {
	machineConf := v1alpha1.NewNutanixMachineConfigGenerate("eksa-unit-test", func(config *v1alpha1.NutanixMachineConfigGenerate) {
		config.Spec.MemorySize = resource.MustParse("16Gi")
	})
	require.NotNil(t, machineConf)
	assert.Equal(t, "eksa-unit-test", machineConf.Name())
	assert.Equal(t, v1alpha1.NutanixMachineConfigKind, machineConf.Kind())
	assert.Equal(t, v1alpha1.SchemeBuilder.GroupVersion.String(), machineConf.APIVersion())
	assert.Equal(t, resource.MustParse("16Gi"), machineConf.Spec.MemorySize)
}

func TestNutanixMachineConfigDefaults(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		validate func(t *testing.T, nutanixMachineConfig *v1alpha1.NutanixMachineConfig) error
	}{
		{
			name:     "machineconfig-with-no-users",
			fileName: "testdata/nutanix/machineconfig-with-no-users.yaml",
			validate: func(t *testing.T, nutanixMachineConfig *v1alpha1.NutanixMachineConfig) error {
				if len(nutanixMachineConfig.Spec.Users) <= 0 {
					return fmt.Errorf("default user was not added")
				}
				return nil
			},
		},
		{
			name:     "machineconfig-with-no-user-name",
			fileName: "testdata/nutanix/machineconfig-with-no-user-name.yaml",
			validate: func(t *testing.T, nutanixMachineConfig *v1alpha1.NutanixMachineConfig) error {
				if len(nutanixMachineConfig.Spec.Users[0].Name) <= 0 {
					return fmt.Errorf("default user name was not added")
				}
				return nil
			},
		},
		{
			name:     "machineconfig-with-no-osfamily",
			fileName: "testdata/nutanix/machineconfig-with-no-osfamily.yaml",
			validate: func(t *testing.T, nutanixMachineConfig *v1alpha1.NutanixMachineConfig) error {
				if nutanixMachineConfig.Spec.OSFamily != v1alpha1.DefaultOSFamily {
					return fmt.Errorf("v1alpha1.Ubuntu OS family was not set")
				}
				return nil
			},
		},
		{
			name:     "machineconfig-with-no-ssh-key",
			fileName: "testdata/nutanix/machineconfig-with-no-ssh-key.yaml",
			validate: func(t *testing.T, nutanixMachineConfig *v1alpha1.NutanixMachineConfig) error {
				if len(nutanixMachineConfig.Spec.Users[0].SshAuthorizedKeys) <= 0 {
					return fmt.Errorf("default ssh key was not added")
				}
				if nutanixMachineConfig.Spec.Users[0].SshAuthorizedKeys[0] == "" {
					return nil
				}
				return nil
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := cluster.ParseConfigFromFile(test.fileName)
			if err != nil {
				t.Fatal(err)
			}
			machineConfigs := config.NutanixMachineConfigs
			nutanixMachineConfig := machineConfigs["eksa-unit-test"]
			if nutanixMachineConfig == nil {
				t.Fatalf("Invalid yaml found")
			}
			nutanixMachineConfig.SetDefaults()
			err = test.validate(t, nutanixMachineConfig)
			if err != nil {
				t.Fatalf("validate failed with error :%s", err)
			}
		})
	}
}

func TestValidateNutanixMachineConfig(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		expectedErr string
	}{
		{
			name:        "invalid-machineconfig-addtional-categories-key",
			fileName:    "testdata/nutanix/invalid-machineconfig-addtional-categories-key.yaml",
			expectedErr: "NutanixMachineConfig: missing category key",
		},
		{
			name:        "invalid-machineconfig-addtional-categories-value",
			fileName:    "testdata/nutanix/invalid-machineconfig-addtional-categories-value.yaml",
			expectedErr: "NutanixMachineConfig: missing category value",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config, err := cluster.ParseConfigFromFile(test.fileName)
			if err != nil {
				t.Fatal(err)
			}
			machineConfigs := config.NutanixMachineConfigs

			nutanixMachineConfig := machineConfigs["eksa-unit-test"]
			if nutanixMachineConfig == nil {
				t.Fatalf("Invalid yaml found")
			}
			err = nutanixMachineConfig.Validate()
			if err == nil {
				t.Fatalf("validate should have failed")
			}
			assert.Contains(t, err.Error(), test.expectedErr, "expected error", test.expectedErr, "got error", err)
		})
	}
}
