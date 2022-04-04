package tar_test

import (
	"os"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/tar"
)

func TestTarFolder(t *testing.T) {
	dstFile := "dst.tar"
	t.Cleanup(func() {
		os.Remove(dstFile)
	})

	g := NewWithT(t)
	g.Expect(tar.TarFolder("testdata", dstFile)).To(Succeed())
	g.Expect(dstFile).To(BeAnExistingFile())
}
