package s3

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/eks-anywhere/pkg/logger"
)

func CleanUpS3Bucket(session *session.Session, bucket string, maxAge float64) error {
	service := s3.New(session)

	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	}
	result, err := service.ListObjectsV2(input)
	if err != nil {
		return fmt.Errorf("listing s3 bucket objects: %v", err)
	}

	var objectList []*string
	logger.V(4).Info("Listing s3 bucket objects")
	for _, object := range result.Contents {
		logger.V(4).Info("s3", "object_name", *(object.Key), "last_modified", *(object.LastModified))
		lastModifiedTime := time.Since(*(object.LastModified)).Seconds()
		if lastModifiedTime > maxAge && *(object.Key) != "eksctl/eksctl" && *(object.Key) != "generated-artifacts/" {
			logger.V(4).Info("Adding object for deletion")
			objectList = append(objectList, object.Key)
		} else {
			logger.V(4).Info("Skipping object deletion")
		}
	}

	if len(objectList) != 0 {
		for _, object := range objectList {
			logger.V(4).Info("Deleting", "object", *object)
			_, err = service.DeleteObject(
				&s3.DeleteObjectInput{
					Bucket: aws.String(bucket),
					Key:    object,
				},
			)
			if err != nil {
				return fmt.Errorf("deleting object %s: %v", *object, err)
			}
		}
	} else {
		logger.V(4).Info("No s3 objects for deletion")
	}

	return nil
}
