package logger_test

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func TestInitZap(t *testing.T) {
	g := NewWithT(t)
	logFile := filepath.Join(t.TempDir(), "test.log")
	err := logger.InitZap(logger.ZapOpts{
		Level:          0,
		OutputFilePath: logFile,
	})
	g.Expect(err).To(BeNil())
	g.Expect(logger.Get()).ToNot(Equal(logr.Discard()))
	g.Expect(logger.GetOutputFilePath()).To(Equal(logFile))
}

func TestNewZapWithNames(t *testing.T) {
	g := NewWithT(t)
	logOut := filepath.Join(t.TempDir(), "test.log")
	l, err := logger.NewZap(logger.ZapOpts{
		Level:          0,
		OutputFilePath: logOut,
		WithNames:      []string{"test-logger"},
	})

	g.Expect(err).To(BeNil())
	g.Expect(l).ToNot(Equal(logr.Discard()))

	l.Info("debug log with name")
	l.Error(errors.New("test error"), "error log with name")

	byteContents, err := os.ReadFile(logOut)
	g.Expect(err).To(BeNil())
	g.Expect(string(byteContents)).To(ContainSubstring("\"N\":\"test-logger\",\"M\":\"debug log with name\""))
	g.Expect(string(byteContents)).To(ContainSubstring("\"N\":\"test-logger\",\"M\":\"error log with name\",\"error\":\"test error\""))
}

func TestNewZapWithOutputFilePath(t *testing.T) {
	g := NewWithT(t)
	logOut := filepath.Join(t.TempDir(), "test.log")
	l, err := logger.NewZap(logger.ZapOpts{
		Level:          0,
		OutputFilePath: logOut,
	})

	g.Expect(err).To(BeNil())
	g.Expect(l).ToNot(Equal(logr.Discard()))

	l.Info("debug log")
	l.Error(errors.New("test error"), "error log")

	byteContents, err := os.ReadFile(logOut)
	g.Expect(err).To(BeNil())
	g.Expect(string(byteContents)).To(ContainSubstring("\"M\":\"debug log\""))
	g.Expect(string(byteContents)).To(ContainSubstring("\"M\":\"error log\",\"error\":\"test error\""))
}

func TestZapWithInvalidOutputPaths(t *testing.T) {
	tests := []struct {
		name  string
		path  string
		match string
	}{
		{
			name:  "output path doesn't exist",
			path:  "/temp/does-not-exist/foo.log",
			match: "no such file or directory",
		},
		{
			name:  "bad schema",
			path:  fmt.Sprintf("foo-%s.log", time.Now().Format("2006-01-02T15:04:05")),
			match: "no sink found for scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			l, err := logger.NewZap(logger.ZapOpts{
				Level:          0,
				OutputFilePath: tt.path,
			})

			g.Expect(l).To(Equal(logr.Discard()))
			g.Expect(err).To(MatchError(ContainSubstring(tt.match)))
		})
	}
}
