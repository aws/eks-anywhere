package yamlutil_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/yamlutil"
)

type yamlHolder struct {
	configMap *corev1.ConfigMap
	secret    *corev1.Secret
}

func (h *yamlHolder) BuildFromParsed(l yamlutil.ObjectLookup) error {
	processConfigMap(h, l)
	processSecret(h, l)

	return nil
}

func processConfigMap(h *yamlHolder, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "ConfigMap" {
			h.configMap = obj.(*corev1.ConfigMap)
		}
	}
}

func processSecret(h *yamlHolder, lookup yamlutil.ObjectLookup) {
	for _, obj := range lookup {
		if obj.GetObjectKind().GroupVersionKind().Kind == "Secret" {
			h.secret = obj.(*corev1.Secret)
		}
	}
}

func TestParserParse(t *testing.T) {
	g := NewWithT(t)
	yaml := `
apiVersion: v1
data:
  Corefile: "d"
kind: ConfigMap
metadata:
  name: aws-iam-authenticator 
  namespace: kube-system
  uid: 4aa825d5-4334-4ce0-a754-0d3a3cceaefd
---
apiVersion: v1
kind: Secret
metadata:
  name: aws-iam-authenticator 
  namespace: kube-system
data:
  password: QWRtaW4=
  username: YWRtaW4=
`
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(
		parser.RegisterMappings(
			yamlutil.NewMapping("Secret", func() yamlutil.APIObject {
				return &corev1.Secret{}
			}),
			yamlutil.NewMapping("ConfigMap", func() yamlutil.APIObject {
				return &corev1.ConfigMap{}
			}),
		),
	).To(Succeed())

	holder := &yamlHolder{}

	g.Expect(parser.Parse([]byte(yaml), holder)).To(Succeed())
	g.Expect(holder).NotTo(BeNil())
	g.Expect(holder.configMap.Data).To(HaveKeyWithValue("Corefile", "d"))
	g.Expect(holder.secret.Data["username"]).To(Equal([]byte("admin")))
}

type reader struct {
	read int
	err  error
}

func (r reader) Read(p []byte) (n int, err error) {
	return r.read, r.err
}

func TestParserReadReaderError(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	holder := &yamlHolder{}
	r := reader{
		err: errors.New("failed from fake reader"),
	}
	g.Expect(parser.Read(r, holder)).To(MatchError(ContainSubstring("failed from fake reader")))
}

func TestParserParseEmptyContent(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	holder := &yamlHolder{}
	g.Expect(parser.Parse([]byte("---"), holder)).NotTo(HaveOccurred())
}

func TestParserParseInvalidKubernetesYaml(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	holder := &yamlHolder{}

	g.Expect(parser.Parse([]byte("1}"), holder)).To(MatchError(ContainSubstring(
		"invalid yaml kubernetes object",
	)))
}

func TestParserParseInvalidRegisterObjectYaml(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	holder := &yamlHolder{}
	g.Expect(
		parser.RegisterMappings(
			yamlutil.NewMapping("Secret", func() yamlutil.APIObject {
				return &corev1.Secret{}
			}),
		),
	).To(Succeed())

	g.Expect(parser.Parse([]byte("kind: Secret\ndata: 111"), holder)).To(MatchError(ContainSubstring(
		"invalid yaml for Secret",
	)))
}

func TestParserParseUnregisteredObject(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	holder := &yamlHolder{}
	g.Expect(parser.Parse([]byte("kind: Secret"), holder)).NotTo(HaveOccurred())
}

func TestParserRegisterMappingDuplicateError(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(parser.RegisterMapping("Secret", func() yamlutil.APIObject {
		return &corev1.Secret{}
	})).To(Succeed())
	g.Expect(parser.RegisterMapping("Secret", func() yamlutil.APIObject {
		return &corev1.ConfigMap{}
	})).To(MatchError(ContainSubstring("mapping for api object Secret already registered")))
}

func TestMappingToAPIObjectMapping(t *testing.T) {
	g := NewWithT(t)
	mapping := yamlutil.NewMapping("Secret", func() *corev1.Secret {
		return &corev1.Secret{}
	})
	apiObjectMapping := mapping.ToAPIObjectMapping()
	g.Expect(apiObjectMapping.Kind).To(Equal("Secret"))
	secret := apiObjectMapping.New()
	g.Expect(secret).To(BeAssignableToTypeOf(&corev1.Secret{}))
}

func TestParserRegisterMappingsError(t *testing.T) {
	g := NewWithT(t)
	parser := yamlutil.NewParser(test.NewNullLogger())
	g.Expect(
		parser.RegisterMappings(
			yamlutil.NewMapping("Secret", func() yamlutil.APIObject {
				return &corev1.Secret{}
			}),
			yamlutil.NewMapping("Secret", func() yamlutil.APIObject {
				return &corev1.ConfigMap{}
			}),
		),
	).To(MatchError(ContainSubstring("mapping for api object Secret already registered")))
}
