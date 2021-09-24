package conformance

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

const (
	destinationFile = "sonobuoy"
	sonobuoyDarwin  = "https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.53.2/sonobuoy_0.53.2_darwin_amd64.tar.gz"
	sonobuoyLinux   = "https://github.com/vmware-tanzu/sonobuoy/releases/download/v0.53.2/sonobuoy_0.53.2_linux_amd64.tar.gz"
)

func Download() error {
	var err error

	if _, err := os.Stat(destinationFile); err == nil {
		fmt.Println("Nothing downloaded file already exists: " + destinationFile)
		return nil
	}

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

	var resp *http.Response
	client := &http.Client{
		Timeout: time.Second * 240,
	}
	resp, err = client.Get(downloadFile)
	if err != nil {
		return fmt.Errorf("error opening download: %v", err)
	}
	defer resp.Body.Close()

	gzf, err := gzip.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("error initializing gzip: %v", err)
	}
	defer gzf.Close()

	tarReader := tar.NewReader(gzf)
	for {
		header, err := tarReader.Next()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("error reading archive: %v", err)
		}

		if header.Typeflag == tar.TypeReg {
			name := header.Name
			if name == destinationFile {
				out, err := os.Create(name)
				if err != nil {
					return fmt.Errorf("error opening sonobuoy file: %v", err)
				}
				defer out.Close()
				_, err = io.Copy(out, tarReader)
				if err != nil {
					return fmt.Errorf("error writing sonobuoy file: %v", err)
				}

				err = os.Chmod(name, 0o755)
				if err != nil {
					return fmt.Errorf("error setting permissions on sonobuoy file: %v", err)
				}
				fmt.Println("Downloaded ./" + destinationFile)
				return nil
			}
		}
	}

	return fmt.Errorf("did not find sonobuoy file in download")
}
