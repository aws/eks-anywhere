package framework

import "github.com/aws/eks-anywhere/pkg/semver"

type commandOpt func(*[]string)

func appendOpt(new ...string) commandOpt {
	return func(args *[]string) {
		*args = append(*args, new...)
	}
}

func withKubeconfig(kubeconfigFile string) commandOpt {
	return appendOpt("--kubeconfig", kubeconfigFile)
}

type VersionOpt func() (binaryPath string, err error)

func ExecuteWithEksaVersion(version *semver.Version) VersionOpt {
	return func() (binaryPath string, err error) {
		return GetReleaseBinaryFromVersion(version)
	}
}

func ExecuteWithLatestMinorReleaseFromVersion(version *semver.Version) VersionOpt {
	return func() (binaryPath string, err error) {
		return GetLatestMinorReleaseBinaryFromVersion(version)
	}
}

func ExecuteWithLatestMinorReleaseFromMain() VersionOpt {
	return func() (binaryPath string, err error) {
		return GetLatestMinorReleaseBinaryFromMain()
	}
}
