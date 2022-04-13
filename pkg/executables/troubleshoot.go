package executables

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"time"
)

const (
	troubleshootPath          = "support-bundle"
	supportBundleArchiveRegex = `support-bundle-([0-9]+(-[0-9]+)+)T([0-9]+(_[0-9]+)+)\.tar\.gz`
)

type Troubleshoot struct {
	Executable
}

func NewTroubleshoot(executable Executable) *Troubleshoot {
	return &Troubleshoot{
		Executable: executable,
	}
}

func (t *Troubleshoot) Collect(ctx context.Context, bundlePath string, sinceTime *time.Time, kubeconfig string) (archivePath string, err error) {
	marshalledTime, err := sinceTime.MarshalText()
	if err != nil {
		return "", fmt.Errorf("could not marshal sinceTime for Collect parameters: %v", err)
	}
	params := []string{bundlePath, "--kubeconfig", kubeconfig, "--interactive=false", "--since-time", string(marshalledTime)}

	output, err := t.Execute(ctx, params...)
	if err != nil {
		return "", fmt.Errorf("executing support-bundle: %v", err)
	}
	archivePath, err = parseArchivePathFromCollectOutput(output.String())
	if err != nil {
		return "", fmt.Errorf("parsing support-bundle output: %v", err)
	}
	return archivePath, nil
}

func (t *Troubleshoot) Analyze(ctx context.Context, bundleSpecPath string, archivePath string) ([]*SupportBundleAnalysis, error) {
	params := []string{"analyze", bundleSpecPath, "--bundle", archivePath, "--output", "json"}
	output, err := t.Execute(ctx, params...)
	if err != nil {
		return nil, fmt.Errorf("analyzing support bundle %s with analyzers %s: %v", archivePath, bundleSpecPath, err)
	}
	var analysisOutput []*SupportBundleAnalysis
	err = json.Unmarshal(output.Bytes(), &analysisOutput)
	if err != nil {
		return nil, fmt.Errorf("unmarshalling support-bundle analyze output: %v", err)
	}
	return analysisOutput, err
}

func parseArchivePathFromCollectOutput(tsLogs string) (archivePath string, err error) {
	r, err := regexp.Compile(supportBundleArchiveRegex)
	if err != nil {
		return "", fmt.Errorf("parsing support-bundle output: %v", err)
	}
	archivePath = r.FindString(tsLogs)
	if archivePath == "" {
		return "", fmt.Errorf("parsing support-bundle output: could not find archive path in output")
	}
	return archivePath, nil
}

type SupportBundleAnalysis struct {
	Title   string `json:"title"`
	IsPass  bool   `json:"isPass"`
	IsFail  bool   `json:"isFail"`
	IsWarn  bool   `json:"isWarn"`
	Message string `json:"message"`
	Uri     string `json:"URI"`
}
