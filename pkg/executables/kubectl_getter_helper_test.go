package executables_test

import (
	"bytes"
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/aws/eks-anywhere/internal/test"
)

type (
	getter            func(*kubectlGetterTest) (client.Object, error)
	kubectlGetterTest struct {
		*kubectlTest
		resourceType, name, namespace string
		json                          string
		getter                        getter
		want                          client.Object
	}
)

func newKubectlGetterTest(t *testing.T) *kubectlGetterTest {
	return &kubectlGetterTest{
		kubectlTest: newKubectlTest(t),
		name:        "name",
		namespace:   "my-ns",
	}
}

func (tt *kubectlGetterTest) withResourceType(r string) *kubectlGetterTest {
	tt.resourceType = r
	return tt
}

func (tt *kubectlGetterTest) withoutNamespace() *kubectlGetterTest {
	tt.namespace = ""
	return tt
}

func (tt *kubectlGetterTest) withJson(j string) *kubectlGetterTest {
	tt.json = j
	return tt
}

func (tt *kubectlGetterTest) withJsonFromFile(file string) *kubectlGetterTest {
	return tt.withJson(test.ReadFile(tt.t, file))
}

func (tt *kubectlGetterTest) withGetter(g getter) *kubectlGetterTest {
	tt.getter = g
	return tt
}

func (tt *kubectlGetterTest) andWant(o client.Object) *kubectlGetterTest {
	tt.want = o
	return tt
}

func (tt *kubectlGetterTest) testSuccess() {
	tt.WithT.THelper()

	argsWithoutName := []interface{}{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, tt.resourceType}
	if tt.namespace != "" {
		argsWithoutName = append(argsWithoutName, "--namespace", tt.namespace)
	}
	args := append(argsWithoutName, tt.name)

	tt.e.EXPECT().Execute(
		tt.ctx, args...,
	).Return(*bytes.NewBufferString(tt.json), nil)

	got, err := tt.getter(tt)
	tt.Expect(err).To(Not(HaveOccurred()), "Getter for %s should succeed", tt.resourceType)
	tt.Expect(got).To(Equal(tt.want), "Getter for %s should return correct object", tt.resourceType)
}

func (tt *kubectlGetterTest) testError() {
	tt.WithT.THelper()

	argsWithoutName := []interface{}{"get", "--ignore-not-found", "-o", "json", "--kubeconfig", tt.cluster.KubeconfigFile, tt.resourceType}
	if tt.namespace != "" {
		argsWithoutName = append(argsWithoutName, "--namespace", tt.namespace)
	}
	args := append(argsWithoutName, tt.name)

	tt.e.EXPECT().Execute(
		tt.ctx, args...,
	).Return(bytes.Buffer{}, errors.New("error in get"))

	_, err := tt.getter(tt)
	tt.Expect(err).To(MatchError(ContainSubstring("error in get")), "Getter for %s should fail", tt.resourceType)
}
