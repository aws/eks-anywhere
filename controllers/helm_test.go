package controllers_test

import (
	"reflect"
	"testing"

	"github.com/aws/eks-anywhere/controllers"
	"github.com/aws/eks-anywhere/pkg/dependencies"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestNewHelmFactory(t *testing.T) {
	type args struct {
		client                client.Client
		dependencyHelmFactory *dependencies.HelmFactory
	}
	tests := []struct {
		name string
		args args
		want *controllers.HelmFactory
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := controllers.NewHelmFactory(tt.args.client, tt.args.dependencyHelmFactory); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHelmFactory() = %v, want %v", got, tt.want)
			}
		})
	}
}
