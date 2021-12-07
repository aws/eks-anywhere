package awsprofiles

type EksAccount int64

const (
	BuildAccount EksAccount = iota
	TestAccount
)

func (s EksAccount) ProfileName() string {
	switch s {
	case BuildAccount:
		return "eks-a-build-account"
	case TestAccount:
		return "eks-a-test-account"
	}
	return "unknown"
}
