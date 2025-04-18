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
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"

	anywhereconstants "github.com/aws/eks-anywhere/pkg/constants"
	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/clients"
	"github.com/aws/eks-anywhere/release/cli/pkg/filereader"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
)

// GojqTemplate is used to build a gojq filter expression that deletes the desired fields.
var GojqTemplate = template.Must(template.New("gojq_query").Funcs(
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

// GetBundleSignature calls KMS and retrieves a signature, then base64-encodes it
// to store in the Bundles manifest annotation.
func GetBundleSignature(ctx context.Context, bundle *anywherev1alpha1.Bundles, key string) (string, error) {
	// Compute the digest from the Bundles manifest, excluding certain fields.
	digest, _, err := getBundleDigest(bundle)
	if err != nil {
		return "", fmt.Errorf("computing digest: %v", err)
	}

	// Create KMS Client for bundle manifest signing
	kmsClient, err := clients.CreateKMSClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating kms client: %v", err)
	}

	// The KMS Sign API requires the raw hash as the Message when MessageType is DIGEST.
	input := &kms.SignInput{
		KeyId:            &key,
		Message:          digest[:],
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
	}
	out, err := kmsClient.Sign(ctx, input)
	if err != nil {
		return "", fmt.Errorf("signing bundle with KMS Sign API: %v", err)
	}
	// Return the base64-encoded signature.
	return base64.StdEncoding.EncodeToString(out.Signature), nil
}

// GetEKSDistroManifestSignature calls KMS and retrieves a signature, then base64-encodes it
// to store in the Bundles manifest annotation.
func GetEKSDistroManifestSignature(ctx context.Context, bundle *anywherev1alpha1.Bundles, key, releaseUrl string) (string, error) {
	// Retrieve the eks-distro release from the release URL.
	eksdRelease, err := filereader.GetEksdRelease(releaseUrl)
	if err != nil {
		return "", fmt.Errorf("getting eks distro release from the %s eksd manifest release url: %v", releaseUrl, err)
	}

	// Compute the digest for the eks-distro release, excluding certain fields.
	digest, _, err := getEKSDistroReleaseDigest(eksdRelease)
	if err != nil {
		return "", fmt.Errorf("computing digest for eks distro manifest: %v", err)
	}

	// Create KMS Client for eks distro manifest signing
	kmsClient, err := clients.CreateKMSClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating kms client: %v", err)
	}

	// The KMS Sign API requires the raw hash as the Message when MessageType is DIGEST.
	input := &kms.SignInput{
		KeyId:            &key,
		Message:          digest[:],
		MessageType:      types.MessageTypeDigest,
		SigningAlgorithm: types.SigningAlgorithmSpecEcdsaSha256,
	}
	out, err := kmsClient.Sign(ctx, input)
	if err != nil {
		return "", fmt.Errorf("signing eks distro manifest with KMS Sign API: %v", err)
	}
	// Return the base64-encoded signature.
	return base64.StdEncoding.EncodeToString(out.Signature), nil
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
	filtered, err := filterExcludes(jsonBytes, anywhereconstants.EKSDistroExcludes)
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

	// Marshal Bundles object to YAML.
	yamlBytes, err := yaml.Marshal(bundle)
	if err != nil {
		return zero, nil, fmt.Errorf("marshalling bundle to YAML: %v", err)
	}

	// Convert YAML to JSON for easier gojq processing.
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("converting YAML to JSON: %v", err)
	}

	// Build and execute the gojq filter that deletes excluded fields.
	filtered, err := filterExcludes(jsonBytes, anywhereconstants.Excludes)
	if err != nil {
		return zero, nil, fmt.Errorf("filtering excluded fields: %v", err)
	}

	// Compute the SHA256 digest of the filtered JSON.
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
	allExcludes := anywhereconstants.AlwaysExcludedFields
	if userExcludes[0] != "" {
		allExcludes = append(allExcludes, userExcludes...)
	}

	// Build the argument to the gojq template
	var tmplBuf bytes.Buffer
	if err := GojqTemplate.Execute(&tmplBuf, map[string]interface{}{
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
