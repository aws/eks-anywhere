package decoder_test

import (
	_ "embed"
	"encoding/base64"
	"os"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

type testContext struct {
	oldCloudStackCloudConfigSecret   string
	isCloudStackCloudConfigSecretSet bool
}

func (tctx *testContext) backupContext() {
	tctx.oldCloudStackCloudConfigSecret, tctx.isCloudStackCloudConfigSecretSet = os.LookupEnv(decoder.EksacloudStackCloudConfigB64SecretKey)
}

func (tctx *testContext) restoreContext() {
	if tctx.isCloudStackCloudConfigSecretSet {
		os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, tctx.oldCloudStackCloudConfigSecret)
	}
}

func TestCloudStackConfigDecoderFromEnv(t *testing.T) {
	tests := []struct {
		name       string
		configFile string
		wantErr    bool
		wantConfig *decoder.CloudStackExecConfig
	}{
		{
			name:       "Valid config",
			configFile: "../testdata/cloudstack_config_valid.ini",
			wantErr:    false,
			wantConfig: &decoder.CloudStackExecConfig{
				Profiles: []decoder.CloudStackProfileConfig{
					{
						Name:          decoder.CloudStackGlobalAZ,
						ApiKey:        "test-key1",
						SecretKey:     "test-secret1",
						ManagementUrl: "http://127.16.0.1:8080/client/api",
						VerifySsl:     "false",
						Timeout:       "",
					},
				},
			},
		},
		{
			name:       "Multiple profiles config",
			configFile: "../testdata/cloudstack_config_multiple_profiles.ini",
			wantErr:    false,
			wantConfig: &decoder.CloudStackExecConfig{
				Profiles: []decoder.CloudStackProfileConfig{
					{
						Name:          decoder.CloudStackGlobalAZ,
						ApiKey:        "test-key1",
						SecretKey:     "test-secret1",
						ManagementUrl: "http://127.16.0.1:8080/client/api",
						VerifySsl:     "false",
					},
					{
						Name:          "instance2",
						ApiKey:        "test-key2",
						SecretKey:     "test-secret2",
						ManagementUrl: "http://127.16.0.2:8080/client/api",
						VerifySsl:     "true",
						Timeout:       "",
					},
				},
			},
		},
		{
			name:       "Missing apikey",
			configFile: "../testdata/cloudstack_config_missing_apikey.ini",
			wantErr:    true,
		},
		{
			name:       "Missing secretkey",
			configFile: "../testdata/cloudstack_config_missing_secretkey.ini",
			wantErr:    true,
		},
		{
			name:       "Missing apiurl",
			configFile: "../testdata/cloudstack_config_missing_apiurl.ini",
			wantErr:    true,
		},
		{
			name:       "Missing verifyssl",
			configFile: "../testdata/cloudstack_config_missing_verifyssl.ini",
			wantErr:    false,
			wantConfig: &decoder.CloudStackExecConfig{
				Profiles: []decoder.CloudStackProfileConfig{
					{
						Name:          decoder.CloudStackGlobalAZ,
						ApiKey:        "test-key1",
						SecretKey:     "test-secret1",
						ManagementUrl: "http://127.16.0.1:8080/client/api",
						VerifySsl:     "true",
						Timeout:       "",
					},
				},
			},
		},
		{
			name:       "Invalid INI format",
			configFile: "../testdata/cloudstack_config_invalid_format.ini",
			wantErr:    true,
		},
		{
			name:       "Invalid veryfyssl value",
			configFile: "../testdata/cloudstack_config_invalid_verifyssl.ini",
			wantErr:    true,
		},
		{
			name:       "No sections",
			configFile: "../testdata/cloudstack_config_no_sections.ini",
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			g := NewWithT(t)
			configString := test.ReadFile(t, tc.configFile)
			encodedConfig := base64.StdEncoding.EncodeToString([]byte(configString))
			tt.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, encodedConfig)

			gotConfig, err := decoder.ParseCloudStackCredsFromEnv()
			if tc.wantErr {
				g.Expect(err).NotTo(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				if !reflect.DeepEqual(tc.wantConfig, gotConfig) {
					t.Errorf("%v got = %v, want %v", tc.name, gotConfig, tc.wantConfig)
				}
			}
		})
	}
}

func TestCloudStackConfigDecoderFromSecrets(t *testing.T) {
	tests := []struct {
		name       string
		secrets    []apiv1.Secret
		wantErr    bool
		wantConfig *decoder.CloudStackExecConfig
	}{
		{
			name: "Valid config",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data: map[string][]byte{
						decoder.APIKeyKey:    []byte("test-key1"),
						decoder.APIUrlKey:    []byte("http://127.16.0.1:8080/client/api"),
						decoder.SecretKeyKey: []byte("test-secret1"),
						decoder.VerifySslKey: []byte("false"),
					},
				},
			},
			wantErr: false,
			wantConfig: &decoder.CloudStackExecConfig{
				Profiles: []decoder.CloudStackProfileConfig{
					{
						Name:          decoder.CloudStackGlobalAZ,
						ApiKey:        "test-key1",
						SecretKey:     "test-secret1",
						ManagementUrl: "http://127.16.0.1:8080/client/api",
						VerifySsl:     "false",
						Timeout:       "",
					},
				},
			},
		},
		{
			name: "Empty config",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data:       map[string][]byte{},
				},
			},
			wantErr:    true,
			wantConfig: nil,
		},
		{
			name: "Missing apikey",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data: map[string][]byte{
						decoder.APIUrlKey:    []byte("http://127.16.0.1:8080/client/api"),
						decoder.SecretKeyKey: []byte("test-secret1"),
						decoder.VerifySslKey: []byte("false"),
					},
				},
			},
			wantErr:    true,
			wantConfig: nil,
		},
		{
			name: "Missing api url",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data: map[string][]byte{
						decoder.APIKeyKey:    []byte("test-key1"),
						decoder.SecretKeyKey: []byte("test-secret1"),
						decoder.VerifySslKey: []byte("false"),
					},
				},
			},
			wantErr:    true,
			wantConfig: nil,
		},
		{
			name: "Missing secret key",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data: map[string][]byte{
						decoder.APIKeyKey:    []byte("test-key1"),
						decoder.APIUrlKey:    []byte("http://127.16.0.1:8080/client/api"),
						decoder.VerifySslKey: []byte("false"),
					},
				},
			},
			wantErr:    true,
			wantConfig: nil,
		},
		{
			name: "Missing verify ssl",
			secrets: []apiv1.Secret{
				{
					ObjectMeta: v1.ObjectMeta{Name: "global"},
					Data: map[string][]byte{
						decoder.APIKeyKey:    []byte("test-key1"),
						decoder.SecretKeyKey: []byte("test-secret1"),
						decoder.APIUrlKey:    []byte("http://127.16.0.1:8080/client/api"),
					},
				},
			},
			wantErr: false,
			wantConfig: &decoder.CloudStackExecConfig{
				Profiles: []decoder.CloudStackProfileConfig{
					{
						Name:          decoder.CloudStackGlobalAZ,
						ApiKey:        "test-key1",
						SecretKey:     "test-secret1",
						ManagementUrl: "http://127.16.0.1:8080/client/api",
						VerifySsl:     "true",
						Timeout:       "",
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(tt *testing.T) {
			g := NewWithT(t)
			gotConfig, err := decoder.ParseCloudStackCredsFromSecrets(tc.secrets)
			if tc.wantErr {
				g.Expect(err).NotTo(BeNil())
			} else {
				g.Expect(err).To(BeNil())
				if !reflect.DeepEqual(tc.wantConfig, gotConfig) {
					t.Errorf("%v got = %v, want %v", tc.name, gotConfig, tc.wantConfig)
				}
			}
		})
	}
}

func TestCloudStackConfigDecoderInvalidEncoding(t *testing.T) {
	g := NewWithT(t)
	t.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, "xxx")

	_, err := decoder.ParseCloudStackCredsFromEnv()
	g.Expect(err).NotTo(BeNil())
}

func TestCloudStackConfigDecoderNoEnvVariable(t *testing.T) {
	var tctx testContext
	tctx.backupContext()
	os.Clearenv()

	g := NewWithT(t)

	_, err := decoder.ParseCloudStackCredsFromEnv()
	g.Expect(err).NotTo(BeNil())
	tctx.restoreContext()
}
