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

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/kms/types"
	"github.com/itchyny/gojq"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

const (
	// DefaultRegion used to create KMS client.
	DefaultRegion = "us-west-2"
	// KmsKey alias.
	KmsKey = "arn:aws:kms:us-west-2:857151390494:alias/signingEKSABundlesKey"

	// SignatureAnnotation applied to the bundle during bundle manifest signing.
	SignatureAnnotation = "anywhere.eks.amazonaws.com/signature"
	// ExcludesAnnotation applied to the bundle during bundle manifest signing.
	ExcludesAnnotation = "anywhere.eks.amazonaws.com/excludes"

	// Excludes is a base64-encoded, newline-delimited list of JSON/YAML paths to remove
	// from the Bundles manifest prior to computing the digest. You can add or remove
	// fields depending on your signing requirements.
	Excludes = "LnNwZWMudmVyc2lvbnNCdW5kbGVzW10uYm9vdHN0cmFwCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmJvdHRsZXJvY2tldEhvc3RDb250YWluZXJzCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNlcnRNYW5hZ2VyCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNpbGl1bQouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5jbG91ZFN0YWNrCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmNsdXN0ZXJBUEkKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uY29udHJvbFBsYW5lCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmRvY2tlcgouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5la3NhCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmV0Y2RhZG1Cb290c3RyYXAKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uZXRjZGFkbUNvbnRyb2xsZXIKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uZmx1eAouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS5oYXByb3h5Ci5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLmtpbmRuZXRkCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLm51dGFuaXgKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10ucGFja2FnZUNvbnRyb2xsZXIKLnNwZWMudmVyc2lvbnNCdW5kbGVzW10uc25vdwouc3BlYy52ZXJzaW9uc0J1bmRsZXNbXS50aW5rZXJiZWxsCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLnVwZ3JhZGVyCi5zcGVjLnZlcnNpb25zQnVuZGxlc1tdLnZTcGhlcmU="
)

// AlwaysExcluded are fields we always exclude from signature generation.
var AlwaysExcluded = []string{
	".status",
	".metadata.creationTimestamp",
	".metadata.annotations",
}

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
	// Compute the digest from the Bundles manifest after excluding certain fields.
	digest, _, err := getDigest(bundle)
	if err != nil {
		return "", fmt.Errorf("computing digest: %v", err)
	}

	// Create KMS Client for bundle manifest signing
	kmsClient, err := CreateKMSClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating kms client: %v", err)
	}

	// The KMS service expects you to send the raw hash in the `Message` field
	// when using `MessageType: DIGEST`.
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

	// Return the base64-encoded signature bytes.
	return base64.StdEncoding.EncodeToString(out.Signature), nil
}

// getDigest converts the Bundles manifest to JSON, excludes certain fields, then
// computes the SHA256 hash of the filtered manifest. It returns the digest and
// the final bytes used to produce that digest.
func getDigest(bundle *anywherev1alpha1.Bundles) ([32]byte, []byte, error) {
	var zero [32]byte

	// Marshal Bundles object to YAML
	yamlBytes, err := yaml.Marshal(bundle)
	if err != nil {
		return zero, nil, fmt.Errorf("marshalling bundle to YAML: %v", err)
	}

	// Convert YAML to JSON for easier gojq processing
	jsonBytes, err := yaml.YAMLToJSON(yamlBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("converting YAML to JSON: %v", err)
	}

	// Build and execute the gojq filter that deletes excluded fields
	filtered, err := filterExcludes(jsonBytes)
	if err != nil {
		return zero, nil, fmt.Errorf("filtering excluded fields: %v", err)
	}

	// Compute the SHA256 sum of the filtered JSON
	digest := sha256.Sum256(filtered)

	return digest, filtered, nil
}

// filterExcludes applies the default and user-specified excludes to the JSON
// representation of the Bundles object using gojq.
func filterExcludes(jsonBytes []byte) ([]byte, error) {
	// Decode the base64-encoded excludes
	exclBytes, err := base64.StdEncoding.DecodeString(Excludes)
	if err != nil {
		return nil, fmt.Errorf("decoding Excludes: %v", err)
	}
	// Convert them into slice of strings
	userExcludes := strings.Split(string(exclBytes), "\n")

	// Combine AlwaysExcluded with userExcludes
	allExcludes := append(AlwaysExcluded, userExcludes...)

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

// CreateKMSClient creates KMS client for bundle manifest signing.
func CreateKMSClient(ctx context.Context) (*kms.Client, error) {
	conf, err := config.LoadDefaultConfig(ctx, config.WithRegion(DefaultRegion))
	if err != nil {
		return nil, fmt.Errorf("loading AWS config in region %q: %v", DefaultRegion, err)
	}
	client := kms.NewFromConfig(conf)

	return client, nil
}
