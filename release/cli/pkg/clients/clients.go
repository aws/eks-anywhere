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

package clients

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	ecrsdk "github.com/aws/aws-sdk-go/service/ecr"
	ecrpublicsdk "github.com/aws/aws-sdk-go/service/ecrpublic"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/pkg/errors"

	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecr"
	"github.com/aws/eks-anywhere/release/cli/pkg/aws/ecrpublic"
)

type SourceClients struct {
	S3  *SourceS3Clients
	ECR *SourceECRClient
}

type ReleaseClients struct {
	S3        *ReleaseS3Clients
	ECRPublic *ReleaseECRPublicClient
}

type SourceS3Clients struct {
	Client *s3.S3
}

type ReleaseS3Clients struct {
	Client   *s3.S3
	Uploader *s3manager.Uploader
}

type SourceECRClient struct {
	EcrClient       *ecrsdk.ECR
	EcrPublicClient *ecrpublicsdk.ECRPublic
	AuthConfig      *docker.AuthConfiguration
}

type ReleaseECRPublicClient struct {
	Client     *ecrpublicsdk.ECRPublic
	AuthConfig *docker.AuthConfiguration
}

// Function to create release clients for dev release.
func CreateDevReleaseClients(dryRun bool) (*SourceClients, *ReleaseClients, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("                 Dev Release Clients Creation")
	fmt.Println("==========================================================")
	if dryRun {
		fmt.Println("Skipping clients creation in dry-run mode")
		return nil, nil, nil
	}

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

	// S3 client and uploader
	s3Client := s3.New(pdxSession)
	uploader := s3manager.NewUploader(pdxSession)

	// Get source ECR auth config
	ecrClient := ecrsdk.New(pdxSession)
	sourceAuthConfig, err := ecr.GetAuthConfig(ecrClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	ecrPublicClient := ecrpublicsdk.New(iadSession)
	releaseAuthConfig, err := ecrpublic.GetAuthConfig(ecrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing source clients
	sourceClients := &SourceClients{
		S3: &SourceS3Clients{
			Client: s3Client,
		},
		ECR: &SourceECRClient{
			EcrClient:  ecrClient,
			AuthConfig: sourceAuthConfig,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		S3: &ReleaseS3Clients{
			Client:   s3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client:     ecrPublicClient,
			AuthConfig: releaseAuthConfig,
		},
	}

	return sourceClients, releaseClients, nil
}

// Function to create clients for staging release.
func CreateStagingReleaseClients() (*SourceClients, *ReleaseClients, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("              Staging Release Clients Creation")
	fmt.Println("==========================================================")

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

	// Source S3 client
	sourceS3Client := s3.New(sourceSession)

	// Release S3 client and uploader
	releaseS3Client := s3.New(releaseSession)
	uploader := s3manager.NewUploader(releaseSession)

	// Get source ECR auth config
	ecrClient := ecrsdk.New(sourceSession)
	sourceAuthConfig, err := ecr.GetAuthConfig(ecrClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	ecrPublicClient := ecrpublicsdk.New(releaseSession)
	releaseAuthConfig, err := ecrpublic.GetAuthConfig(ecrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing source clients
	sourceClients := &SourceClients{
		S3: &SourceS3Clients{
			Client: sourceS3Client,
		},
		ECR: &SourceECRClient{
			EcrClient:  ecrClient,
			AuthConfig: sourceAuthConfig,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		S3: &ReleaseS3Clients{
			Client:   releaseS3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client:     ecrPublicClient,
			AuthConfig: releaseAuthConfig,
		},
	}

	return sourceClients, releaseClients, nil
}

// Function to create clients for production release.
func CreateProdReleaseClients() (*SourceClients, *ReleaseClients, error) {
	fmt.Println("\n==========================================================")
	fmt.Println("             Production Release Clients Creation")
	fmt.Println("==========================================================")

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

	// Source S3 client
	sourceS3Client := s3.New(sourceSession)

	// Release S3 client and uploader
	releaseS3Client := s3.New(releaseSession)
	uploader := s3manager.NewUploader(releaseSession)

	// Get source ECR Public auth config
	sourceEcrPublicClient := ecrpublicsdk.New(sourceSession)
	sourceAuthConfig, err := ecrpublic.GetAuthConfig(sourceEcrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Get release ECR Public auth config
	releaseEcrPublicClient := ecrpublicsdk.New(releaseSession)
	releaseAuthConfig, err := ecrpublic.GetAuthConfig(releaseEcrPublicClient)
	if err != nil {
		return nil, nil, errors.Cause(err)
	}

	// Constructing release clients
	sourceClients := &SourceClients{
		S3: &SourceS3Clients{
			Client: sourceS3Client,
		},
		ECR: &SourceECRClient{
			EcrPublicClient: sourceEcrPublicClient,
			AuthConfig:      sourceAuthConfig,
		},
	}

	// Constructing release clients
	releaseClients := &ReleaseClients{
		S3: &ReleaseS3Clients{
			Client:   releaseS3Client,
			Uploader: uploader,
		},
		ECRPublic: &ReleaseECRPublicClient{
			Client:     releaseEcrPublicClient,
			AuthConfig: releaseAuthConfig,
		},
	}

	return sourceClients, releaseClients, nil
}
