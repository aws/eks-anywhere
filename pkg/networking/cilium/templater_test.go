package cilium_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/cluster"
	"github.com/aws/eks-anywhere/pkg/networking/cilium"
	"github.com/aws/eks-anywhere/pkg/networking/cilium/mocks"
)

type templaterTest struct {
	*WithT
	ctx                     context.Context
	t                       *cilium.Templater
	h                       *mocks.MockHelm
	manifest                []byte
	uri, version, namespace string
	spec, currentSpec       *cluster.Spec
}

func newtemplaterTest(t *testing.T) *templaterTest {
	ctrl := gomock.NewController(t)
	h := mocks.NewMockHelm(ctrl)
	return &templaterTest{
		WithT:     NewWithT(t),
		ctx:       context.Background(),
		h:         h,
		t:         cilium.NewTemplater(h),
		manifest:  []byte("manifestContent"),
		uri:       "placeholder",
		version:   "v1.9.11-eksa.1",
		namespace: "kube-system",
		currentSpec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium.Version = "v1.9.10-eksa.1"
			s.VersionsBundle.Cilium.Cilium.URI = "public.ecr.aws/isovalent/cilium:v1.9.10-eksa.1"
			s.VersionsBundle.Cilium.Operator.URI = "public.ecr.aws/isovalent/operator:v1.9.10-eksa.1"
		}),
		spec: test.NewClusterSpec(func(s *cluster.Spec) {
			s.VersionsBundle.Cilium.Version = "v1.9.11-eksa.1"
			s.VersionsBundle.Cilium.Cilium.URI = "public.ecr.aws/isovalent/cilium:v1.9.11-eksa.1"
			s.VersionsBundle.Cilium.Operator.URI = "public.ecr.aws/isovalent/operator:v1.9.11-eksa.1"
		}),
	}
}

func (t *templaterTest) expectHelmTemplateWith(wantValues gomock.Matcher) *gomock.Call {
	return t.h.EXPECT().Template(t.ctx, t.uri, t.version, t.namespace, wantValues)
}

func eqMap(m map[string]interface{}) gomock.Matcher {
	return &mapMatcher{m: m}
}

// mapMacher implements a gomock matcher for maps
// transforms any map or struct into a map[string]interface{} and uses DeepEqual to compare
type mapMatcher struct {
	m map[string]interface{}
}

func (e *mapMatcher) Matches(x interface{}) bool {
	xJson, err := json.Marshal(x)
	if err != nil {
		return false
	}
	xMap := &map[string]interface{}{}
	err = json.Unmarshal(xJson, xMap)
	if err != nil {
		return false
	}

	return reflect.DeepEqual(e.m, *xMap)
}

func (e *mapMatcher) String() string {
	return fmt.Sprintf("matches map %v", e.m)
}

func TestTemplaterGenerateUpgradePreflightManifestSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
			"enabled": false,
		},
		"preflight": map[string]interface{}{
			"enabled": true,
		},
		"agent": false,
	}

	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(eqMap(wantValues)).Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateUpgradePreflightManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateUpgradePreflightManifest() should return right manifest")
}

func TestTemplaterGenerateUpgradePreflightManifestError(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(gomock.Any()).Return(nil, errors.New("error from helm")) // Using any because we only want to test the returned error

	_, err := tt.t.GenerateUpgradePreflightManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateUpgradePreflightManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error from helm")))
}

func TestTemplaterGenerateManifestSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
	}

	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(eqMap(wantValues)).Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateManifest(tt.ctx, tt.spec)).To(Equal(tt.manifest), "templater.GenerateManifest() should return right manifest")
}

func TestTemplaterGenerateManifestError(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(gomock.Any()).Return(nil, errors.New("error from helm")) // Using any because we only want to test the returned error

	_, err := tt.t.GenerateManifest(tt.ctx, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error from helm")))
}

func TestTemplaterGenerateUpgradeManifestSuccess(t *testing.T) {
	wantValues := map[string]interface{}{
		"cni": map[string]interface{}{
			"chainingMode": "portmap",
		},
		"ipam": map[string]interface{}{
			"mode": "kubernetes",
		},
		"identityAllocationMode": "crd",
		"prometheus": map[string]interface{}{
			"enabled": true,
		},
		"rollOutCiliumPods": true,
		"tunnel":            "geneve",
		"image": map[string]interface{}{
			"repository": "public.ecr.aws/isovalent/cilium",
			"tag":        "v1.9.11-eksa.1",
		},
		"operator": map[string]interface{}{
			"image": map[string]interface{}{
				"repository": "public.ecr.aws/isovalent/operator",
				"tag":        "v1.9.11-eksa.1",
			},
			"prometheus": map[string]interface{}{
				"enabled": true,
			},
		},
		"upgradeCompatibility": "1.9",
	}

	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(eqMap(wantValues)).Return(tt.manifest, nil)

	tt.Expect(tt.t.GenerateUpgradeManifest(tt.ctx, tt.currentSpec, tt.spec)).To(Equal(tt.manifest), "templater.GenerateUpgradeManifest() should return right manifest")
}

func TestTemplaterGenerateUpgradeManifestError(t *testing.T) {
	tt := newtemplaterTest(t)
	tt.expectHelmTemplateWith(gomock.Any()).Return(nil, errors.New("error from helm")) // Using any because we only want to test the returned error

	_, err := tt.t.GenerateUpgradeManifest(tt.ctx, tt.currentSpec, tt.spec)
	tt.Expect(err).To(HaveOccurred(), "templater.GenerateUpgradeManifest() should fail")
	tt.Expect(err).To(MatchError(ContainSubstring("error from helm")))
}
