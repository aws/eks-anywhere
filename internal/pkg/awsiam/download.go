package awsiam

import (
	"bytes"
	"fmt"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"
	"golang.org/x/sys/unix"

	"github.com/aws/eks-anywhere/internal/pkg/files"
)

const awsIamClientBinary = "aws-iam-authenticator"

func DownloadAwsIamAuthClient(eksdRelease *eksdv1alpha1.Release) error {
	k, err := getKernalName()
	if err != nil {
		return fmt.Errorf("error getting kernel name: %v", err)
	}
	uri, err := getAwsIamAuthClientUri(eksdRelease, strings.ToLower(k))
	if err != nil {
		return fmt.Errorf("error getting %s uri: %v", awsIamClientBinary, err)
	}
	err = files.GzipFileDownloadExtract(uri, awsIamClientBinary, "bin")
	if err != nil {
		return fmt.Errorf("error downloading %s binary: %v", awsIamClientBinary, err)
	}
	return nil
}

func getKernalName() (string, error) {
	var utsname unix.Utsname
	err := unix.Uname(&utsname)
	if err != nil {
		return "", fmt.Errorf("uname call failure: %v", err)
	}
	return string(bytes.Trim(utsname.Sysname[:], "\x00")), nil
}

func getAwsIamAuthClientUri(eksdRelease *eksdv1alpha1.Release, kernalName string) (string, error) {
	for _, component := range eksdRelease.Status.Components {
		if component.Name == awsIamClientBinary {
			for _, asset := range component.Assets {
				if asset.Type == "Archive" && asset.OS == kernalName && asset.Arch[0] == "amd64" {
					return asset.Archive.URI, nil
				}
			}
		}
	}
	return "", fmt.Errorf("component %s not found in EKS-D release spec", awsIamClientBinary)
}
