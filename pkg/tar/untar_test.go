package tar_test

import (
	stdtar "archive/tar"
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/tar"
)

func TestUntarFile(t *testing.T) {
	g := NewWithT(t)
	dstFile := "dst.tar"
	untarFolder := "dst-untar"
	g.Expect(os.MkdirAll(untarFolder, os.ModePerm))
	t.Cleanup(func() {
		os.Remove(dstFile)
		os.RemoveAll(untarFolder)
	})

	g.Expect(tar.TarFolder("testdata", dstFile)).To(Succeed())
	g.Expect(dstFile).To(BeAnExistingFile())

	g.Expect(tar.UntarFile(dstFile, untarFolder)).To(Succeed())
	g.Expect(untarFolder).To(BeADirectory())
	g.Expect(filepath.Join(untarFolder, "dummy1")).To(BeARegularFile())
	g.Expect(filepath.Join(untarFolder, "dummy2")).To(BeARegularFile())
	g.Expect(filepath.Join(untarFolder, "dummy3")).To(BeADirectory())
	g.Expect(filepath.Join(untarFolder, "dummy3", "dummy4")).To(BeARegularFile())
}

func TestUntarFile_DirTraversalComponents(t *testing.T) {
	// This test ensures Untar fails when a tarball contains paths with directory traversal
	// components. It addresses https://cwe.mitre.org/data/definitions/22.html.
	g := NewWithT(t)

	dir := t.TempDir()
	tarPath := filepath.Join(dir, "test")
	fh, err := os.Create(tarPath)
	g.Expect(err).To(Succeed())

	createArbitraryTarball(t, fh)

	g.Expect(tar.UntarFile(tarPath, dir)).ToNot(Succeed())
}

func createArbitraryTarball(t *testing.T, w io.Writer) {
	t.Helper()

	tb := stdtar.NewWriter(w)

	data := bytes.NewBufferString("Hello, world!")
	header := stdtar.Header{
		Name:     "../foobar",
		Mode:     int64(fs.ModePerm),
		Typeflag: stdtar.TypeReg,
		Size:     int64(data.Len()),
	}

	if err := tb.WriteHeader(&header); err != nil {
		t.Fatal(err)
	}

	_, err := io.Copy(tb, data)
	if err != nil {
		t.Fatal(err)
	}

	tb.Close()
}
