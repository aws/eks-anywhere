package test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mockT implements TestHelper interface to record failures
// instead of stopping execution. This allows testing negative cases.
type mockT struct {
	failed bool
	logs   []string
}

func (m *mockT) Helper() {}

func (m *mockT) Fatalf(format string, args ...interface{}) {
	m.failed = true
	m.logs = append(m.logs, fmt.Sprintf(format, args...))
}

func TestParseMultiDocYAML(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantObjects int
		wantErr     bool
	}{
		{
			name: "single document",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: test`,
			wantObjects: 1,
			wantErr:     false,
		},
		{
			name: "multi document",
			input: `apiVersion: v1
kind: Pod
metadata:
  name: test1
---
apiVersion: v1
kind: Service
metadata:
  name: test2`,
			wantObjects: 2,
			wantErr:     false,
		},
		{
			name:        "empty document",
			input:       "",
			wantObjects: 0,
			wantErr:     false,
		},
		{
			name:        "invalid yaml",
			input:       "invalid: yaml: content:",
			wantObjects: 0,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objects, err := ParseMultiDocYAML([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseMultiDocYAML() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(objects) != tt.wantObjects {
				t.Errorf("ParseMultiDocYAML() got %d objects, want %d", len(objects), tt.wantObjects)
			}
		})
	}
}

func TestFindObjectByKind(t *testing.T) {
	objects := []map[string]interface{}{
		{"kind": "Pod", "metadata": map[string]interface{}{"name": "test1"}},
		{"kind": "Service", "metadata": map[string]interface{}{"name": "test2"}},
	}

	tests := []struct {
		name    string
		kind    string
		wantErr bool
	}{
		{"finds pod", "Pod", false},
		{"finds service", "Service", false},
		{"kind not found", "Deployment", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := FindObjectByKind(objects, tt.kind)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindObjectByKind() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && obj == nil {
				t.Error("FindObjectByKind() returned nil object")
			}
		})
	}
}

func TestFindObjectByKindAndName(t *testing.T) {
	objects := []map[string]interface{}{
		{"kind": "Pod", "metadata": map[string]interface{}{"name": "test1"}},
		{"kind": "Pod", "metadata": map[string]interface{}{"name": "test2"}},
		{"kind": "Service", "metadata": map[string]interface{}{"name": "test1"}},
	}

	tests := []struct {
		name     string
		kind     string
		objName  string
		wantErr  bool
		wantName string
	}{
		{"finds pod test1", "Pod", "test1", false, "test1"},
		{"finds pod test2", "Pod", "test2", false, "test2"},
		{"finds service test1", "Service", "test1", false, "test1"},
		{"kind not found", "Deployment", "test1", true, ""},
		{"name not found", "Pod", "test3", true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			obj, err := FindObjectByKindAndName(objects, tt.kind, tt.objName)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindObjectByKindAndName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if obj == nil {
					t.Error("FindObjectByKindAndName() returned nil object")
					return
				}
				metadata := obj["metadata"].(map[string]interface{})
				if metadata["name"] != tt.wantName {
					t.Errorf("FindObjectByKindAndName() name = %v, want %v", metadata["name"], tt.wantName)
				}
			}
		})
	}
}

func TestGetYAMLPath(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"replicas": float64(3),
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "test",
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		path      string
		wantValue interface{}
		wantErr   bool
	}{
		{"simple path", "spec.replicas", float64(3), false},
		{"nested path", "spec.template.metadata.labels.app", "test", false},
		{"path not found", "spec.notexist", nil, true},
		{"invalid path depth", "spec.replicas.invalid", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := GetYAMLPath(obj, tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetYAMLPath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && value != tt.wantValue {
				t.Errorf("GetYAMLPath() = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestAssertYAMLSubset(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir := t.TempDir()

	// Test case 1: Valid subset
	t.Run("valid subset", func(t *testing.T) {
		actualYAML := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
spec:
  replicas: 3
  kubeadmConfigSpec:
    files:
    - path: /etc/test.conf
      owner: root:root
      content: "large content here"
    - path: /etc/other.conf
      owner: root:root`)

		expectedFile := filepath.Join(tmpDir, "expected.yaml")
		expectedContent := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
spec:
  replicas: 3
  kubeadmConfigSpec:
    files:
    - path: /etc/test.conf
      owner: root:root`)

		err := os.WriteFile(expectedFile, expectedContent, 0o644)
		if err != nil {
			t.Fatalf("failed to write expected file: %v", err)
		}

		AssertYAMLSubset(t, actualYAML, expectedFile)
	})

	// Test case 2: Order-agnostic array matching
	t.Run("order agnostic arrays", func(t *testing.T) {
		actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: container2
    image: img2
  - name: container1
    image: img1`)

		expectedFile := filepath.Join(tmpDir, "expected_order.yaml")
		expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: container1
    image: img1`)

		err := os.WriteFile(expectedFile, expectedContent, 0o644)
		if err != nil {
			t.Fatalf("failed to write expected file: %v", err)
		}

		AssertYAMLSubset(t, actualYAML, expectedFile)
	})

	// Test case 3: Lenient matching ignores extra fields
	t.Run("extra fields ignored", func(t *testing.T) {
		actualYAML := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
spec:
  kubeadmConfigSpec:
    files:
    - path: /etc/test.conf
      owner: root:root
      content: "actual large content"
      permissions: "0644"`)

		expectedFile := filepath.Join(tmpDir, "expected_no_extra.yaml")
		expectedContent := []byte(`apiVersion: controlplane.cluster.x-k8s.io/v1beta1
kind: KubeadmControlPlane
metadata:
  name: test
spec:
  kubeadmConfigSpec:
    files:
    - path: /etc/test.conf
      owner: root:root`)

		err := os.WriteFile(expectedFile, expectedContent, 0o644)
		if err != nil {
			t.Fatalf("failed to write expected file: %v", err)
		}

		AssertYAMLSubset(t, actualYAML, expectedFile)
	})
}

func TestAssertContainsItemAtPath(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"name": "item1", "value": "val1"},
				map[string]interface{}{"name": "item2", "value": "val2"},
			},
		},
	}

	t.Run("item exists in array", func(t *testing.T) {
		// Should not fail
		AssertContainsItemAtPath(t, obj, "spec.items",
			map[string]interface{}{"name": "item1"})
	})
}

func TestAssertContainsItemAtPath_NestedArrays(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name":  "container1",
					"args":  []interface{}{"arg2", "arg1", "arg3"},
					"ports": []interface{}{8080, 9090},
				},
				map[string]interface{}{
					"name": "container2",
					"env":  []interface{}{"ENV1", "ENV2"},
				},
			},
		},
	}

	t.Run("order agnostic nested array matching", func(t *testing.T) {
		// Should match even though order is different
		AssertContainsItemAtPath(t, obj, "spec.containers",
			map[string]interface{}{
				"name": "container1",
				"args": []interface{}{"arg1", "arg2", "arg3"},
			})
	})

	t.Run("subset nested array matching", func(t *testing.T) {
		// Should match - expected args are subset of actual
		AssertContainsItemAtPath(t, obj, "spec.containers",
			map[string]interface{}{
				"name": "container1",
				"args": []interface{}{"arg1", "arg3"},
			})
	})

	t.Run("nested array with different order", func(t *testing.T) {
		// Should match ports in different order
		AssertContainsItemAtPath(t, obj, "spec.containers",
			map[string]interface{}{
				"name":  "container1",
				"ports": []interface{}{9090, 8080},
			})
	})
}

func TestAssertContainsItemAtPath_NestedArrays_Negative(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"containers": []interface{}{
				map[string]interface{}{
					"name": "container1",
					"args": []interface{}{"arg1", "arg2"},
				},
			},
		},
	}

	t.Run("nested array with extra expected items should fail", func(t *testing.T) {
		mockT := &mockT{}
		// Expected has arg3 which doesn't exist in actual
		AssertContainsItemAtPath(mockT, obj, "spec.containers",
			map[string]interface{}{
				"name": "container1",
				"args": []interface{}{"arg1", "arg2", "arg3"},
			})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail due to extra array item")
		}
	})

	t.Run("nested array with wrong items should fail", func(t *testing.T) {
		mockT := &mockT{}
		AssertContainsItemAtPath(mockT, obj, "spec.containers",
			map[string]interface{}{
				"name": "container1",
				"args": []interface{}{"arg1", "wrong"},
			})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail due to wrong array item")
		}
	})
}

func TestAssertNotContainsItemAtPath(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"name": "item1"},
			},
		},
	}

	t.Run("item does not exist", func(t *testing.T) {
		// Should not fail
		AssertNotContainsItemAtPath(t, obj, "spec.items",
			map[string]interface{}{"name": "item2"})
	})

	t.Run("path does not exist", func(t *testing.T) {
		// Should not fail - missing path means item doesn't exist
		AssertNotContainsItemAtPath(t, obj, "spec.notexist",
			map[string]interface{}{"name": "item1"})
	})
}

// Negative test cases for Assert functions
// These tests use a mock that embeds *testing.T but intercepts Fatalf calls
// to verify that Assert functions properly fail in negative scenarios.

func TestAssertContainsItemAtPath_Negative(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"name": "item1", "value": "val1"},
				map[string]interface{}{"name": "item2", "value": "val2"},
			},
		},
	}

	t.Run("item not in array should fail", func(t *testing.T) {
		mockT := &mockT{}
		AssertContainsItemAtPath(mockT, obj, "spec.items",
			map[string]interface{}{"name": "item3"})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail, but it passed")
		}
		if len(mockT.logs) == 0 {
			t.Error("expected error message to be logged")
		}
		if !strings.Contains(mockT.logs[0], "expected array to contain item") {
			t.Errorf("expected error about item not found, got: %s", mockT.logs[0])
		}
	})

	t.Run("path does not exist should fail", func(t *testing.T) {
		mockT := &mockT{}
		AssertContainsItemAtPath(mockT, obj, "spec.notexist",
			map[string]interface{}{"name": "item1"})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail, but it passed")
		}
		if len(mockT.logs) == 0 {
			t.Error("expected error message to be logged")
		}
		if !strings.Contains(mockT.logs[0], "expected path") {
			t.Errorf("expected error about missing path, got: %s", mockT.logs[0])
		}
	})

	t.Run("path points to non-array should fail", func(t *testing.T) {
		objWithString := map[string]interface{}{
			"spec": map[string]interface{}{
				"items": "not an array",
			},
		}
		mockT := &mockT{}
		AssertContainsItemAtPath(mockT, objWithString, "spec.items",
			map[string]interface{}{"name": "item1"})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail, but it passed")
		}
	})

	t.Run("item with different value should fail", func(t *testing.T) {
		mockT := &mockT{}
		AssertContainsItemAtPath(mockT, obj, "spec.items",
			map[string]interface{}{"name": "item1", "value": "wrong"})

		if !mockT.failed {
			t.Error("expected AssertContainsItemAtPath to fail, but it passed")
		}
	})
}

func TestAssertNotContainsItemAtPath_Negative(t *testing.T) {
	obj := map[string]interface{}{
		"spec": map[string]interface{}{
			"items": []interface{}{
				map[string]interface{}{"name": "item1", "value": "val1"},
				map[string]interface{}{"name": "item2", "value": "val2"},
			},
		},
	}

	t.Run("item exists in array should fail", func(t *testing.T) {
		mockT := &mockT{}
		AssertNotContainsItemAtPath(mockT, obj, "spec.items",
			map[string]interface{}{"name": "item1"})

		if !mockT.failed {
			t.Error("expected AssertNotContainsItemAtPath to fail, but it passed")
		}
		if len(mockT.logs) == 0 {
			t.Error("expected error message to be logged")
		}
		if !strings.Contains(mockT.logs[0], "expected array to NOT contain item") {
			t.Errorf("expected error about item found, got: %s", mockT.logs[0])
		}
	})

	t.Run("partial match exists should fail", func(t *testing.T) {
		mockT := &mockT{}
		// Lenient matching means item1 with value will match item1 without value
		AssertNotContainsItemAtPath(mockT, obj, "spec.items",
			map[string]interface{}{"name": "item1"})

		if !mockT.failed {
			t.Error("expected AssertNotContainsItemAtPath to fail, but it passed")
		}
	})
}

func TestAssertYAMLSubset_MissingField(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: container1`)

	expectedFile := filepath.Join(tmpDir, "expected_missing_field.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  replicas: 3`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to missing field")
	}
	if len(mockT.logs) == 0 {
		t.Error("expected error message to be logged")
	}
}

func TestAssertYAMLSubset_TypeMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  replicas: "3"`)

	expectedFile := filepath.Join(tmpDir, "expected_type_mismatch.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  replicas: 3`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to type mismatch")
	}
}

func TestAssertYAMLSubset_ValueMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  replicas: 5`)

	expectedFile := filepath.Join(tmpDir, "expected_value_mismatch.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  replicas: 3`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to value mismatch")
	}
}

func TestAssertYAMLSubset_MissingObject(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test1`)

	expectedFile := filepath.Join(tmpDir, "expected_missing_object.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test2`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to missing object")
	}
	if len(mockT.logs) == 0 {
		t.Error("expected error message to be logged")
	}
	if !strings.Contains(mockT.logs[0], "not found in actual YAML") {
		t.Errorf("expected error about missing object, got: %s", mockT.logs[0])
	}
}

func TestAssertYAMLSubset_ArraySubsetMismatch(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: container1
    image: img1`)

	expectedFile := filepath.Join(tmpDir, "expected_array_mismatch.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test
spec:
  containers:
  - name: container1
    image: img1
  - name: container2
    image: img2`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to array subset mismatch")
	}
	if len(mockT.logs) == 0 {
		t.Error("expected error message to be logged")
	}
}

func TestAssertYAMLSubset_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`invalid: yaml: content:`)

	expectedFile := filepath.Join(tmpDir, "expected_valid.yaml")
	expectedContent := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test`)

	err := os.WriteFile(expectedFile, expectedContent, 0o644)
	if err != nil {
		t.Fatalf("failed to write expected file: %v", err)
	}

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to invalid YAML")
	}
	if len(mockT.logs) == 0 {
		t.Error("expected error message to be logged")
	}
	if !strings.Contains(mockT.logs[0], "failed to parse actual YAML") {
		t.Errorf("expected error about parsing YAML, got: %s", mockT.logs[0])
	}
}

func TestAssertYAMLSubset_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()

	actualYAML := []byte(`apiVersion: v1
kind: Pod
metadata:
  name: test`)

	expectedFile := filepath.Join(tmpDir, "nonexistent.yaml")

	mockT := &mockT{}
	AssertYAMLSubset(mockT, actualYAML, expectedFile)

	if !mockT.failed {
		t.Error("expected AssertYAMLSubset to fail due to missing file")
	}
	if len(mockT.logs) == 0 {
		t.Error("expected error message to be logged")
	}
	if !strings.Contains(mockT.logs[0], "failed to read expected file") {
		t.Errorf("expected error about missing file, got: %s", mockT.logs[0])
	}
}
