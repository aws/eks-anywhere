// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	eksdv1alpha1 "github.com/aws/eks-distro-build-tooling/release/api/v1alpha1"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"
	"sigs.k8s.io/yaml"
)

type SourceClients struct {
	Docker *DockerClients
	S3     *SourceS3Clients
}

type ReleaseClients struct {
	Docker    *DockerClients
	S3        *ReleaseS3Clients
	ECRPublic *ReleaseECRPublicClient
}

type DockerClients struct {
	Client     *docker.Client
	AuthConfig *docker.AuthConfiguration
}

type SourceS3Clients struct {
	Client     *s3.S3
	Downloader *s3manager.Downloader
}

type ReleaseS3Clients struct {
	Client   *s3.S3
	Uploader *s3manager.Uploader
}

type ReleaseECRPublicClient struct {
	Client *ecrpublic.ECRPublic
}

// Function to create release clients for dev release
func (r *ReleaseConfig) CreateDevReleaseClients() (*SourceClients, *ReleaseClients, error) {
	fmt.Println("Creating new dev release clients for S3, docker and ECR public")

	// PDX session for eks-a-build-prod-pdx
	pdxSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// IAD session for eks-a-build-prod-pdx
	iadSession, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// S3 client and managers
	s3Client := s3.New(pdxSession)
	downloader := s3manager.NewDownloader(pdxSession)
	uploader := s3manager.NewUploader(pdxSession)

	// Create new docker client
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get source ECR auth config
	fmt.Printf("Source container registry is: %s", r.SourceContainerRegistry)
	ecrClient := ecr.New(pdxSession)
	sourceAuthConfig, err := getEcrAuthConfig(ecrClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	fmt.Printf("Release container registry is: %s", r.ReleaseContainerRegistry)
	ecrPublicClient := ecrpublic.New(iadSession)
	releaseAuthConfig, err := getEcrPublicAuthConfig(ecrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing source clients
	sourceClients := &SourceClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: sourceAuthConfig,
		},
		S3: &SourceS3Clients{
			Client:     s3Client,
			Downloader: downloader,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: releaseAuthConfig,
		},
		S3: &ReleaseS3Clients{
			Client:   s3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client: ecrPublicClient,
		},
	}

	return sourceClients, releaseClients, nil
}

// Function to create clients for staging release
func (r *ReleaseConfig) CreateStagingReleaseClients() (*SourceClients, *ReleaseClients, error) {
	fmt.Println("Creating new staging release clients for S3, docker and ECR public")

	// Session for eks-a-build-prod-pdx
	sourceSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-west-2"),
		},
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Session for eks-a-artifact-beta-iad
	releaseSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
		Profile: "artifacts-staging",
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Source S3 client and downloader
	sourceS3Client := s3.New(sourceSession)
	downloader := s3manager.NewDownloader(sourceSession)

	// Release S3 client and uploader
	releaseS3Client := s3.New(releaseSession)
	uploader := s3manager.NewUploader(releaseSession)

	// Create new docker client
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get source ECR auth config
	fmt.Printf("Source container registry is: %s", r.SourceContainerRegistry)
	ecrClient := ecr.New(sourceSession)
	sourceAuthConfig, err := getEcrAuthConfig(ecrClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	fmt.Printf("Release container registry is: %s", r.ReleaseContainerRegistry)
	ecrPublicClient := ecrpublic.New(releaseSession)
	releaseAuthConfig, err := getEcrPublicAuthConfig(ecrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing source clients
	sourceClients := &SourceClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: sourceAuthConfig,
		},
		S3: &SourceS3Clients{
			Client:     sourceS3Client,
			Downloader: downloader,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: releaseAuthConfig,
		},
		S3: &ReleaseS3Clients{
			Client:   releaseS3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client: ecrPublicClient,
		},
	}

	return sourceClients, releaseClients, nil
}

// Function to create clients for production release
func (r *ReleaseConfig) CreateProdReleaseClients() (*SourceClients, *ReleaseClients, error) {
	fmt.Println("Creating new production release clients for S3, docker and ECR public")

	// Session for eks-a-artifact-beta-iad
	sourceSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
		Profile: "artifacts-staging",
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Session for eks-a-artifact-prod-iad
	releaseSession, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String("us-east-1"),
		},
		Profile: "artifacts-production",
	})
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Source S3 client and downloader
	sourceS3Client := s3.New(sourceSession)
	downloader := s3manager.NewDownloader(sourceSession)

	// Release S3 client and uploader
	releaseS3Client := s3.New(releaseSession)
	uploader := s3manager.NewUploader(releaseSession)

	// Create new docker client
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get source ECR Public auth config
	fmt.Printf("Source container registry is: %s", r.SourceContainerRegistry)
	sourceEcrPublicClient := ecrpublic.New(sourceSession)
	sourceAuthConfig, err := getEcrPublicAuthConfig(sourceEcrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	fmt.Printf("Release container registry is: %s", r.SourceContainerRegistry)
	releaseEcrPublicClient := ecrpublic.New(releaseSession)
	releaseAuthConfig, err := getEcrPublicAuthConfig(releaseEcrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing release clients
	sourceClients := &SourceClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: sourceAuthConfig,
		},
		S3: &SourceS3Clients{
			Client:     sourceS3Client,
			Downloader: downloader,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		Docker: &DockerClients{
			Client:     dockerClient,
			AuthConfig: releaseAuthConfig,
		},
		S3: &ReleaseS3Clients{
			Client:   releaseS3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client: releaseEcrPublicClient,
		},
	}

	return sourceClients, releaseClients, nil
}

// Function to retrieve auth configuration to authenticate with ECR registry
func getEcrAuthConfig(ecrClient *ecr.ECR) (*docker.AuthConfiguration, error) {
	// Get ECR authorization token
	authTokenOutput, err := ecrClient.GetAuthorizationToken(&ecr.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, errors.Cause(err)
	}
	authToken := *authTokenOutput.AuthorizationData[0].AuthorizationToken

	// Decode authorization token to get credential pair
	creds, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, errors.Cause(err)
	}

	// Get password from credential pair
	credsSplit := strings.Split(string(creds), ":")
	password := credsSplit[1]

	// Construct docker auth configuration
	authConfig := &docker.AuthConfiguration{
		Username: "AWS",
		Password: password,
	}

	return authConfig, nil
}

// Function to retrieve auth configuration to authenticate with ECR Public registry
func getEcrPublicAuthConfig(ecrPublicClient *ecrpublic.ECRPublic) (*docker.AuthConfiguration, error) {
	// Get ECR Public authorization token
	authTokenOutput, err := ecrPublicClient.GetAuthorizationToken(&ecrpublic.GetAuthorizationTokenInput{})
	if err != nil {
		return nil, errors.Cause(err)
	}
	authToken := *authTokenOutput.AuthorizationData.AuthorizationToken

	// Decode authorization token to get credential pair
	creds, err := base64.StdEncoding.DecodeString(authToken)
	if err != nil {
		return nil, errors.Cause(err)
	}

	// Get password from credential pair
	credsSplit := strings.Split(string(creds), ":")
	password := credsSplit[1]

	// Construct docker auth configuration
	authConfig := &docker.AuthConfiguration{
		Username: "AWS",
		Password: password,
	}

	return authConfig, nil
}

func getEksdRelease(eksdReleaseURL string) (*eksdv1alpha1.Release, error) {
	content, err := readHttpFile(eksdReleaseURL)
	if err != nil {
		return nil, err
	}

	eksd := &eksdv1alpha1.Release{}
	if err = yaml.UnmarshalStrict(content, eksd); err != nil {
		return nil, errors.Wrapf(err, "failed to unmarshal eksd manifest")
	}

	return eksd, nil
}

func readHttpFile(uri string) ([]byte, error) {
	fmt.Printf("Downloading %s\n", uri)
	resp, err := http.Get(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading file from url [%s]", uri)
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "failed reading file from url [%s]", uri)
	}

	return data, nil
}
