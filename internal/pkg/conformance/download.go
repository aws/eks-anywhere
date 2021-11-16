package conformance

import (
	"bytes"
	"fmt"

	"golang.org/x/sys/unix"

	"github.com/aws/eks-anywhere/internal/pkg/files"
)

const (
	destinationFile = "sonobuoy"
	sonobuoyDarwin  = "https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.53.2/sonobuoy_0.53.2_darwin_amd64.tar.gz"
	sonobuoyLinux   = "https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.53.2/sonobuoy_0.53.2_linux_amd64.tar.gz"
)

func Download() error {
	var err error
	var utsname unix.Utsname
	err = unix.Uname(&utsname)
	if err != nil {
		return fmt.Errorf("uname call failure: %v", err)
	}

	var downloadFile string
	sysname := string(bytes.Trim(utsname.Sysname[:], "\x00"))
	if sysname == "Darwin" {
		downloadFile = sonobuoyDarwin
	} else {
		downloadFile = sonobuoyLinux
	}
	fmt.Println("Downloading sonobuoy for " + sysname + ": " + downloadFile)
	err = files.GzipFileDownloadExtract(downloadFile, destinationFile, "")
	if err != nil {
		return fmt.Errorf("failed to download sonobouy: %v", err)
	}
	return nil
}
