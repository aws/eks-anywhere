package artifacts

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	cb "github.com/aws/aws-sdk-go/service/codebuild"
	"golang.org/x/sync/errgroup"

	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere-test-tool/pkg/filewriter"
	"github.com/aws/eks-anywhere-test-tool/pkg/s3"
	"github.com/aws/eks-anywhere-test-tool/pkg/testresults"
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

func WithAllArtifacts() FetchArtifactsOpt {
	return func(options *fetchArtifactConfig) (err error) {
		options.fetchAll = true
		return err
	}
}

type fetchArtifactConfig struct {
	buildId  string
	bucket   string
	project  string
	fetchAll bool
}

type testArtifactFetcher struct {
	testAccountS3Client         *s3.S3
	buildAccountCodebuildClient *codebuild.Codebuild
	buildAccountCwClient        *cloudwatch.Cloudwatch
	writer                      filewriter.FileWriter
	retrier                     *retrier.Retrier
}

func New(testAccountS3Client *s3.S3, buildAccountCodebuildCient *codebuild.Codebuild, writer filewriter.FileWriter, cwClient *cloudwatch.Cloudwatch) *testArtifactFetcher {
	return &testArtifactFetcher{
		testAccountS3Client:         testAccountS3Client,
		buildAccountCodebuildClient: buildAccountCodebuildCient,
		writer:                      writer,
		retrier:                     fileWriterRetrier(),
		buildAccountCwClient:        cwClient,
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

	var p *cb.Build
	var err error

	if config.buildId == "" {
		p, err = l.buildAccountCodebuildClient.FetchLatestBuildForProject(config.project)
		if err != nil {
			return fmt.Errorf("failed to get latest build for project: %v", err)
		}
		config.buildId = *p.Id

	} else {
		p, err = l.buildAccountCodebuildClient.FetchBuildForProject(config.buildId)
		if err != nil {
			return fmt.Errorf("failed to get build for project: %v", err)
		}
	}

	g := p.Logs.GroupName
	s := p.Logs.StreamName

	logs, err := l.buildAccountCwClient.GetLogs(*g, *s)
	if err != nil {
		return fmt.Errorf("fetching cloudwatch logs: %v", err)
	}

	_, failedTests, err := testresults.GetFailedTests(logs)
	if err != nil {
		return err
	}
	failedTestIds := testresults.TestResultsJobIdMap(failedTests)

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
		keySplit := strings.Split(*obj.Key, "/")

		_, ok := failedTestIds[keySplit[0]]
		if !ok && !config.fetchAll {
			continue
		}

		errs.Go(func() error {
			logger.Info("Fetching object", "key", obj.Key, "bucket", config.bucket)
			o, err := l.testAccountS3Client.GetObject(config.bucket, *obj.Key)
			if err != nil {
				return err
			}
			logger.Info("Fetched object", "key", obj.Key, "bucket", config.bucket)

			logger.Info("Writing object to file", "key", obj.Key, "bucket", config.bucket)
			err = l.retrier.Retry(func() error {
				return l.writer.WriteTestArtifactsS3ToFile(*obj.Key, o)
			})
			if err != nil {
				logger.Info("error occurred while writing file", "err", err)
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
		"/oidc/",
	}

	excludedSuffixes := []string{
		"/e2e.test",
		"/eksctl-anywhere",
		".csv",
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
