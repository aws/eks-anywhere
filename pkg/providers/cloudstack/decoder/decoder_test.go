package decoder_test

import (
	_ "embed"
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/providers/cloudstack/decoder"
)

const (
	apiKey                     = "test-key"
	secretKey                  = "test-secret"
	apiUrl                     = "http://127.16.0.1:8080/client/api"
	verifySsl                  = "false"
	defaultVerifySsl           = "true"
	validCloudStackCloudConfig = "W0dsb2JhbF0KYXBpLWtleSA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAphcGktdXJsID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgPSBmYWxzZQo="
	missingApiKey              = "W0dsb2JhbF0Kc2VjcmV0LWtleSA9IHRlc3Qtc2VjcmV0CmFwaS11cmwgPSBodHRwOi8vMTI3LjE2LjAuMTo4MDgwL2NsaWVudC9hcGkKdmVyaWZ5LXNzbCA9IGZhbHNlCg=="
	missingSecretKey           = "W0dsb2JhbF0KYXBpLWtleSA9IHRlc3Qta2V5CmFwaS11cmwgPSBodHRwOi8vMTI3LjE2LjAuMTo4MDgwL2NsaWVudC9hcGkKdmVyaWZ5LXNzbCA9IGZhbHNlCg=="
	missingApiUrl              = "W0dsb2JhbF0KYXBpLWtleSA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAp2ZXJpZnktc3NsID0gZmFsc2UK"
	missingVerifySsl           = "W0dsb2JhbF0KYXBpLWtleSA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAphcGktdXJsID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCg=="
	invalidVerifySslValue      = "W0dsb2JhbF0KYXBpLWtleSA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAphcGktdXJsID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgPSBUVFRUVAo="
	missingGlobalSection       = "YXBpLWtleSA9IHRlc3Qta2V5CnNlY3JldC1rZXkgPSB0ZXN0LXNlY3JldAphcGktdXJsID0gaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgPSBmYWxzZQo="
	invalidINI                 = "W0dsb2JhbF0KYXBpLWtleSA7IHRlc3Qta2V5CnNlY3JldC1rZXkgOyB0ZXN0LXNlY3JldAphcGktdXJsIDsgaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgOyBmYWxzZQo="
	invalidEncoding            = "=====W0dsb2JhbF0KYXBpLWtleSA7IHRlc3Qta2V5CnNlY3JldC1rZXkgOyB0ZXN0LXNlY3JldAphcGktdXJsIDsgaHR0cDovLzEyNy4xNi4wLjE6ODA4MC9jbGllbnQvYXBpCnZlcmlmeS1zc2wgOyBmYWxzZQo======"
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

func TestValidConfigShouldSucceedtoParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, validCloudStackCloudConfig)
	execConfig, err := decoder.ParseCloudStackSecret()
	g.Expect(err).To(BeNil(), "An error occurred when parsing a valid secret")
	g.Expect(execConfig.ApiKey).To(Equal(apiKey))
	g.Expect(execConfig.SecretKey).To(Equal(secretKey))
	g.Expect(execConfig.ManagementUrl).To(Equal(apiUrl))
	g.Expect(execConfig.VerifySsl).To(Equal(verifySsl))

	tctx.restoreContext()
}

func TestMissingApiKeyShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, missingApiKey)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestMissingSecretKeyShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, missingSecretKey)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestMissingApiUrlShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, missingApiUrl)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestMissingVerifySslShouldSetDefaultValue(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, missingVerifySsl)
	execConfig, err := decoder.ParseCloudStackSecret()
	g.Expect(err).To(BeNil(), "An error occurred when parsing a valid secret")
	g.Expect(execConfig.ApiKey).To(Equal(apiKey))
	g.Expect(execConfig.SecretKey).To(Equal(secretKey))
	g.Expect(execConfig.ManagementUrl).To(Equal(apiUrl))
	g.Expect(execConfig.VerifySsl).To(Equal(defaultVerifySsl))

	tctx.restoreContext()
}

func TestInvalidVerifySslShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, invalidVerifySslValue)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestMissingGlobalSectionShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, missingGlobalSection)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestInvalidINIShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, invalidINI)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestMissingEnvVariableShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Unsetenv(decoder.EksacloudStackCloudConfigB64SecretKey)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}

func TestInvalidEncodingShouldFailToParse(t *testing.T) {
	var tctx testContext
	tctx.backupContext()

	g := NewWithT(t)
	os.Setenv(decoder.EksacloudStackCloudConfigB64SecretKey, invalidEncoding)
	_, err := decoder.ParseCloudStackSecret()
	g.Expect(err).ToNot(BeNil())

	tctx.restoreContext()
}
