package codebuild

import (
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/codebuild"

	"github.com/aws/eks-anywhere-test-tool/pkg/awsprofiles"
	"github.com/aws/eks-anywhere-test-tool/pkg/constants"
	"github.com/aws/eks-anywhere/pkg/logger"
)

type Codebuild struct {
	session *session.Session
	svc     *codebuild.CodeBuild
}

func New(account awsprofiles.EksAccount) (*Codebuild, error) {
	logger.V(2).Info("creating codebuild client")
	sess, err := session.NewSessionWithOptions(session.Options{
		Profile: account.ProfileName(),
		Config: aws.Config{
			Region:                        aws.String(constants.AwsAccountRegion),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("got error when setting up session: %v", err)
	}

	svc := codebuild.New(sess)
	logger.V(2).Info("created codebuild client")

	return &Codebuild{
		session: sess,
		svc:     svc,
	}, nil
}

func (c *Codebuild) FetchBuildForProject(id string) (*codebuild.Build, error) {
	return c.getBuildById(id)
}

func (c *Codebuild) FetchLatestBuildForProject(project string) (*codebuild.Build, error) {
	builds := c.FetchBuildsForProject(project)

	// Find latest build that is not in progress
	for _, id := range builds.Ids {
		b, err := c.getBuildById(*id)
		if err != nil {
			return nil, err
		}
		if *b.BuildStatus != codebuild.StatusTypeInProgress {
			return b, nil
		}
	}

	return nil, errors.New("can't find a build for project that has already finished")
}

func (c *Codebuild) getBuildById(id string) (*codebuild.Build, error) {
	i := []*string{aws.String(id)}
	latestBuild, err := c.svc.BatchGetBuilds(&codebuild.BatchGetBuildsInput{Ids: i})
	if err != nil {
		return nil, fmt.Errorf("got an error when fetching latest build for project: %v", err)
	}
	if len(latestBuild.Builds) < 1 {
		return nil, fmt.Errorf("no builds found with id %s", id)
	}
	return latestBuild.Builds[0], nil
}

func (c *Codebuild) FetchBuildsForProject(project string) *codebuild.ListBuildsForProjectOutput {
	// we're using this to get the latest build, so we don't care about pagination at the moment
	builds, err := c.svc.ListBuildsForProject(&codebuild.ListBuildsForProjectInput{
		NextToken:   nil,
		ProjectName: aws.String(project),
		SortOrder:   aws.String(codebuild.SortOrderTypeDescending),
	})
	if err != nil {
		fmt.Printf("Got an error when fetching builds for project: %v", err)
		os.Exit(1)
	}
	return builds
}
