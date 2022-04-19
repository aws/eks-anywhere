package oidc

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"

	jose "gopkg.in/square/go-jose.v2"
)

type MinimalProvider struct {
	Discovery, Keys, PrivateKey []byte
	KeyID                       string
}

type discoveryResponse struct {
	Issuer                           string   `json:"issuer"`
	JwksUri                          string   `json:"jwks_uri"`
	AuthorizationEndpoint            string   `json:"authorization_endpoint"`
	ResponseTypesSupported           []string `json:"response_types_supported"`
	SubjectTypesSupported            []string `json:"subject_types_supported"`
	IdTokenSigningAlgValuesSupported []string `json:"id_token_signing_alg_values_supported"`
	ClaimsSupported                  []string `json:"claims_supported"`
}

func GenerateMinimalProvider(issuerURL string) (*MinimalProvider, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, fmt.Errorf("generating rsa key for OIDC: %v", err)
	}

	pubKey := &privateKey.PublicKey
	pubKeyEncoded, err := marshalPubKey(pubKey)
	if err != nil {
		return nil, err
	}

	// Set the key id to the sha1 of the pubkey
	sha := sha1.New()
	_, err = sha.Write(pubKeyEncoded)
	if err != nil {
		return nil, fmt.Errorf("generating sha1 checksum for pubkey: %v", err)
	}

	kid := fmt.Sprintf("%x", sha.Sum(nil))
	keys := []jose.JSONWebKey{
		{
			Key:       pubKey,
			KeyID:     kid,
			Algorithm: string(jose.RS256),
			Use:       "sig",
		},
	}

	keysJson, err := json.MarshalIndent(keyResponse{Keys: keys}, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("marshalling keys.json for OIDC: %v", err)
	}

	d := &discoveryResponse{
		Issuer:                           issuerURL,
		JwksUri:                          fmt.Sprintf("%s/keys.json", issuerURL),
		AuthorizationEndpoint:            "urn:kubernetes:programmatic_authorization",
		ResponseTypesSupported:           []string{"id_token"},
		SubjectTypesSupported:            []string{"public"},
		IdTokenSigningAlgValuesSupported: []string{"RS256"},
		ClaimsSupported:                  []string{"sub", "iss"},
	}

	discoveryJson, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return nil, fmt.Errorf("marshalling json discovery for OIDC: %v", err)
	}

	marshalledPrivateKey, err := marshalPrivateKey(privateKey)
	if err != nil {
		return nil, err
	}

	return &MinimalProvider{
		Discovery:  discoveryJson,
		Keys:       keysJson,
		PrivateKey: marshalledPrivateKey,
		KeyID:      kid,
	}, nil
}

func marshalPubKey(pubKey *rsa.PublicKey) ([]byte, error) {
	pubKeyBytes, err := x509.MarshalPKIXPublicKey(pubKey)
	if err != nil {
		return nil, fmt.Errorf("marshalling pub key for oidc provider: %v", err)
	}

	pubKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubKeyBytes,
	}

	buffer := &bytes.Buffer{}
	err = pem.Encode(buffer, pubKeyBlock)
	if err != nil {
		return nil, fmt.Errorf("encoding public key for oidc provider: %v", err)
	}

	return buffer.Bytes(), nil
}

func marshalPrivateKey(k *rsa.PrivateKey) ([]byte, error) {
	pubKeyBytes := x509.MarshalPKCS1PrivateKey(k)

	keyBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: pubKeyBytes,
	}

	buffer := &bytes.Buffer{}
	err := pem.Encode(buffer, keyBlock)
	if err != nil {
		return nil, fmt.Errorf("encoding private key for oidc provider: %v", err)
	}

	return buffer.Bytes(), nil
}
