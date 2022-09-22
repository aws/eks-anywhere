package yamlutil_test

import (
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
	parser := yamlutil.NewParser[yamlHolder](test.NewNullLogger())
	// yamlutil.RegisterMapping[*corev1.Secret](parser, "Secret")
	parser.RegisterMapping("ConfigMap", func() yamlutil.APIObject {
		return &corev1.ConfigMap{}
	})
	parser.RegisterMapping("Secret", func() yamlutil.APIObject {
		return &corev1.Secret{}
	})
	parser.RegisterProcessors(processConfigMap, processSecret)

	holder, err := parser.Parse([]byte(yaml))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(holder).NotTo(BeNil())
	g.Expect(holder.configMap.Data).To(HaveKeyWithValue("Corefile", "d"))
	g.Expect(holder.secret.Data["username"]).To(Equal([]byte("admin")))
}
