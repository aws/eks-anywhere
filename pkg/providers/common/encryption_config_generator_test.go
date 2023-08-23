package common_test

import (
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	. "github.com/aws/eks-anywhere/pkg/providers/common"
)

const expectedEncryptionConfig = "testdata/expected_encryption_config.yaml"

func TestGenerateKMSEncryptionConfigurationEmpty(t *testing.T) {
	g := NewWithT(t)
	tests := []struct {
		name   string
		config *[]v1alpha1.EtcdEncryption
		want   string
	}{
		{
			name:   "nil config",
			config: nil,
			want:   "",
		},
		{
			name:   "empty config",
			config: &[]v1alpha1.EtcdEncryption{},
			want:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(*testing.T) {
			got, err := GenerateKMSEncryptionConfiguration(tt.config)
			g.Expect(err).ToNot(HaveOccurred())
			g.Expect(got).To(Equal(tt.want))
		})
	}
}

func TestGenerateEncryptionConfiguration(t *testing.T) {
	encryptionConf := &[]v1alpha1.EtcdEncryption{
		{
			Providers: []v1alpha1.EtcdEncryptionProvider{
				{
					KMS: &v1alpha1.KMS{
						Name:                "config1",
						SocketListenAddress: "unix:///var/run/kmsplugin/socket1-new.sock",
						CacheSize:           v1alpha1.DefaultKMSCacheSize,
						Timeout:             &v1alpha1.DefaultKMSTimeout,
					},
				},
				{
					KMS: &v1alpha1.KMS{
						Name:                "config2",
						SocketListenAddress: "unix:///var/run/kmsplugin/socket1-old.sock",
						CacheSize:           v1alpha1.DefaultKMSCacheSize,
						Timeout:             &v1alpha1.DefaultKMSTimeout,
					},
				},
			},
			Resources: []string{
				"secrets",
				"crd1.anywhere.eks.amazonsaws.com",
			},
		},
		{
			Providers: []v1alpha1.EtcdEncryptionProvider{
				{
					KMS: &v1alpha1.KMS{
						Name:                "config3",
						SocketListenAddress: "unix:///var/run/kmsplugin/socket2-new.sock",
						CacheSize:           v1alpha1.DefaultKMSCacheSize,
						Timeout:             &v1alpha1.DefaultKMSTimeout,
					},
				},
				{
					KMS: &v1alpha1.KMS{
						Name:                "config4",
						SocketListenAddress: "unix:///var/run/kmsplugin/socket2-old.sock",
						CacheSize:           v1alpha1.DefaultKMSCacheSize,
						Timeout:             &v1alpha1.DefaultKMSTimeout,
					},
				},
			},
			Resources: []string{
				"configmaps",
				"crd2.anywhere.eks.amazonsaws.com",
			},
		},
	}

	conf, err := GenerateKMSEncryptionConfiguration(encryptionConf)
	if err != nil {
		t.Fatal(err)
	}
	test.AssertContentToFile(t, conf, expectedEncryptionConfig)
}
