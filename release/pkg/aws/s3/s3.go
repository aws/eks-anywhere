package s3

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
)

func DownloadFile(filePath, bucket, key string) error {
	objectURL := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)

	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return errors.Cause(err)
	}

	fd, err := os.Create(filePath)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()

	// Get the data
	resp, err := http.Get(objectURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(fd, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func UploadFile(filePath string, bucket, key *string, s3Uploader *s3manager.Uploader) error {
	fd, err := os.Open(filePath)
	if err != nil {
		return errors.Cause(err)
	}
	defer fd.Close()

	result, err := s3Uploader.Upload(&s3manager.UploadInput{
		Bucket: bucket,
		Key:    key,
		Body:   fd,
		ACL:    aws.String("public-read"),
	})
	if err != nil {
		return errors.Cause(err)
	}

	fmt.Printf("Artifact file uploaded to %s\n", result.Location)
	return nil
}

func KeyExists(bucket string, key string) bool {
	objectUrl := fmt.Sprintf("https://%s.s3.amazonaws.com/%s", bucket, key)

	resp, err := http.Head(objectUrl)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}

	return true
}
