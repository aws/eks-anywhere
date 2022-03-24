package artifacts

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
	"github.com/aws/eks-anywhere-test-tool/pkg/s3"
	"github.com/aws/eks-anywhere/pkg/logger"
	"github.com/aws/eks-anywhere/pkg/retrier"
)

type FetchArtifactsOpt func(options *fetchArtifactConfig) (err error)

func WithCodebuildBuild(buildId string) FetchArtifactsOpt {
	return func(options *fetchArtifactConfig) (err error) {
		options.buildId = buildId
		logger.Info("user provided build ID detected", "buildId", buildId)
		return err
	}
}

func WithCodebuildProject(project string) FetchArtifactsOpt {
	return func(options *fetchArtifactConfig) (err error) {
		options.project = project
		logger.Info("user provided project ID detected", "project", project)
		return err
	}
}

type fetchArtifactConfig struct {
	buildId string
	bucket  string
	project string
}

type testArtifactFetcher struct {
	testAccountS3Client         *s3.S3
	buildAccountCodebuildClient *codebuild.Codebuild
	writer                      filewriter.FileWriter
	retrier                     *retrier.Retrier
}

func New(testAccountS3Client *s3.S3, buildAccountCodebuildCient *codebuild.Codebuild, writer filewriter.FileWriter) *testArtifactFetcher {
	return &testArtifactFetcher{
		testAccountS3Client:         testAccountS3Client,
		buildAccountCodebuildClient: buildAccountCodebuildCient,
		writer:                      writer,
		retrier:                     fileWriterRetrier(),
	}
}

func (l *testArtifactFetcher) FetchArtifacts(opts ...FetchArtifactsOpt) error {
	config := &fetchArtifactConfig{
		bucket:  os.Getenv(constants.E2eArtifactsBucketEnvVar),
		project: constants.EksATestCodebuildProject,
	}

	for _, opt := range opts {
		err := opt(config)
		if err != nil {
			return fmt.Errorf("failed to set options on fetch artifacts config: %v", err)
		}
	}

	if config.buildId == "" {
		p, err := l.buildAccountCodebuildClient.FetchLatestBuildForProject(config.project)
		if err != nil {
			return fmt.Errorf("failed to get latest build for project: %v", err)
		}
		config.buildId = *p.Id
	}

	logger.Info("Fetching build artifacts...")

	objects, err := l.testAccountS3Client.ListObjects(config.bucket, config.buildId)
	logger.V(5).Info("Listed objects", "bucket", config.bucket, "prefix", config.buildId, "objects", len(objects))
	if err != nil {
		return fmt.Errorf("listing objects in bucket %s at key %s: %v", config.bucket, config.buildId, err)
	}

	errs, _ := errgroup.WithContext(context.Background())

	for _, object := range objects {
		if excludedKey(*object.Key) {
			continue
		}
		obj := *object
		errs.Go(func() error {
			logger.Info("Fetching object", "key", obj.Key, "bucket", config.bucket)
			o, err := l.testAccountS3Client.GetObject(config.bucket, *obj.Key)
			if err != nil {
				return err
			}
			logger.Info("Fetched object", "key", obj.Key, "bucket", config.bucket)

			logger.Info("Writing object to file", "key", obj.Key, "bucket", config.bucket)
			err = l.retrier.Retry(func() error {
				return l.writer.WriteS3KeyToFile(*obj.Key, o)
			})
			if err != nil {
				logger.Info("error occured while writing file", "err", err)
				return fmt.Errorf("writing object %s from bucket %s to file: %v", *obj.Key, config.bucket, err)
			}
			return nil
		})
	}
	return errs.Wait()
}

func excludedKey(key string) bool {
	excludedKeys := []string{
		"/.git/",
	}

	excludedSuffixes := []string{
		"/e2e.test",
		"/eksctl-anywhere",
	}

	for _, s := range excludedKeys {
		if strings.Contains(key, s) {
			return true
		}
	}

	for _, s := range excludedSuffixes {
		if strings.HasSuffix(key, s) {
			return true
		}
	}
	return false
}

func fileWriterRetrier() *retrier.Retrier {
	return retrier.New(time.Minute, retrier.WithRetryPolicy(func(totalRetries int, err error) (retry bool, wait time.Duration) {
		generator := rand.New(rand.NewSource(time.Now().UnixNano()))
		minWait := 1
		maxWait := 5
		waitWithJitter := time.Duration(generator.Intn(maxWait-minWait)+minWait) * time.Second
		if isTooManyOpenFilesError(err) && totalRetries < 15 {
			logger.V(2).Info("Too many files open, retrying")
			return true, waitWithJitter
		}
		return false, 0
	}))
}

func isTooManyOpenFilesError(err error) bool {
	return strings.Contains(err.Error(), "too many open files")
}
