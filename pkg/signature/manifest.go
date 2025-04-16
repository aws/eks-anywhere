// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package signature

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"github.com/golang-jwt/jwt/v5"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ValidateSignature validates the signature annotation of the bundles object using KMS public key.
func ValidateSignature(bundle *anywherev1alpha1.Bundles, pubKey string) (valid bool, err error) {
	bundleSig := bundle.Annotations[constants.SignatureAnnotation]
	if bundleSig == "" {
		return false, errors.New("missing bundle signature annotation")
	}

	digest, _, err := getBundleDigest(bundle)
	if err != nil {
		return false, err
	}

	sig, err := base64.StdEncoding.DecodeString(bundleSig)
	if err != nil {
		return false, fmt.Errorf("bundle signature in metadata isn't base64 encoded: %w", err)
	}

	pubkey, err := parsePublicKey(pubKey)
	if err != nil {
		return false, err
	}

	return ecdsa.VerifyASN1(pubkey, digest[:], sig), nil
}

// ValidateEKSDistroManifestSignature validates the signature annotation of the bundles object using KMS public key.
func ValidateEKSDistroManifestSignature(release *eksdv1alpha1.Release, signature, pubKey, releaseChannel string) (valid bool, err error) {
	if signature == "" {
		return false, fmt.Errorf("missing %s eks distro manifest signature annotation", releaseChannel)
	}

	digest, _, err := getEKSDistroReleaseDigest(release)
	if err != nil {
		return false, err
	}

	sig, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false, fmt.Errorf("eks distro manifest signature in metadata for %s release channel isn't base64 encoded: %w", releaseChannel, err)
	}

	pubkey, err := parsePublicKey(pubKey)
	if err != nil {
		return false, err
	}

	return ecdsa.VerifyASN1(pubkey, digest[:], sig), nil
}

// getEksdDigest computes the SHA256 digest for an EKS Distro release object.
// It follows similar steps as getBundleDigest() for Bundles by marshalling the object,
// converting it to JSON, filtering out undesired fields, and then computing the hash.
func getEKSDistroReleaseDigest(release *eksdv1alpha1.Release) ([32]byte, []byte, error) {
	var zero [32]byte

	// Marshal the eks-distro release object to YAML.
	yamlBytes, err := yaml.Marshal(release)
	if err != nil {
		return zero, nil, fmt.Errorf("marshalling eks distro release to YAML: %v", err)
	}

	// Convert the YAML to JSON for easier gojq processing.
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("converting eks distro release YAML to JSON: %v", err)
	}

	// Build and execute the gojq filter that deletes excluded fields.
	filtered, err := filterExcludes(jsonBytes, constants.EKSDistroExcludes)
	if err != nil {
		return zero, nil, fmt.Errorf("filtering excluded fields: %v", err)
	}

	// Compute the SHA256 digest of the filtered JSON.
	digest := sha256.Sum256(filtered)
	return digest, filtered, nil
}

// getBundleDigest converts the Bundles manifest to JSON, excludes certain fields, then
// computes the SHA256 hash of the filtered manifest. It returns the digest and
// the final bytes used to produce that digest.
func getBundleDigest(bundle *anywherev1alpha1.Bundles) ([32]byte, []byte, error) {
	var zero [32]byte

	// Marshal Bundles object to YAML
	yamlBytes, err := yaml.Marshal(bundle)
	if err != nil {
		return zero, nil, fmt.Errorf("marshalling bundle to YAML: %w", err)
	}

	// Convert YAML to JSON for easier gojq processing
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("converting YAML to JSON: %w", err)
	}

	// Build and execute the gojq filter that deletes excluded fields
	filtered, err := filterExcludes(jsonBytes, constants.Excludes)
	if err != nil {
		return zero, nil, fmt.Errorf("filtering excluded fields: %w", err)
	}

	// Compute the SHA256 sum of the filtered JSON
	digest := sha256.Sum256(filtered)

	return digest, filtered, nil
}

// filterExcludes applies the default and user-specified excludes to the JSON
// representation of the Bundles object using gojq.
func filterExcludes(jsonBytes []byte, excludes string) ([]byte, error) {
	// Decode the base64-encoded excludes
	exclBytes, err := base64.StdEncoding.DecodeString(excludes)
	if err != nil {
		return nil, fmt.Errorf("decoding Excludes: %v", err)
	}
	// Convert them into slice of strings
	userExcludes := strings.Split(string(exclBytes), "\n")

	// Combine AlwaysExcluded with userExcludes
	allExcludes := constants.AlwaysExcludedFields
	if userExcludes[0] != "" {
		allExcludes = append(allExcludes, userExcludes...)
	}

	// Build the argument to the gojq template
	var tmplBuf bytes.Buffer
	gojqTemplate := template.Must(template.New("gojq_query").Funcs(
		template.FuncMap{
			"StringsJoin": strings.Join,
			"Escape": func(in string) string {
				// We need to escape '.' for certain gojq path usage
				// to avoid ambiguities in the path expressions.
				return strings.ReplaceAll(in, ".", "\\\\.")
			},
		},
	).Parse(`
	del({{ StringsJoin .Excludes ", " }})
	`))
	if err := gojqTemplate.Execute(&tmplBuf, map[string]interface{}{
		"Excludes": allExcludes,
	}); err != nil {
		return nil, fmt.Errorf("executing gojq template: %v", err)
	}

	// Parse the final gojq query
	query, err := gojq.Parse(tmplBuf.String())
	if err != nil {
		return nil, fmt.Errorf("gojq parse error: %v", err)
	}

	// Unmarshal the JSON into a generic interface so gojq can operate
	var input interface{}
	if err := json.Unmarshal(jsonBytes, &input); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON: %v", err)
	}

	// Run the query
	iter := query.Run(input)
	finalVal, ok := iter.Next()
	if !ok {
		return nil, errors.New("gojq produced no result")
	}
	if errVal, ok := finalVal.(error); ok {
		return nil, fmt.Errorf("gojq execution error: %v", errVal)
	}

	// Marshal the filtered result back to JSON
	filtered, err := json.Marshal(finalVal)
	if err != nil {
		return nil, fmt.Errorf("marshalling final result to JSON: %v", err)
	}
	return filtered, nil
}

func parsePublicKey(key string) (*ecdsa.PublicKey, error) {
	pubdecoded, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, fmt.Errorf("decoding the public key as string: %w", err)
	}

	pubparsed, err := x509.ParsePKIXPublicKey(pubdecoded)
	if err != nil {
		return nil, fmt.Errorf("parsing the public key (not PKIX): %w", err)
	}

	pubkey, ok := pubparsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("parsing the public key (not ECDSA): %T", pubparsed)
	}
	return pubkey, nil
}

// ParseLicense parses licenseKey jwt token using the public key and returns token fields.
func ParseLicense(licenseToken string, key string) (*jwt.Token, error) {
	tokenKey, err := parsePublicKey(key)
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse(licenseToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, fmt.Errorf("signing method not supported: %v", t.Header["alg"])
		}
		return tokenKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parsing licenseToken: %w", err)
	}

	if !token.Valid {
		return nil, errors.New("licenseToken is not valid")
	}

	return token, nil
}
