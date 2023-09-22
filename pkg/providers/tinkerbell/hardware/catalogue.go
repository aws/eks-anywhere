package hardware

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"

	tinkv1alpha1 "github.com/tinkerbell/tink/pkg/apis/core/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	yamlutil "k8s.io/apimachinery/pkg/util/yaml"
	"sigs.k8s.io/yaml"

	eksav1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1"
	rufiov1alpha1 "github.com/aws/eks-anywhere/pkg/api/v1alpha1/thirdparty/tinkerbell/rufio"
	"github.com/aws/eks-anywhere/pkg/templater"
)

// Indexer provides indexing behavior for objects.
type Indexer interface {
	// Lookup retrieves objects associated with the index => value pair.
	Lookup(index, value string) ([]interface{}, error)
	// Insert inserts v int the index.
	Insert(v interface{}) error
	// IndexField associated index with fn such that Lookup may be used to retrieve objects.
	IndexField(index string, fn KeyExtractorFunc)
	// Remove deletes v from the index.
	Remove(v interface{}) error
}

// Catalogue represents a catalogue of Tinkerbell hardware manifests to be used with Tinkerbells
// Kubefied back-end.
type Catalogue struct {
	hardware      []*tinkv1alpha1.Hardware
	hardwareIndex Indexer

	bmcs     []*rufiov1alpha1.Machine
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
		bmcIndex:      NewFieldIndexer(&rufiov1alpha1.Machine{}),
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

// ParseYAMLCatalogue parses a YAML document, r, that represents a set of Kubernetes manifests.
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
		case "Machine":
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
	var bmc rufiov1alpha1.Machine
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

// MarshalCatalogue marshals c into YAML that can be submitted to a Kubernetes cluster.
func MarshalCatalogue(c *Catalogue) ([]byte, error) {
	var marshallables []eksav1alpha1.Marshallable
	for _, hw := range c.AllHardware() {
		marshallables = append(marshallables, hw)
	}
	for _, bmc := range c.AllBMCs() {
		marshallables = append(marshallables, bmc)
	}
	for _, secret := range c.AllSecrets() {
		marshallables = append(marshallables, secret)
	}
	resources := make([][]byte, 0, len(marshallables))
	for _, marshallable := range marshallables {
		resource, err := yaml.Marshal(marshallable)
		if err != nil {
			return nil, fmt.Errorf("failed marshalling resource for hardware spec: %v", err)
		}
		resources = append(resources, resource)
	}
	return templater.AppendYamlResources(resources...), nil
}

// NewMachineCatalogueWriter creates a MachineWriter instance that writes Machine instances to
// catalogue including its Machine and Secret data.
func NewMachineCatalogueWriter(catalogue *Catalogue) MachineWriter {
	return MultiMachineWriter(
		NewHardwareCatalogueWriter(catalogue),
		NewBMCCatalogueWriter(catalogue),
		NewSecretCatalogueWriter(catalogue),
	)
}
