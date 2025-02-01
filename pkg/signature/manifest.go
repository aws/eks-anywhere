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

	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

// ValidateSignature validates the signature annotation of the bundles object using KMS public key.
func ValidateSignature(bundle *anywherev1alpha1.Bundles, pubKey string) (valid bool, err error) {
	bundleSig := bundle.Annotations[constants.SignatureAnnotation]
	if bundleSig == "" {
		return false, errors.New("missing signature annotation")
	}

	digest, _, err := getDigest(bundle)
	if err != nil {
		return false, err
	}

	sig, err := base64.StdEncoding.DecodeString(bundleSig)
	if err != nil {
		return false, fmt.Errorf("signature in metadata isn't base64 encoded: %w", err)
	}

	pubdecoded, err := base64.StdEncoding.DecodeString(pubKey)
	if err != nil {
		return false, fmt.Errorf("decoding the public key as string: %w", err)
	}

	pubparsed, err := x509.ParsePKIXPublicKey(pubdecoded)
	if err != nil {
		return false, fmt.Errorf("parsing the public key (not PKIX): %w", err)
	}

	pubkey, ok := pubparsed.(*ecdsa.PublicKey)
	if !ok {
		return false, fmt.Errorf("parsing the public key (not ECDSA): %T", pubparsed)
	}

	return ecdsa.VerifyASN1(pubkey, digest[:], sig), nil
}

// getDigest converts the Bundles manifest to JSON, excludes certain fields, then
// computes the SHA256 hash of the filtered manifest. It returns the digest and
// the final bytes used to produce that digest.
func getDigest(bundle *anywherev1alpha1.Bundles) ([32]byte, []byte, error) {
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
	filtered, err := filterExcludes(jsonBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("filtering excluded fields: %w", err)
	}

	// Compute the SHA256 sum of the filtered JSON
	digest := sha256.Sum256(filtered)

	return digest, filtered, nil
}

// filterExcludes applies the default and user-specified excludes to the JSON
// representation of the Bundles object using gojq.
// This function has dependency on constants.AlwaysExcludedFields and constants.Excludes fields.
func filterExcludes(jsonBytes []byte) ([]byte, error) {
	// Decode the base64-encoded excludes
	exclBytes, err := base64.StdEncoding.DecodeString(constants.Excludes)
	if err != nil {
		return nil, fmt.Errorf("decoding Excludes: %w", err)
	}
	// Convert them into slice of strings
	userExcludes := strings.Split(string(exclBytes), "\n")

	// Combine AlwaysExcluded with userExcludes
	allExcludes := append(constants.AlwaysExcludedFields, userExcludes...)

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
		return nil, fmt.Errorf("executing gojq template: %w", err)
	}

	// Parse the final gojq query
	query, err := gojq.Parse(tmplBuf.String())
	if err != nil {
		return nil, fmt.Errorf("gojq parse error: %w", err)
	}

	// Unmarshal the JSON into a generic interface so gojq can operate
	var input interface{}
	if err := json.Unmarshal(jsonBytes, &input); err != nil {
		return nil, fmt.Errorf("unmarshalling JSON: %w", err)
	}

	// Run the query
	iter := query.Run(input)
	finalVal, ok := iter.Next()
	if !ok {
		return nil, errors.New("gojq produced no result")
	}
	if errVal, ok := finalVal.(error); ok {
		return nil, fmt.Errorf("gojq execution error: %w", errVal)
	}

	// Marshal the filtered result back to JSON
	filtered, err := json.Marshal(finalVal)
	if err != nil {
		return nil, fmt.Errorf("marshalling final result to JSON: %w", err)
	}
	return filtered, nil
}
