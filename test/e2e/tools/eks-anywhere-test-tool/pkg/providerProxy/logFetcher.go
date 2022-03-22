package providerProxy

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere-test-tool/pkg/cloudwatch"
	"github.com/aws/eks-anywhere-test-tool/pkg/codebuild"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type FetchSessionOpts func(options *fetchSessionsConfig) (err error)

func WithCodebuildBuild(buildId string) FetchSessionOpts {
	return func(options *fetchSessionsConfig) (err error) {
		options.buildId = buildId
		logger.Info("user provided build ID detected", "buildId", buildId)
		return err
	}
}

func WithCodebuildProject(project string) FetchSessionOpts {
	return func(options *fetchSessionsConfig) (err error) {
		options.project = project
		logger.Info("user provided project ID detected", "project", project)
		return err
	}
}

type fetchSessionsConfig struct {
	buildId string
	project string
}

type (
	requestFilter   func(logs []*cloudwatchlogs.OutputLogEvent) (filteredLogs []*cloudwatchlogs.OutputLogEvent, err error)
	requestConsumer func(logs []*cloudwatchlogs.OutputLogEvent) error
)

type ProxyFetcherOpt func(*proxyLogFetcher)

func WithLogStdout() ProxyFetcherOpt {
	return func(l *proxyLogFetcher) {
		l.processRequests = func(logs []*cloudwatchlogs.OutputLogEvent) error { return nil }
	}
}

type proxyLogFetcher struct {
	buildAccountCwClient        *cloudwatch.Cloudwatch
	testAccountCwClient         *cloudwatch.Cloudwatch
	buildAccountCodebuildClient *codebuild.Codebuild
	writer                      *requestWriter
	filterRequests              requestFilter
	processRequests             requestConsumer
}

func New(buildAccountCwClient *cloudwatch.Cloudwatch, testAccountCwClient *cloudwatch.Cloudwatch, buildAccountCodebuildClient *codebuild.Codebuild, opts ...ProxyFetcherOpt) *proxyLogFetcher {
	l := &proxyLogFetcher{
		buildAccountCwClient:        buildAccountCwClient,
		testAccountCwClient:         testAccountCwClient,
		buildAccountCodebuildClient: buildAccountCodebuildClient,
	}
	for _, o := range opts {
		o(l)
	}

	defaultOutputFolder := fmt.Sprintf("provider-proxy-logs-%s", time.Now().Format(time.RFC3339))

	if l.filterRequests == nil {
		l.filterRequests = noFilter
	}

	if l.processRequests == nil {
		_ = l.ensureWriter(defaultOutputFolder)
		l.processRequests = l.writer.writeRequest
	}

	return l
}

func (l *proxyLogFetcher) FetchProviderProxyLogs(opts ...FetchSessionOpts) error {
	config := &fetchSessionsConfig{
		project: constants.EksATestCodebuildProject,
	}

	for _, opt := range opts {
		err := opt(config)
		if err != nil {
			return fmt.Errorf("failed to set options on fetch logs config: %v", err)
		}
	}

	if config.buildId == "" {
		p, err := l.buildAccountCodebuildClient.FetchLatestBuildForProject(config.project)
		if err != nil {
			return fmt.Errorf("failed to get latest build for project: %v", err)
		}
		config.buildId = *p.Id
		logger.Info("Using latest build for selected project", "buildID", config.buildId, "project", config.project)
	}

	logs, err := l.FetchProviderProxyLogsForbuild(config.project, config.buildId)
	if err != nil {
		return err
	}
	err = l.processRequests(logs)
	if err != nil {
		return err
	}
	return nil
}

func (l *proxyLogFetcher) FetchProviderProxyLogsForbuild(project string, buildId string) ([]*cloudwatchlogs.OutputLogEvent, error) {
	logger.Info("Fetching provider proxy logs...")
	build, err := l.buildAccountCodebuildClient.FetchBuildForProject(buildId)
	if err != nil {
		return nil, fmt.Errorf("error fetching build for project %s: %v", project, err)
	}

	buildStart := build.StartTime.UnixNano() / 1e6
	logger.Info("Starting log time", "Start time", buildStart)

	buildEnd := build.EndTime.UnixNano() / 1e6
	logger.Info("Ending log time", "Start time", buildEnd)

	logs, err := l.buildAccountCwClient.GetLogsInTimeframe(constants.CiProxyLogGroup, constants.CiProxyLogStream, buildStart, buildEnd)
	if err != nil {
		return nil, fmt.Errorf("error when fetching cloudwatch logs: %v", err)
	}
	filteredLogs, err := l.filterRequests(logs)
	return filteredLogs, err
}

func (l *proxyLogFetcher) ensureWriter(folderPath string) error {
	if l.writer != nil {
		return nil
	}

	var err error
	l.writer, err = newRequestWriter(folderPath)
	if err != nil {
		return err
	}

	return nil
}

func noFilter(logs []*cloudwatchlogs.OutputLogEvent) (outputLogs []*cloudwatchlogs.OutputLogEvent, err error) {
	return logs, nil
}
