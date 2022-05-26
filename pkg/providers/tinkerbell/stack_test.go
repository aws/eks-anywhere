package tinkerbell

import (
	"os"
	"testing"

	"github.com/aws/eks-anywhere/pkg/constants"
	"github.com/aws/eks-anywhere/release/api/v1alpha1"
	"gopkg.in/yaml.v3"
)

func TestStackInstallWithBoots(t *testing.T) {
	bundle := v1alpha1.TinkerbellStackBundle{
		Boots: v1alpha1.TinkerbellServiceBundle{
			Image: v1alpha1.Image{URI: "boots:latest"},
		},
		Hegel: v1alpha1.TinkerbellServiceBundle{
			Image: v1alpha1.Image{URI: "hegel:latest"},
		},
		Tink: v1alpha1.TinkBundle{
			TinkController: v1alpha1.Image{URI: "tink-controller-manager:latest"},
			TinkServer:     v1alpha1.Image{URI: "tink-server:latest"},
			TinkWorker:     v1alpha1.Image{URI: "tink-worker:latest"},
		},
	}
	stack := newStack(bundle, nil, nil, nil, "1.2.3.4").
		withNamespace(constants.EksaSystemNamespace, false).
		withBoots().
		withHegel().
		withTinkController().
		withTinkServer()

	t.Log(stack.getValues())
}

func TestStackInstallWithoutBoots(t *testing.T) {
	bundle := v1alpha1.TinkerbellStackBundle{
		Boots: v1alpha1.TinkerbellServiceBundle{
			Image: v1alpha1.Image{URI: "boots:latest"},
		},
		Hegel: v1alpha1.TinkerbellServiceBundle{
			Image: v1alpha1.Image{URI: "hegel:latest"},
		},
		Tink: v1alpha1.TinkBundle{
			TinkController: v1alpha1.Image{URI: "tink-controller-manager:latest"},
			TinkServer:     v1alpha1.Image{URI: "tink-server:latest"},
			TinkWorker:     v1alpha1.Image{URI: "tink-worker:latest"},
		},
	}
	stack := newStack(bundle, nil, nil, nil, "1.2.3.4").
		withNamespace(constants.EksaSystemNamespace, false).
		withBoots().
		withHegel().
		withTinkController().
		withTinkServer()

	out, _ := yaml.Marshal(stack.values)
	os.WriteFile("test.yaml", out, os.ModePerm)

	// t.Log(string(out))
}
