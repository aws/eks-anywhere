package s3

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type S3 struct {
	session *session.Session
	svc     *s3.S3
	retrier *retrier.Retrier
}

func New(account awsprofiles.EksAccount) (*S3, error) {
	logger.V(2).Info("creating S3 client")
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: account.ProfileName(),
		Config: aws.Config{
			Region:                        aws.String(constants.AwsAccountRegion),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	})
	if err != nil {
		fmt.Printf("Got error when setting up session: %v", err)
		os.Exit(1)
	}

	svc := s3.New(sess)
	logger.V(2).Info("created S3 client")

	return &S3{
		session: sess,
		svc:     svc,
		retrier: getObjectRetirer(),
	}, nil
}

func (s *S3) ListObjects(bucket string, prefix string) (listedObjects []*s3.Object, err error) {
	var nextToken *string
	var objects []*s3.Object

	input := &s3.ListObjectsV2Input{
		Bucket:            aws.String(bucket),
		Prefix:            aws.String(prefix),
		ContinuationToken: nextToken,
	}

	for {
		l, err := s.svc.ListObjectsV2(input)
		if err != nil {
			return nil, fmt.Errorf("failed to list objects: %v", err)
		}
		objects = append(objects, l.Contents...)
		if !aws.BoolValue(l.IsTruncated) {
			logger.Info("finished fetching objects", "bucket", bucket, "prefix", prefix)
			logger.V(3).Info("token comparison", "nextToken", nextToken, "nextContinuatonToken", l.NextContinuationToken)
			break
		}
		nextToken = l.NextContinuationToken
		input.ContinuationToken = nextToken
		logger.Info("fetched objects", "bucket", bucket, "prefix", prefix, "events", len(l.Contents))
		logger.V(3).Info("token comparison", "nextToken", nextToken, "nextContinuatonToken", l.NextContinuationToken)
	}
	return objects, nil
}

func (s *S3) GetObject(bucket string, key string) ([]byte, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	var obj *s3.GetObjectOutput
	var err error
	err = s.retrier.Retry(func() error {
		obj, err = s.svc.GetObject(input)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get object at key %s: %v", key, err)
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(obj.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read object at key %s: %v", key, err)
	}
	return buf.Bytes(), nil
}

func getObjectRetirer() *retrier.Retrier {
	return retrier.New(time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		generator := rand.New(rand.NewSource(time.Now().UnixNano()))
		minWait := 1
		maxWait := 5
		waitWithJitter := time.Duration(generator.Intn(maxWait-minWait)+minWait) * time.Second
		if isThrottledError(err) && totalRetries < 15 {
			logger.V(2).Info("Throttled by S3, retrying")
			return true, waitWithJitter
		}
		return false, 0
	}))
}

func isThrottledError(err error) bool {
	return strings.Contains(err.Error(), "no such host")
}
