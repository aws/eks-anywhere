package tar_test

import (
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/tar"
)

func TestUnGzipTarFile(t *testing.T) {
	g := NewWithT(t)
	dstFile := "dst.tar.gz"
	untarFolder := "dst-untar"
	g.Expect(os.MkdirAll(untarFolder, os.ModePerm))
	t.Cleanup(func() {
		os.Remove(dstFile)
		os.RemoveAll(untarFolder)
	})

	g.Expect(tar.GzipTarFolder("testdata", dstFile)).To(Succeed())
	g.Expect(dstFile).To(BeAnExistingFile())

	g.Expect(tar.UnGzipTarFile(dstFile, untarFolder)).To(Succeed())
	g.Expect(untarFolder).To(BeADirectory())
	g.Expect(filepath.Join(untarFolder, "dummy1")).To(BeARegularFile())
	g.Expect(filepath.Join(untarFolder, "dummy2")).To(BeARegularFile())
	g.Expect(filepath.Join(untarFolder, "dummy3")).To(BeADirectory())
	g.Expect(filepath.Join(untarFolder, "dummy3", "dummy4")).To(BeARegularFile())
}
