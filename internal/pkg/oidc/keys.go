package oidc

import (
	"encoding/json"
	"fmt"

	jose "gopkg.in/square/go-jose.v2"
)

type keyResponse struct {
	Keys []jose.JSONWebKey `json:"keys"`
}

func GetKeyID(keysBytes []byte) (string, error) {
	keys, err := parseKeys(keysBytes)
	if err != nil {
		return "", err
	}

	if len(keys.Keys) != 1 {
		return "", fmt.Errorf("keys array in keys.json should have size 1 but has size %d", len(keys.Keys))
	}

	return keys.Keys[0].KeyID, nil
}

func parseKeys(bytes []byte) (*keyResponse, error) {
	keys := &keyResponse{}
	err := json.Unmarshal(bytes, keys)
	if err != nil {
		return nil, fmt.Errorf("parsing keys.json to get kid: %v", err)
	}

	return keys, nil
}
