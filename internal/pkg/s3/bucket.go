package s3

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
)

func GetBucketPublicURL(session *session.Session, bucket string) string {
	return fmt.Sprintf("https://s3-%s.amazonaws.com/%s", *session.Config.Region, bucket)
}
