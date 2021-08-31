package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Download(session *session.Session, key, bucket string) ([]byte, error) {
	d := s3manager.NewDownloader(session)
	b := aws.NewWriteAtBuffer([]byte{})
	_, err := d.Download(b, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, fmt.Errorf("error downloading [%s] from [%s] bucket: %v", key, bucket, err)
	}

	return b.Bytes(), nil
}
