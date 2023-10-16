package e2e

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/go-logr/logr"

	"github.com/aws/eks-anywhere/internal/pkg/api"
	"github.com/aws/eks-anywhere/internal/pkg/s3"
	"github.com/aws/eks-anywhere/internal/pkg/ssm"
	"github.com/aws/eks-anywhere/pkg/networkutils"
	e2etests "github.com/aws/eks-anywhere/test/framework"
)

var requiredFiles = []string{cliBinary, e2eBinary}

const (
	cliBinary                  = "eksctl-anywhere"
	e2eBinary                  = "e2e.test"
	bundlesReleaseManifestFile = "local-bundle-release.yaml"
	eksAComponentsManifestFile = "local-eksa-components.yaml"
	testNameFile               = "e2e-test-name"
	maxUserWatches             = 524288
	maxUserInstances           = 512
	key                        = "Integration-Test"
	tag                        = "EKSA-E2E"
)

type E2ESession struct {
	session             *session.Session
	instanceProfileName string
	storageBucket       string
	jobId               string
	instanceId          string
	ipPool              networkutils.IPPool
	testEnvVars         map[string]string
	bundlesOverride     bool
	cleanupVms          bool
	requiredFiles       []string
	branchName          string
	hardware            []*api.Hardware
	logger              logr.Logger
}

func newE2ESession(instanceId string, conf instanceRunConf) (*E2ESession, error) {
	e := &E2ESession{
		session:             conf.session,
		instanceId:          instanceId,
		instanceProfileName: conf.instanceProfileName,
		storageBucket:       conf.storageBucket,
		jobId:               conf.jobId,
		ipPool:              conf.ipPool,
		testEnvVars:         make(map[string]string),
		bundlesOverride:     conf.bundlesOverride,
		cleanupVms:          conf.cleanupVms,
		requiredFiles:       requiredFiles,
		branchName:          conf.branchName,
		hardware:            conf.hardware,
		logger:              conf.logger,
	}

	return e, nil
}

func (e *E2ESession) setup(regex string) error {
	err := e.uploadRequiredFiles()
	if err != nil {
		return err
	}

	e.logger.V(1).Info("Waiting until SSM is ready")
	err = ssm.WaitForSSMReady(e.session, e.instanceId, ssmTimeout)
	if err != nil {
		return fmt.Errorf("waiting for ssm in new instance: %v", err)
	}

	err = e.updateFSInotifyResources()
	if err != nil {
		return err
	}
	err = e.createTestNameFile(regex)
	if err != nil {
		return err
	}

	err = e.downloadRequiredFilesInInstance()
	if err != nil {
		return err
	}

	err = e.setupOIDC(regex)
	if err != nil {
		return err
	}

	err = e.setupEtcdEncryption(regex)
	if err != nil {
		return err
	}

	err = e.setupVSphereEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupTinkerbellEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupCloudStackEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupNutanixEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupSnowEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupFluxEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupFluxGitEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupProxyEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupRegistryMirrorEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupAwsIam(regex)
	if err != nil {
		return err
	}

	err = e.setupNTPEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupBottlerocketKubernetesSettingsEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupOOBEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupPackagesEnv(regex)
	if err != nil {
		return err
	}

	err = e.setupCertManagerEnv(regex)
	if err != nil {
		return err
	}

	ipPool := e.ipPool.ToString()
	if ipPool != "" {
		e.testEnvVars[e2etests.ClusterIPPoolEnvVar] = ipPool
	}

	// Adding JobId to Test Env variables
	e.testEnvVars[e2etests.JobIdVar] = e.jobId
	e.testEnvVars[e2etests.BundlesOverrideVar] = strconv.FormatBool(e.bundlesOverride)
	e.testEnvVars[e2etests.CleanupVmsVar] = strconv.FormatBool(e.cleanupVms)

	if e.branchName != "" {
		e.testEnvVars[e2etests.BranchNameEnvVar] = e.branchName
	}

	e.testEnvVars[e2etests.ClusterPrefixVar] = clusterPrefix(e.branchName, e.instanceId)
	return nil
}

func (e *E2ESession) updateFSInotifyResources() error {
	command := fmt.Sprintf("sudo sysctl fs.inotify.max_user_watches=%v && sudo sysctl fs.inotify.max_user_instances=%v", maxUserWatches, maxUserInstances)

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("updating fs inotify resources: %v", err)
	}
	e.logger.V(1).Info("Successfully updated the fs inotify user watches and instances")

	return nil
}

func (e *E2ESession) uploadRequiredFile(file string) error {
	uploadFile := fmt.Sprintf("bin/%s", file)
	key := fmt.Sprintf("%s/%s", e.jobId, file)
	e.logger.V(1).Info("Uploading file to s3 bucket", "file", file, "key", key)
	err := s3.UploadFile(e.session, uploadFile, key, e.storageBucket)
	if err != nil {
		return fmt.Errorf("uploading file [%s]: %v", file, err)
	}

	return nil
}

func (e *E2ESession) uploadRequiredFiles() error {
	if e.bundlesOverride {
		e.requiredFiles = append(e.requiredFiles, bundlesReleaseManifestFile)
		if _, err := os.Stat(fmt.Sprintf("bin/%s", eksAComponentsManifestFile)); err == nil {
			e.requiredFiles = append(e.requiredFiles, eksAComponentsManifestFile)
		} else if errors.Is(err, os.ErrNotExist) {
			e.logger.V(0).Info("WARNING: no components manifest override found, but bundle override is present. " +
				"If the EKS-A components have changed be sure to provide a components override!")
		} else {
			return err
		}
	}
	for _, file := range e.requiredFiles {
		err := e.uploadRequiredFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *E2ESession) downloadRequiredFileInInstance(file string) error {
	e.logger.V(1).Info("Downloading from s3 in instance", "file", file)

	command := fmt.Sprintf("aws s3 cp s3://%s/%s/%[3]s ./bin/ && chmod 645 ./bin/%[3]s", e.storageBucket, e.jobId, file)
	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("downloading file in instance: %v", err)
	}
	e.logger.V(1).Info("Successfully downloaded file")

	return nil
}

func (e *E2ESession) downloadRequiredFilesInInstance() error {
	for _, file := range e.requiredFiles {
		err := e.downloadRequiredFileInInstance(file)
		if err != nil {
			return err
		}
	}
	return nil
}

func (e *E2ESession) createTestNameFile(testName string) error {
	command := fmt.Sprintf("echo \"%s\" > %s", testName, testNameFile)

	if err := ssm.Run(e.session, logr.Discard(), e.instanceId, command, ssmTimeout); err != nil {
		return fmt.Errorf("creating test name file in instance: %v", err)
	}
	e.logger.V(1).Info("Successfully created test name file")

	return nil
}

func clusterPrefix(branch, instanceId string) (clusterPrefix string) {
	if branch == "" {
		return instanceId
	}
	forbiddenChars := []string{"."}
	sanitizedBranch := strings.ToLower(branch)
	for _, char := range forbiddenChars {
		sanitizedBranch = strings.ReplaceAll(sanitizedBranch, char, "-")
	}

	if len(sanitizedBranch) > 7 {
		sanitizedBranch = sanitizedBranch[:7]
	}

	if len(instanceId) > 7 {
		instanceId = instanceId[:7]
	}

	clusterPrefix = fmt.Sprintf("%s-%s", sanitizedBranch, instanceId)
	return clusterPrefix
}

func (e *E2ESession) clusterName(branch, instanceId, testName string) (clusterName string) {
	clusterName = fmt.Sprintf("%s-%s", clusterPrefix(branch, instanceId), e2etests.GetTestNameHash(testName))
	if len(clusterName) > 63 {
		e.logger.Info("Cluster name is longer than 63 characters; truncating to 63 characters.", "original cluster name", clusterName, "truncated cluster name", clusterName[:63])
		clusterName = clusterName[:63]
	}
	return clusterName
}
