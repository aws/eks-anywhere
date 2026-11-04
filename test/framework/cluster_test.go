package framework

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/mock/gomock"
	yamlv3 "gopkg.in/yaml.v3"

	packagesv1 "github.com/aws/eks-anywhere-packages/api/v1alpha1"
	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
)

type fileDumpTest struct {
	T
	log string
}

func (t *fileDumpTest) Logf(format string, args ...any) {
	t.log += fmt.Sprintf(format, args...)
}

func TestDumpFileRedactsLicenseTokens(t *testing.T) {
	content := `apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: Cluster
metadata:
  name: test-cluster
spec:
  kubernetesVersion: "1.31"
  licenseToken: &license dummy-license-value
  controlPlaneConfiguration:
    count: 1
---
apiVersion: anywhere.eks.amazonaws.com/v1alpha1
kind: VSphereMachineConfig
metadata:
  name: test-machine
spec:
  template: dummy-template
  licenseToken: |
    dummy-multiline-license
    dummy-license-continuation
`
	logger := dumpDummyFile(t, "cluster.yaml", content)
	for _, token := range []string{"dummy-license-value", "dummy-multiline-license", "dummy-license-continuation"} {
		if strings.Contains(logger.log, token) {
			t.Fatal("dumpFile logged a license token")
		}
	}
	output := strings.TrimPrefix(logger.log, "dummy file:\n")
	decoder := yamlv3.NewDecoder(strings.NewReader(output))
	for _, expected := range []struct {
		kind string
		name string
	}{
		{kind: "Cluster", name: "test-cluster"},
		{kind: "VSphereMachineConfig", name: "test-machine"},
	} {
		assertDumpedResource(t, decoder, expected.kind, expected.name)
	}
}

func assertDumpedResource(t *testing.T, decoder *yamlv3.Decoder, kind, name string) {
	t.Helper()
	var document struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name string `yaml:"name"`
		} `yaml:"metadata"`
		Spec map[string]any `yaml:"spec"`
	}
	if err := decoder.Decode(&document); err != nil {
		t.Fatal(err)
	}
	if document.Kind != kind || document.Metadata.Name != name {
		t.Fatalf("dumpFile lost useful resource information: %+v", document)
	}
	if document.Spec["licenseToken"] != "<redacted>" {
		t.Fatal("dumpFile did not mark the license token as redacted")
	}
	if kind == "Cluster" && document.Spec["kubernetesVersion"] != "1.31" {
		t.Fatal("dumpFile lost the Kubernetes version")
	}
	if kind == "VSphereMachineConfig" && document.Spec["template"] != "dummy-template" {
		t.Fatal("dumpFile lost the machine template")
	}
}

func TestDumpFileOmitsInvalidConfigWithLicense(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cluster.yaml")
	if err := os.WriteFile(path, []byte("licenseToken: dummy-license\ninvalid: ["), 0o600); err != nil {
		t.Fatal(err)
	}
	logger := &fileDumpTest{T: t}
	dumpFile("Cluster config", path, logger)
	if strings.Contains(logger.log, "dummy-license") || !strings.Contains(logger.log, "file content omitted") {
		t.Fatal("dumpFile exposed an invalid config containing a license token")
	}
}

func TestDumpFileRedactsStructuredLicenseKeysAndAliases(t *testing.T) {
	for _, content := range []string{
		`metadata: {name: dummy-cluster}
"license\u0054oken": dummy-escaped-license
`,
		`{"metadata":{"name":"dummy-cluster"},"license\u0054oken":"dummy-escaped-license"}`,
		`metadata: {name: dummy-cluster}
copiedToken: &token dummy-escaped-license
licenseToken: *token
anotherCopy: *token
`,
		`metadata: {name: dummy-cluster}
licenseToken: &token dummy-escaped-license
anotherCopy: *token
`,
		`metadata: {name: dummy-cluster}
licenseToken:
  nested: &token dummy-escaped-license
anotherCopy: *token
`,
	} {
		t.Run(content, func(t *testing.T) {
			logger := dumpDummyFile(t, "cluster.yaml", content)
			if strings.Contains(logger.log, "dummy-escaped-license") {
				t.Fatal("dumpFile exposed a structured license token")
			}
			if !strings.Contains(logger.log, "dummy-cluster") {
				t.Fatal("dumpFile lost the resource name")
			}
			var document yamlv3.Node
			output := strings.TrimPrefix(logger.log, "dummy file:\n")
			if err := yamlv3.Unmarshal([]byte(output), &document); err != nil {
				t.Fatal("dumpFile produced invalid YAML")
			}
		})
	}
}

func TestDumpFileRedactsHardwareCSVPasswords(t *testing.T) {
	content := "hostname,bmc_password,bmc_ip,ip_address,bmc_password\n" +
		"dummy-host,\"dummy-password,with-comma\",192.0.2.10,192.0.2.11,dummy-second-password\n"
	logger := dumpDummyFile(t, "hardware.csv", content)
	if strings.Contains(logger.log, "dummy-password") || strings.Contains(logger.log, "dummy-second-password") {
		t.Fatal("dumpFile exposed a hardware password")
	}
	output := strings.TrimPrefix(logger.log, "dummy file:\n")
	records, err := csv.NewReader(strings.NewReader(output)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := "dummy-host,<redacted>,192.0.2.10,192.0.2.11,<redacted>"
	if len(records) != 2 || strings.Join(records[1], ",") != want {
		t.Fatal("dumpFile lost hardware identification or redaction")
	}
}

func TestDumpFileOmitsInvalidStructuredFiles(t *testing.T) {
	for _, test := range []struct {
		name    string
		content string
	}{
		{name: "cluster.yaml", content: "\"license\\u0054oken\": dummy-private-value\ninvalid: ["},
		{name: "hardware.csv", content: "hostname,bmc_password\nhost,dummy-private-value,extra"},
		{name: "quoted.csv", content: "hostname,bmc_password\nhost,\"dummy-private-value"},
	} {
		t.Run(test.name, func(t *testing.T) {
			logger := dumpDummyFile(t, test.name, test.content)
			if strings.Contains(logger.log, "dummy-private-value") || !strings.Contains(logger.log, "file content omitted") {
				t.Fatal("dumpFile exposed an invalid structured file")
			}
		})
	}
}

func dumpDummyFile(t *testing.T, name, content string) *fileDumpTest {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	logger := &fileDumpTest{T: t}
	dumpFile("dummy file", path, logger)
	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(stored) != content {
		t.Fatal("dumpFile changed the input file")
	}
	return logger
}

func TestValidatePackageBundleControllerRegistry(t *testing.T) {
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)
	testPbc := &packagesv1.PackageBundleController{
		Spec: packagesv1.PackageBundleControllerSpec{
			DefaultRegistry:      "123.ecr",
			DefaultImageRegistry: "123.ecr",
		},
	}
	respJSON, err := json.Marshal(testPbc)
	if err != nil {
		t.Errorf("marshaling test service: %s", err)
	}
	ret := bytes.NewBuffer(respJSON)
	expectedParam := []string{"get", "packagebundlecontroller.packages.eks.amazonaws.com", "test", "-o", "json", "--kubeconfig", "test/test-eks-a-cluster.kubeconfig", "--namespace", "eksa-packages", "--ignore-not-found=true"}
	t.Run("CuratedPackagesTest", func(t *testing.T) {
		e := &ClusterE2ETest{
			T:                   t,
			ClusterConfigFolder: "test",
			ClusterName:         "test",
			KubectlClient:       executables.NewKubectl(executable),
		}

		executable.EXPECT().Execute(gomock.Any(), gomock.Eq(expectedParam)).Return(*ret, nil).AnyTimes()
		e.ValidatePackageBundleControllerRegistry()
	})
}
