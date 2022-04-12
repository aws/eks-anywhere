package files_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/pkg/files"
	"github.com/aws/eks-anywhere/internal/test"
)

func TestGzipFileDownloadExtract(t *testing.T) {
	g := NewWithT(t)
	server := test.NewHTTPServerForFile(t, "testdata/dummy1.tar.gz")
	dst := "testdata/bin"
	file := "dummy1"
	g.Expect(os.MkdirAll(dst, os.ModePerm))
	t.Cleanup(func() {
		os.RemoveAll(dst)
	})
	g.Expect(files.GzipFileDownloadExtract(server.URL, file, dst)).To(Succeed())
	targetFile := filepath.Join(dst, file)
	g.Expect(targetFile).To(BeARegularFile())
	test.AssertFilesEquals(t, targetFile, "testdata/dummy1")
}
