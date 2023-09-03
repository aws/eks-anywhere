package yaml

import (
	"io"

	"sigs.k8s.io/yaml"
)

// K8sEncoder leverages the Kubernetes YAML package (sigs.k8s.io/yaml) to provide an Encoder data
// structure that is friendlier to the io package.
type K8sEncoder struct {
	out io.Writer
}

// NewK8sEncoder creates a K8sEncoder instance that writes to out.
func NewK8sEncoder(out io.Writer) K8sEncoder {
	return K8sEncoder{out}
}

// Encode marshals v into YAML and writes it to e's output stream.
func (e K8sEncoder) Encode(v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}

	_, err = e.out.Write(b)
	return err
}
