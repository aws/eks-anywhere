/*
Package test provides YAML template matching utilities for testing Kubernetes manifests.

# Array Matching Semantics

All array matching is ORDER-AGNOSTIC and SUBSET-BASED at all nesting levels:
  - Arrays are matched by content, not position
  - Expected items can appear in any order in actual arrays
  - Actual arrays can contain extra items not in expected

This approach works well for most Kubernetes fields where order doesn't matter
(e.g., containers, env vars, volumes, labels). However, for order-sensitive
fields like args, command, or initContainers, exact matching would be more appropriate.

Example:

	Actual:   args: ["--port", "8080", "--host", "localhost"]
	Expected: args: ["--host", "localhost", "--port", "8080"]
	Result:   MATCHES (order-agnostic)

# Lenient Matching

Expected objects can be subsets of actual objects - actual can have extra fields.
This allows testing specific properties without requiring complete object definitions.
*/
package test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// TestHelper is an interface that matches the subset of testing.T methods
// used by our Assert functions. This allows using mock implementations for testing.
type TestHelper interface {
	Helper()
	Fatalf(format string, args ...interface{})
}

// ParseMultiDocYAML splits a multi-document YAML into separate objects.
func ParseMultiDocYAML(data []byte) ([]map[string]interface{}, error) {
	var objects []map[string]interface{}
	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var obj map[string]interface{}
		if err := decoder.Decode(&obj); err != nil {
			if err == io.EOF {
				break
			}
			return nil, fmt.Errorf("unmarshaling YAML document: %w", err)
		}

		// Skip empty documents
		if obj != nil {
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

// FindObjectByKind finds the first object in a multi-doc YAML matching the given kind.
func FindObjectByKind(objects []map[string]interface{}, kind string) (map[string]interface{}, error) {
	for _, obj := range objects {
		if k, ok := obj["kind"].(string); ok && k == kind {
			return obj, nil
		}
	}
	return nil, fmt.Errorf("object with kind %q not found", kind)
}

// FindObjectByKindAndName finds an object matching both kind and metadata.name.
func FindObjectByKindAndName(objects []map[string]interface{}, kind, name string) (map[string]interface{}, error) {
	for _, obj := range objects {
		if k, ok := obj["kind"].(string); ok && k == kind {
			if metadataRaw, ok := obj["metadata"]; ok {
				if metadata, ok := metadataRaw.(map[string]interface{}); ok {
					if n, ok := metadata["name"].(string); ok && n == name {
						return obj, nil
					}
				}
			}
		}
	}
	return nil, fmt.Errorf("object with kind %q and name %q not found", kind, name)
}

// GetYAMLPath extracts a value from a YAML object using dot-notation path.
// Example: GetYAMLPath(obj, "spec.kubeadmConfigSpec.files").
func GetYAMLPath(obj map[string]interface{}, path string) (interface{}, error) {
	parts := strings.Split(path, ".")
	current := interface{}(obj)

	for i, part := range parts {
		switch v := current.(type) {
		case map[string]interface{}:
			val, ok := v[part]
			if !ok {
				return nil, fmt.Errorf("path component %q not found at level %d", part, i)
			}
			current = val
		default:
			return nil, fmt.Errorf("cannot traverse path %q: expected map at level %d, got %T", path, i, current)
		}
	}

	return current, nil
}

// containsItem checks if an array contains an item that matches all fields in the provided item map.
// Uses lenient matching - actual items can have extra fields.
func containsItem(array interface{}, item map[string]interface{}) bool {
	items, ok := array.([]interface{})
	if !ok {
		return false
	}

	for _, arrItem := range items {
		if valueMatches(arrItem, item, "") {
			return true
		}
	}
	return false
}

// AssertContainsItemAtPath asserts that an array at a YAML path contains the specified item.
func AssertContainsItemAtPath(t TestHelper, obj map[string]interface{}, path string, item map[string]interface{}) {
	t.Helper()
	array, err := GetYAMLPath(obj, path)
	if err != nil {
		t.Fatalf("expected path %q to exist, but got error: %v", path, err)
	}
	if !containsItem(array, item) {
		t.Fatalf("at path %q: expected array to contain item %v", path, item)
	}
}

// AssertNotContainsItemAtPath asserts that an array at a YAML path does NOT contain the specified item.
func AssertNotContainsItemAtPath(t TestHelper, obj map[string]interface{}, path string, item map[string]interface{}) {
	t.Helper()
	array, err := GetYAMLPath(obj, path)
	if err != nil {
		return
	}
	if containsItem(array, item) {
		t.Fatalf("at path %q: expected array to NOT contain item %v", path, item)
	}
}

/*
AssertYAMLSubset verifies that expected YAML snippets are subsets of actual generated YAML.
- Parses both actual and expected as multi-doc YAML.
- Matches objects by kind and metadata.name from expected.
- Lenient matching: actual can have extra fields not in expected.
- Array matching: order-agnostic, verifies all expected items exist.
*/
func AssertYAMLSubset(t TestHelper, actualYAML []byte, expectedFile string) {
	t.Helper()

	actualObjects, err := ParseMultiDocYAML(actualYAML)
	if err != nil {
		t.Fatalf("failed to parse actual YAML: %v", err)
	}

	expectedData, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("failed to read expected file %s: %v", expectedFile, err)
	}

	expectedObjects, err := ParseMultiDocYAML(expectedData)
	if err != nil {
		t.Fatalf("failed to parse expected YAML from %s: %v", expectedFile, err)
	}

	for i, expectedObj := range expectedObjects {
		kind, _ := GetYAMLPath(expectedObj, "kind")
		name, _ := GetYAMLPath(expectedObj, "metadata.name")

		if kind == nil || name == nil {
			t.Fatalf("expected object[%d] missing kind or metadata.name", i)
		}

		actualObj, err := FindObjectByKindAndName(actualObjects, kind.(string), name.(string))
		if err != nil {
			t.Fatalf("expected object kind=%s name=%s not found in actual YAML", kind, name)
		}

		verifySubset(t, actualObj, expectedObj, fmt.Sprintf("object[%d] kind=%s name=%s", i, kind, name))
	}
}

// verifySubset recursively verifies that expected is a subset of actual (lenient matching).
func verifySubset(t TestHelper, actual, expected map[string]interface{}, path string) {
	t.Helper()

	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			t.Fatalf("at %s.%s: field missing in actual", path, key)
		}

		verifyValue(t, actualValue, expectedValue, fmt.Sprintf("%s.%s", path, key))
	}
}

// verifyValue verifies a single value matches (with type-specific handling).
func verifyValue(t TestHelper, actual, expected interface{}, path string) {
	t.Helper()

	switch ev := expected.(type) {
	case map[string]interface{}:
		av, ok := actual.(map[string]interface{})
		if !ok {
			t.Fatalf("at %s: expected object, got %T", path, actual)
		}
		verifySubset(t, av, ev, path)

	case []interface{}:
		av, ok := actual.([]interface{})
		if !ok {
			t.Fatalf("at %s: expected array, got %T", path, actual)
		}
		verifyArraySubset(t, av, ev, path)

	default:
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("at %s: expected %v (%T), got %v (%T)", path, expected, expected, actual, actual)
		}
	}
}

// verifyArraySubset verifies all expected items exist in actual array (order-agnostic).
func verifyArraySubset(t TestHelper, actual, expected []interface{}, path string) {
	t.Helper()

	if !arraySubsetMatches(actual, expected, path) {
		t.Fatalf("at %s: array subset verification failed - not all expected items found", path)
	}
}

// arraySubsetMatches checks if all expected items exist in actual array (order-agnostic, subset-based).
// NOTE: This order-agnostic behavior is appropriate for most Kubernetes fields but may not
// be suitable for order-sensitive fields like args, command, or initContainers where
// sequential order has semantic meaning.
func arraySubsetMatches(actual, expected []interface{}, path string) bool {
	for _, expectedItem := range expected {
		found := false
		for j, actualItem := range actual {
			itemPath := fmt.Sprintf("%s[%d]", path, j)
			if valueMatches(actualItem, expectedItem, itemPath) {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

// mapSubset checks if expected map is a subset of actual map (lenient).
func mapSubset(actual, expected map[string]interface{}, path string) bool {
	for key, expectedValue := range expected {
		actualValue, exists := actual[key]
		if !exists {
			return false
		}

		fieldPath := fmt.Sprintf("%s.%s", path, key)
		if !valueMatches(actualValue, expectedValue, fieldPath) {
			return false
		}
	}
	return true
}

// valueMatches checks if values match (with type-specific handling).
func valueMatches(actual, expected interface{}, path string) bool {
	switch ev := expected.(type) {
	case map[string]interface{}:
		av, ok := actual.(map[string]interface{})
		if !ok {
			return false
		}
		return mapSubset(av, ev, path)

	case []interface{}:
		av, ok := actual.([]interface{})
		if !ok {
			return false
		}
		return arraySubsetMatches(av, ev, path)

	default:
		return reflect.DeepEqual(actual, expected)
	}
}
