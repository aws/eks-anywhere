package s3

import (
	"fmt"
	"io"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func Download(session *session.Session, key, bucket string) ([]byte, error) {
	b := aws.NewWriteAtBuffer([]byte{})
	if err := download(session, key, bucket, b); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func download(session *session.Session, key, bucket string, w io.WriterAt) error {
	d := s3manager.NewDownloader(session)
	_, err := d.Download(w, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("downloading [%s] from [%s] bucket: %v", key, bucket, err)
	}

	return nil
}

func DownloadToDisk(session *session.Session, key, bucket, dst string) error {
	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("unable to create destination file %s: %v", dst, err)
	}
	defer file.Close()

	if err := download(session, key, bucket, file); err != nil {
		return err
	}

	return nil
}
