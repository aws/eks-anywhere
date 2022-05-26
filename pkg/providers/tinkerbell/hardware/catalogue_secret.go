package hardware

import (
	corev1 "k8s.io/api/core/v1"
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
