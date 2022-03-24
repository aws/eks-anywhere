package s3

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type UploadOpt func(*s3manager.UploadInput)

func WithPublicRead() UploadOpt {
	return func(s *s3manager.UploadInput) {
		s.ACL = aws.String("public-read")
	}
}

func UploadFile(session *session.Session, file, key, bucket string, opts ...UploadOpt) error {
	fileBody, err := os.Open(file)
	if err != nil {
		return fmt.Errorf("opening file for upload: %v", err)
	}

	return upload(session, fileBody, key, bucket, opts...)
}

func Upload(session *session.Session, body []byte, key, bucket string, opts ...UploadOpt) error {
	return upload(session, bytes.NewBuffer(body), key, bucket, opts...)
}

func upload(session *session.Session, body io.Reader, key, bucket string, opts ...UploadOpt) error {
	s3Uploader := s3manager.NewUploader(session)
	i := &s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   body,
	}

	for _, opt := range opts {
		opt(i)
	}

	_, err := s3Uploader.Upload(i)
	if err != nil {
		return fmt.Errorf("uploading to s3: %v", err)
	}

	return nil
}
