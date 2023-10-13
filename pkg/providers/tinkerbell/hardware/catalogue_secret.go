package hardware

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/cluster-api/cmd/clusterctl/api/v1alpha3"

	"github.com/aws/eks-anywhere/pkg/constants"
)

// IndexSecret indexes Secret instances on index by extracfting the key using fn.
func (c *Catalogue) IndexSecret(index string, fn KeyExtractorFunc) {
	c.secretIndex.IndexField(index, fn)
}

// InsertSecret inserts Secrets into the catalogue. If any indexes exist, the Secret is indexed.
func (c *Catalogue) InsertSecret(secret *corev1.Secret) error {
	if err := c.secretIndex.Insert(secret); err != nil {
		return err
	}
	c.secrets = append(c.secrets, secret)
	return nil
}

// AllSecrets retrieves a copy of the catalogued Secret instances.
func (c *Catalogue) AllSecrets() []*corev1.Secret {
	secrets := make([]*corev1.Secret, len(c.secrets))
	copy(secrets, c.secrets)
	return secrets
}

// LookupSecret retrieves Secret instances on index with a key of key. Multiple Secrets _may_
// have the same key hence it can return multiple Secrets.
func (c *Catalogue) LookupSecret(index, key string) ([]*corev1.Secret, error) {
	untyped, err := c.secretIndex.Lookup(index, key)
	if err != nil {
		return nil, err
	}

	secrets := make([]*corev1.Secret, len(untyped))
	for i, v := range untyped {
		secrets[i] = v.(*corev1.Secret)
	}

	return secrets, nil
}

// TotalSecrets returns the total Secrets registered in the catalogue.
func (c *Catalogue) TotalSecrets() int {
	return len(c.secrets)
}

const SecretNameIndex = ".ObjectMeta.Name"

// WithSecretNameIndex creates a Secret index using SecretNameIndex on Secret.ObjectMeta.Name.
func WithSecretNameIndex() CatalogueOption {
	return func(c *Catalogue) {
		c.IndexSecret(SecretNameIndex, func(o interface{}) string {
			secret := o.(*corev1.Secret)
			return secret.ObjectMeta.Name
		})
	}
}

// SecretCatalogueWriter converts Machine instances to Tinkerbell BaseboardManagement and inserts them
// in a catalogue.
type SecretCatalogueWriter struct {
	catalogue *Catalogue
}

var _ MachineWriter = &SecretCatalogueWriter{}

// NewSecretCatalogueWriter creates a new SecretCatalogueWriter instance.
func NewSecretCatalogueWriter(catalogue *Catalogue) *SecretCatalogueWriter {
	return &SecretCatalogueWriter{catalogue: catalogue}
}

// Write converts m to a Tinkerbell BaseboardManagement and inserts it into w's Catalogue.
func (w *SecretCatalogueWriter) Write(m Machine) error {
	if m.HasBMC() {
		for _, s := range baseboardManagementSecretFromMachine(m) {
			if err := w.catalogue.InsertSecret(s); err != nil {
				return err
			}
		}
	}
	return nil
}

func baseboardManagementSecretFromMachine(m Machine) []*corev1.Secret {
	var s []*corev1.Secret

	if m.BMCOptions != nil && m.BMCOptions.RPC.ConsumerURL != "" {
		for idx, secret := range m.BMCOptions.RPC.HMAC.Secrets {
			s = append(s, &corev1.Secret{
				TypeMeta: newSecretTypeMeta(),
				ObjectMeta: v1.ObjectMeta{
					Name:      fmt.Sprintf("%v-%v", formatBMCSecretRef(m), idx),
					Namespace: constants.EksaSystemNamespace,
					Labels: map[string]string{
						v1alpha3.ClusterctlMoveLabel: "true",
					},
				},
				Type: "Opaque",
				Data: map[string][]byte{
					"secret": []byte(secret),
				},
			})
		}
	} else {
		s = append(s, &corev1.Secret{
			TypeMeta: newSecretTypeMeta(),
			ObjectMeta: v1.ObjectMeta{
				Name:      formatBMCSecretRef(m),
				Namespace: constants.EksaSystemNamespace,
				Labels: map[string]string{
					v1alpha3.ClusterctlMoveLabel: "true",
				},
			},
			Type: "kubernetes.io/basic-auth",
			Data: map[string][]byte{
				"username": []byte(m.BMCUsername),
				"password": []byte(m.BMCPassword),
			},
		})
	}
	return s
}
