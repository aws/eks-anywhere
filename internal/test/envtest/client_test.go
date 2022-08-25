package envtest_test

import (
	"context"
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/aws/eks-anywhere/internal/test/envtest"
)

type notFailT struct {
	*testing.T
	failed       bool
	panicMessage string
}

func (n *notFailT) Fatal(args ...interface{}) {
	n.Logf("Expected failure: %s", fmt.Sprint(args...))
	n.failed = true
	panic(n.panicMessage)
}

func newNotFailT(t *testing.T) *notFailT {
	return &notFailT{
		T:            t,
		panicMessage: "failed in notFailT",
	}
}

func expectToFailTest(t *testing.T, f func(t testing.TB)) {
	t.Helper()
	testT := newNotFailT(t)
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(string); !ok || s != testT.panicMessage {
				panic(r)
			}
		}
	}()

	f(testT)
	t.Fatal("Expected to fail test but didn't")
}

func TestCreateObjs(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	ctx := context.Background()
	secret := &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "s",
			Namespace: "eksa-system",
		},
	}
	cm := &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "ConfigMap",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cm",
			Namespace: "eksa-system",
		},
	}

	envtest.CreateObjs(ctx, t, client, secret, cm)
}

func TestCreateObjsErrorGet(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	ctx := context.Background()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "s",
			Namespace: "eksa-system",
		},
	}

	expectToFailTest(t, func(tb testing.TB) {
		envtest.CreateObjs(ctx, tb, client, secret)
	})
}

func TestCreateObjsErrorCreate(t *testing.T) {
	client := fake.NewClientBuilder().Build()
	ctx := context.Background()
	secret := &corev1.Secret{}

	expectToFailTest(t, func(tb testing.TB) {
		envtest.CreateObjs(ctx, tb, client, secret)
	})
}
