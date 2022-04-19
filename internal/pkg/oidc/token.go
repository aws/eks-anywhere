package oidc

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"time"

	jose "gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

type oidcTokenClaim struct {
	Issuer           string `json:"iss,omitempty"`
	Subject          string `json:"sub,omitempty"`
	kid              string
	keyFile          string
	Role             string   `json:"role,omitempty"`
	Email            string   `json:"email,omitempty"`
	Audience         []string `json:"aud,omitempty"`
	Groups           []string `json:"groups,omitempty"`
	ExpiresAt        int64    `json:"exp,omitempty"`
	IssuedAt         int64    `json:"iat,omitempty"`
	NotBefore        int64    `json:"nbf,omitempty"`
	KubernetesAccess string   `json:"kubernetesAccess,omitempty"`
}

type JWTOpt func(*oidcTokenClaim)

func NewJWT(issuerName, kid, keyFile string, opts ...JWTOpt) (string, error) {
	now := time.Now()
	o := &oidcTokenClaim{
		Issuer:    fmt.Sprintf("https://%s", issuerName),
		Audience:  []string{},
		ExpiresAt: now.Add(time.Hour * 24).Unix(),
		IssuedAt:  now.Unix(),
		NotBefore: now.Unix(),
		Groups:    []string{},
		kid:       kid,
		keyFile:   keyFile,
	}

	for _, opt := range opts {
		opt(o)
	}

	return o.generateToken()
}

func WithEmail(email string) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.Email = email
	}
}

func WithGroup(group string) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.Groups = append(o.Groups, group)
	}
}

func WithRole(role string) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.Role = role
	}
}

func WithKubernetesAccess(access bool) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.KubernetesAccess = fmt.Sprintf("%t", access)
	}
}

func WithAudience(audience string) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.Audience = append(o.Audience, audience)
	}
}

func WithSubject(subject string) JWTOpt {
	return func(o *oidcTokenClaim) {
		o.Subject = subject
	}
}

func (o *oidcTokenClaim) generateToken() (string, error) {
	keyContent, err := ioutil.ReadFile(o.keyFile)
	if err != nil {
		return "", fmt.Errorf("could not read key file: %v", err)
	}
	block, _ := pem.Decode(keyContent)
	if block == nil {
		return "", fmt.Errorf("decoding PEM file %s", o.keyFile)
	}
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parsing key content of %s: %v", o.keyFile, err)
	}

	jwkKey := &jose.JSONWebKey{
		Algorithm: string(jose.RS256),
		Key:       privKey,
		KeyID:     o.kid,
		Use:       "sig",
	}

	signer, err := jose.NewSigner(
		jose.SigningKey{
			Algorithm: jose.RS256,
			Key:       jwkKey,
		},
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create signer: %v", err)
	}

	token, err := jwt.Signed(signer).Claims(o).CompactSerialize()
	if err != nil {
		return "", fmt.Errorf("signing token: %v", err)
	}

	return token, nil
}
