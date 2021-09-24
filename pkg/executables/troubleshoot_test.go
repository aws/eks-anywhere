package executables_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/golang/mock/gomock"

	"github.com/aws/eks-anywhere/pkg/executables"
	mockexecutables "github.com/aws/eks-anywhere/pkg/executables/mocks"
	"github.com/aws/eks-anywhere/pkg/types"
)

const (
	archivePath     = "support-bundle-2021-09-17T16_22_45.tar.gz"
	bundlePath      = "./testBundleThatDoesNotExist.yaml"
	sinceTimeString = "2021-09-17T16:22:45Z"
)

func TestTroubleshootCollectSuccess(t *testing.T) {
	ts, ctx, cluster, e := newTroubleshoot(t)
	sinceTime, err := time.Parse(time.RFC3339, sinceTimeString)
	if err != nil {
		t.Errorf("Troubleshoot.Collect() error: failed to parse time: %v", err)
	}
	expectedParams := []string{bundlePath, "--kubeconfig", cluster.KubeconfigFile, "--interactive=false", "--since-time", sinceTimeString}
	returnBuffer := bytes.Buffer{}
	returnBuffer.Write([]byte(archivePath))
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParams)).Return(returnBuffer, nil)
	if _, err := ts.Collect(ctx, bundlePath, &sinceTime, cluster.KubeconfigFile); err != nil {
		t.Errorf("Troubleshoot.Collect() error = %v, want nil", err)
	}
}

func TestTroubleshootAnalyzeSuccess(t *testing.T) {
	ts, ctx, _, e := newTroubleshoot(t)
	var returnValues []*executables.SupportBundleAnalysis
	returnValues = append(returnValues, &executables.SupportBundleAnalysis{})
	returnJson, err := json.Marshal(returnValues)
	if err != nil {
		return
	}
	returnBuffer := bytes.Buffer{}
	returnBuffer.Write(returnJson)

	expectedParams := []string{"analyze", bundlePath, "--bundle", archivePath, "--output", "json"}
	e.EXPECT().Execute(ctx, gomock.Eq(expectedParams)).Return(returnBuffer, nil)
	if _, err := ts.Analyze(ctx, bundlePath, archivePath); err != nil {
		t.Errorf("Troubleshoot.Analyze() error = %v, want nil", err)
	}
}

func newTroubleshoot(t *testing.T) (*executables.Troubleshoot, context.Context, *types.Cluster, *mockexecutables.MockExecutable) {
	kubeconfigFile := "c.kubeconfig"
	cluster := &types.Cluster{
		KubeconfigFile: kubeconfigFile,
		Name:           "test-cluster",
	}

	ctx := context.Background()
	ctrl := gomock.NewController(t)
	executable := mockexecutables.NewMockExecutable(ctrl)

	return executables.NewTroubleshoot(executable), ctx, cluster, executable
}
