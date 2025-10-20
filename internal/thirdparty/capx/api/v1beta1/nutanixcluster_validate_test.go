/*
Copyright 2024 Nutanix

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1_test

import (
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apiextensions-apiserver/pkg/apiserver/validation"
	runtime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/kustomize/api/krusty"
	"sigs.k8s.io/kustomize/kyaml/filesys"
	"sigs.k8s.io/kustomize/kyaml/resid"
)

func TestNutanixClusterOpenAPIValidationPCAddress(t *testing.T) {
	t.Parallel()
	g := NewGomegaWithT(t)

	k := krusty.MakeKustomizer(krusty.MakeDefaultOptions())
	kRes, err := k.Run(filesys.MakeFsOnDisk(), filepath.Join("..", "..", "config", "crd"))
	g.Expect(err).NotTo(HaveOccurred())
	ncCRDRes, err := kRes.GetById(resid.ResId{
		Gvk:       resid.Gvk{Kind: "CustomResourceDefinition", Version: "v1", Group: "apiextensions.k8s.io"},
		Name:      "nutanixclusters.infrastructure.cluster.x-k8s.io",
		Namespace: resid.TotallyNotANamespace,
	})
	g.Expect(err).NotTo(HaveOccurred())
	ncCRD := &apiextensionsv1.CustomResourceDefinition{}
	_, _, err = scheme.Codecs.UniversalDeserializer().Decode([]byte(ncCRDRes.MustYaml()), nil, ncCRD)
	g.Expect(err).NotTo(HaveOccurred())

	testCases := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{name: "invalid empty", value: "", wantErr: true},

		{name: "valid localhost", value: "localhost"},
		{name: "valid domain name", value: "example.com"},
		{name: "valid localhost ip", value: "127.0.0.1"},
		{name: "valid ipv4", value: "192.168.0.1"},
		{name: "valid ipv6", value: "2345:0425:2CA1:0000:0000:0567:5673:23b5"},
		{name: "valid ipv6 with leading zeros removed", value: "2345:425:2CA1:0000:0000:567:5673:23b5"},

		{name: "invalid with port (hostname)", value: "localhost:8080", wantErr: true},
		{name: "invalid with port (ipv4)", value: "127.0.0.1:8080", wantErr: true},
		{name: "invalid with scheme (hostname)", value: "https://localhost", wantErr: true},
		{name: "invalid with scheme (ipv4)", value: "https://127.0.0.1", wantErr: true},
		{name: "invalid with scheme (ipv6)", value: "https://2345:0425:2CA1:0000:0000:0567:5673:23b5", wantErr: true},
		{name: "invalid ipv6", value: "234H:0425:2CA1:0000:0000:0567:5673:23b5", wantErr: true},
	}

	crdScheme := runtime.NewScheme()
	g.Expect(apiextensionsv1.AddToScheme(crdScheme)).NotTo(HaveOccurred())

	for _, tc := range testCases {
		tc := tc // Capture range variable.
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			g := NewGomegaWithT(t)

			for _, v := range ncCRD.Spec.Versions {
				pcAddressProp := v.Schema.OpenAPIV3Schema.Properties["spec"].Properties["prismCentral"].Properties["address"]

				unversionedCRValidation := &apiextensions.CustomResourceValidation{}
				err := crdScheme.Convert(
					&apiextensionsv1.CustomResourceValidation{OpenAPIV3Schema: &pcAddressProp},
					unversionedCRValidation,
					nil,
				)
				g.Expect(err).NotTo(HaveOccurred())

				validator, _, err := validation.NewSchemaValidator(unversionedCRValidation)
				g.Expect(err).NotTo(HaveOccurred())

				if tc.wantErr {
					g.Expect(validator.Validate(tc.value).Errors).ToNot(BeEmpty())
				} else {
					g.Expect(validator.Validate(tc.value).Errors).To(BeEmpty())
				}
			}
		})
	}
}
