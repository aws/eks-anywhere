package awsiam

import (
	"fmt"
	"runtime"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/eks-anywhere/internal/pkg/files"
)

const awsIamClientBinary = "aws-iam-authenticator"

func DownloadAwsIamAuthClient(eksdRelease *eksdv1alpha1.Release) error {
	uri, err := getAwsIamAuthClientUri(eksdRelease, getKernelName())
	if err != nil {
		return fmt.Errorf("getting %s uri: %v", awsIamClientBinary, err)
	}
	err = files.GzipFileDownloadExtract(uri, awsIamClientBinary, "bin")
	if err != nil {
		return fmt.Errorf("downloading %s binary: %v", awsIamClientBinary, err)
	}
	return nil
}

func getKernelName() string {
	return runtime.GOOS
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
