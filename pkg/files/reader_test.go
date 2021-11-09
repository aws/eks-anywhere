package files_test

import (
	"embed"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/files"
)

//go:embed testdata
var testdataFS embed.FS

func TestReaderReadFileError(t *testing.T) {
	tests := []struct {
		testName string
		uri      string
		filePath string
	}{
		{
			testName: "missing local file",
			uri:      "fake-local-file.yaml",
		},
		{
			testName: "missing embed file",
			uri:      "embed:///fake-local-file.yaml",
		},
		{
			testName: "invalid uri",
			uri:      ":domain.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			r := files.NewReader()
			_, err := r.ReadFile(tt.uri)
			g.Expect(err).NotTo(BeNil())
		})
	}
}

func TestReaderReadFileSuccess(t *testing.T) {
	tests := []struct {
		testName string
		uri      string
		filePath string
	}{
		{
			testName: "local file",
			uri:      "testdata/file.yaml",
			filePath: "testdata/file.yaml",
		},
		{
			testName: "embed file",
			uri:      "embed:///testdata/file.yaml",
			filePath: "testdata/file.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			r := files.NewReader(files.WithEmbedFS(testdataFS))
			got, err := r.ReadFile(tt.uri)
			g.Expect(err).To(BeNil())
			test.AssertContentToFile(t, string(got), tt.filePath)
		})
	}
}
