package hardware

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	pbnjv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/pbnj/api/v1alpha1"
	tinkv1alpha1 "github.com/tinkerbell/cluster-api-provider-tinkerbell/tink/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"
)

const Provisioning = "provisioning"

// Catalogue represents a catalogue of Tinkerbell hardware manifests to be used with Tinkerbells
// Kubefied back-end.
type Catalogue struct {
	hardware      []*tinkv1alpha1.Hardware
	hardwareIndex Indexer

	bmcs     []*pbnjv1alpha1.BMC
	bmcIndex Indexer

	secrets     []*corev1.Secret
	secretIndex Indexer
}

// CatalogueOption defines an option to be applied in Catalogue instantiation.
type CatalogueOption func(*Catalogue)

// NewCatalogue creates a new Catalogue instance.
func NewCatalogue(opts ...CatalogueOption) *Catalogue {
	catalogue := &Catalogue{
		hardwareIndex: NewFieldIndexer(&tinkv1alpha1.Hardware{}),
		bmcIndex:      NewFieldIndexer(&pbnjv1alpha1.BMC{}),
		secretIndex:   NewFieldIndexer(&corev1.Secret{}),
	}

	for _, opt := range opts {
		opt(catalogue)
	}

	return catalogue
}

// ParseYAMLCatalogueFromFile parses filename, a YAML document, using ParseYamlCatalogue.
func ParseYAMLCatalogueFromFile(catalogue *Catalogue, filename string) error {
	fh, err := os.Open(filename)
	if err != nil {
		return err
	}

	return ParseYAMLCatalogue(catalogue, fh)
}

// ParseCatalogue parses a YAML document, r, that represents a set of Kubernetes manifests.
// Manifests parsed include CAPT Hardware, PBnJ BMCs and associated Core API Secret.
func ParseYAMLCatalogue(catalogue *Catalogue, r io.Reader) error {
	document := yamlutil.NewYAMLReader(bufio.NewReader(r))
	for {
		manifest, err := document.Read()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}

		var resource unstructured.Unstructured
		if err = yaml.Unmarshal(manifest, &resource); err != nil {
			return err
		}

		switch resource.GetKind() {
		case "Hardware":
			if err := catalogueSerializedHardware(catalogue, manifest); err != nil {
				return err
			}
		case "BMC":
			if err := catalogueSerializedBMC(catalogue, manifest); err != nil {
				return err
			}
		case "Secret":
			if err := catalogueSerializedSecret(catalogue, manifest); err != nil {
				return err
			}
		}
	}
}

func catalogueSerializedHardware(catalogue *Catalogue, manifest []byte) error {
	var hardware tinkv1alpha1.Hardware
	if err := yaml.UnmarshalStrict(manifest, &hardware); err != nil {
		return fmt.Errorf("unable to parse hardware manifest: %v", err)
	}
	if err := catalogue.InsertHardware(&hardware); err != nil {
		return err
	}
	return nil
}

func catalogueSerializedBMC(catalogue *Catalogue, manifest []byte) error {
	var bmc pbnjv1alpha1.BMC
	if err := yaml.UnmarshalStrict(manifest, &bmc); err != nil {
		return fmt.Errorf("unable to parse bmc manifest: %v", err)
	}
	if err := catalogue.InsertBMC(&bmc); err != nil {
		return err
	}
	return nil
}

func catalogueSerializedSecret(catalogue *Catalogue, manifest []byte) error {
	var secret corev1.Secret
	if err := yaml.UnmarshalStrict(manifest, &secret); err != nil {
		return fmt.Errorf("unable to parse secret manifest: %v", err)
	}
	if err := catalogue.InsertSecret(&secret); err != nil {
		return err
	}
	return nil
}
