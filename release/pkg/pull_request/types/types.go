package types

type PullRequestConfig struct {
	GithubUser         string
	Branch             string
	ReleaseType        string
	ReleaseEnvironment string
	BundleNumber       string
	CliMinVersion      string
	CliMaxVersion      string
	ReleaseNumber      string
	ReleaseVersion     string
	DryRun             bool
}

const (
	BundleKind = "bundle"
	EksAKind   = "eks-a"
)
