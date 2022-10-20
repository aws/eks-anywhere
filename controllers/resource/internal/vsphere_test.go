package internal_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"

	"github.com/aws/eks-anywhere/controllers/resource/internal"
)

func getSecret() *corev1.Secret {
	return &corev1.Secret{
		Data: map[string][]byte{"username": []byte("username"), "password": []byte("password"), "usernameCSI": []byte("usernameCSI"), "passwordCSI": []byte("passwordCSI"), "usernameCP": []byte("usernameCP"), "passwordCP": []byte("passwordCP")},
	}
}

func TestGetVSphereCredValues(t *testing.T) {
	g := NewWithT(t)
	s := getSecret()
	values, _ := internal.GetVSphereCredValues(s)
	g.Expect(values).ToNot(BeNil())
}

func TestGetVSphereCredValuesError(t *testing.T) {
	for _, k := range []string{"username", "password"} {
		t.Run(k, func(t *testing.T) {
			s := getSecret()
			delete(s.Data, k)
			g := NewWithT(t)
			_, err := internal.GetVSphereCredValues(s)
			target := fmt.Sprintf("unable to retrieve %s from secret", k)
			g.Expect(err.Error()).To(BeEquivalentTo(target))
		})
	}
}

func TestGetVSphereCredValuesMissingPassword(t *testing.T) {
	tests := []struct {
		defaultField  string
		secretField   string
		templateField string
	}{
		{
			defaultField:  "username",
			secretField:   "usernameCSI",
			templateField: "eksaCSIUsername",
		},
		{
			defaultField:  "username",
			secretField:   "usernameCP",
			templateField: "eksaCloudProviderUsername",
		},
		{
			defaultField:  "password",
			secretField:   "passwordCSI",
			templateField: "eksaCSIPassword",
		},
		{
			defaultField:  "password",
			secretField:   "passwordCP",
			templateField: "eksaCloudProviderPassword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.secretField, func(t *testing.T) {
			s := getSecret()
			delete(s.Data, tt.secretField)
			g := NewWithT(t)
			values, _ := internal.GetVSphereCredValues(s)
			target := s.Data[tt.defaultField]
			res := values[tt.templateField]
			g.Expect(res).To(BeEquivalentTo(target))
		})
	}
}
