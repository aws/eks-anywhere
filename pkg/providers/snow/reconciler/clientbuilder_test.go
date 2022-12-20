package reconciler_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/providers/snow/reconciler"
)

const caBundle = `-----BEGIN CERTIFICATE-----
MIIDXjCCAkagAwIBAgIIb5m0RljJCMEwDQYJKoZIhvcNAQENBQAwODE2MDQGA1UE
AwwtSklELTIwNjg0MzQyMDAwMi0xOTItMTY4LTEtMjM1LTIyLTAxLTA2LTIyLTA0
MB4XDTIxMDExMTIyMDc1OFoXDTI1MTIxNjIyMDc1OFowODE2MDQGA1UEAwwtSklE
LTIwNjg0MzQyMDAwMi0xOTItMTY4LTEtMjM1LTIyLTAxLTA2LTIyLTA0MIIBIjAN
BgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAmOTQDBfBtPcVDFg/a59dk+rYrPRU
f5zl7JgFAEw1n82SkbNm4srwloj8pCuD1nJAlN+3LKoiby9jU8ZqoQKqppJaK1QK
dv27JYNlWorG9r6KrFkiETn2cxuAwcRBvq4UF76WdNr7zFjI108byPp9Pd0mxKiQ
6WVaxcKX9AEcarB/GfidHO95Aay6tiBU1SQvBJro3L1/UFu5STSpZai9zx+VkWTJ
D0JXh7eLF4yL0N1oU0hX2CGDxDz4VlJmBOvbnRuwsOruRMtUFRUy59cPzr//4fjd
4S7AYbeOVPwEP7q19NZ6+P7E71jTq1rz8RhAnW/JcbTKS0KqgBUPz0U4qQIDAQAB
o2wwajAMBgNVHRMEBTADAQH/MB0GA1UdDgQWBBQTaZzL2goqq7/MbJEfNRuzbwih
kTA7BgNVHREENDAyhjBJRDpKSUQtMjA2ODQzNDIwMDAyLTE5Mi0xNjgtMS0yMzUt
MjItMDEtMDYtMjItMDQwDQYJKoZIhvcNAQENBQADggEBAEzel+UsphUx49EVAyWB
PzSzoE7X62fg/b4gU7ifFHpWpYpAPsbapz9/Tywc4TGRItfctXYZsjchJKiutGU2
zX4rt1NSHkx72iMl3obQ2jQmTD8f9LyCqya+QM4CA74kk6v2ng1EiwMYvQlTvWY4
FEWv21yNRs2yiRuHWjRYH4TF54cCoDQGpFpsOFi0L4V/yo1XuimSLx2vvKZ0lCNt
KxC1oCgCxxNkOa/6iLk6qVANoX5KIVsataVhvGK+9mwWn8+dnMFneMiWd/jvi+dh
eywldVELBWRKELDdBc9Xb4i5BETF6dUlmvpWgpOXXO3uJlIRGZCVFLsgQ511oMxM
rEA=
-----END CERTIFICATE-----
`

const validCredentials = `[1.2.3.4]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow

[1.2.3.5]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow
`

const invalidCredentialsIniFormat = `1.2.3.5
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow
`

const invalidCredentialsMissingAccessKey = `[1.2.3.5]
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
region = snow
`

const invalidCredentialsMissingSecretKey = `[1.2.3.5]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
region = snow
`

const invalidCredentialsMissingRegion = `[1.2.3.5]
aws_access_key_id = ABCDEFGHIJKLMNOPQR2T
aws_secret_access_key = AfSD7sYz/TBZtzkReBl6PuuISzJ2WtNkeePw+nNzJ
`

func TestBuildSnowAwsClientMap(t *testing.T) {
	tests := []struct {
		name          string
		secretContent string
		secret        *corev1.Secret
		wantErr       string
	}{
		{
			name:    "valid",
			secret:  testSecret(validCredentials),
			wantErr: "",
		},
		{
			name:    "invalid ini format",
			secret:  testSecret(invalidCredentialsIniFormat),
			wantErr: "loading values from credentials: key-value delimiter not found: 1.2.3.5",
		},
		{
			name:    "missing access key",
			secret:  testSecret(invalidCredentialsMissingAccessKey),
			wantErr: "parsing configuration for 1.2.3.5: unable to set aws_access_key_id",
		},
		{
			name:    "missing secret key",
			secret:  testSecret(invalidCredentialsMissingSecretKey),
			wantErr: "parsing configuration for 1.2.3.5: unable to set aws_secret_access_key",
		},
		{
			name:    "missing region",
			secret:  testSecret(invalidCredentialsMissingRegion),
			wantErr: "parsing configuration for 1.2.3.5: unable to set region",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			ctx := context.Background()

			objs := []runtime.Object{tt.secret}

			cb := fake.NewClientBuilder()
			cl := cb.WithRuntimeObjects(objs...).Build()
			clientBuilder := reconciler.NewAwsClientBuilder(cl)

			_, err := clientBuilder.Get(ctx)
			if tt.wantErr == "" {
				g.Expect(err).To(Succeed())
			} else {
				g.Expect(err).To(MatchError(ContainSubstring(tt.wantErr)))
			}
		})
	}
}

func TestBuildSnowAwsClientMapNonexistentSecret(t *testing.T) {
	g := NewWithT(t)
	ctx := context.Background()

	cb := fake.NewClientBuilder()
	cl := cb.Build()
	clientBuilder := reconciler.NewAwsClientBuilder(cl)

	_, err := clientBuilder.Get(ctx)
	g.Expect(err).To(MatchError(ContainSubstring("getting snow credentials: secrets \"capas-manager-bootstrap-credentials\" not found")))
}

func testSecret(creds string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.CapasSystemNamespace,
			Name:      reconciler.BoostrapSecretName,
		},
		Data: map[string][]byte{
			"credentials": []byte(creds),
			"ca-bundle":   []byte(caBundle),
		},
	}
}
