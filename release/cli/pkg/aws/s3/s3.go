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

package s3

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

// Read reads the content of an object from an S3 bucket.
// It returns an io.ReadCloser that should be closed after use.
func Read(bucket, key string) (io.ReadCloser, error) {
	objectURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)
	resp, err := http.Get(objectURL)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("download file returned code %s", resp.Status)
	}
	return resp.Body, nil
}

func DownloadFile(filePath, bucket, key string, s3Downloader *s3manager.Downloader, private bool) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return errors.Cause(err)
	}

	fd, err := os.Create(filePath)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()

	if private {
		_, err = s3Downloader.Download(fd, &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			return err
		}
	} else {
		body, err := Read(bucket, key)
		if err != nil {
			return err
		}

		defer body.Close()

		if _, err = io.Copy(fd, body); err != nil {
			return err
		}
	}

	return nil
}

func UploadFile(filePath string, bucket, key *string, s3Uploader *s3manager.Uploader, private bool) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()

	objectCannedACL := s3.ObjectCannedACLPublicRead
	if private {
		objectCannedACL = s3.ObjectCannedACLPrivate
	}
	result, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    key,
		Body:   fd,
		ACL:    aws.String(objectCannedACL),
	})
	if err != nil {
		return errors.Cause(err)
	}

	fmt.Printf("Artifact file uploaded to %s\n", result.Location)
	return nil
}

func KeyExists(s3Client *s3.S3, bucket, key string, private bool) (bool, error) {
	if private {
		_, err := s3Client.HeadObject(&s3.HeadObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
		if err != nil {
			if aerr, ok := err.(awserr.Error); ok && aerr.Code() == "NotFound" {
				return false, nil
			}
			return false, fmt.Errorf("calling S3 HeadObject API to check if object is present: %v", err)
		}
	} else {
		objectUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)

		resp, err := http.Head(objectUrl)
		if err != nil {
			return false, fmt.Errorf("making HTTP HEAD request to check if object is present: %v", err)
		}

		if resp.StatusCode != http.StatusOK {
			return false, nil
		}
	}

	return true, nil
}
