package framework

import (
	"strings"

	"github.com/aws/eks-anywhere/pkg/semver"
	releasev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
)

type CommandOpt func(*string, *[]string) (err error)

func appendOpt(new ...string) CommandOpt {
	return func(binaryPath *string, args *[]string) (err error) {
		*args = append(*args, new...)
		return nil
	}
}

func withKubeconfig(kubeconfigFile string) CommandOpt {
	return appendOpt("--kubeconfig", kubeconfigFile)
}

func WithForce() CommandOpt {
	return appendOpt("--force-cleanup")
}

func ExecuteWithEksaVersion(version *semver.Version) CommandOpt {
	return executeWithBinaryCommandOpt(func() (string, error) {
		return GetReleaseBinaryFromVersion(version)
	})
}

func ExecuteWithEksaRelease(release *releasev1alpha1.EksARelease) CommandOpt {
	return executeWithBinaryCommandOpt(func() (string, error) {
		return getBinary(release)
	})
}

func ExecuteWithLatestMinorReleaseFromVersion(version *semver.Version) CommandOpt {
	return executeWithBinaryCommandOpt(func() (string, error) {
		return GetLatestMinorReleaseBinaryFromVersion(version)
	})
}

func ExecuteWithLatestMinorReleaseFromMain() CommandOpt {
	return executeWithBinaryCommandOpt(func() (string, error) {
		return GetLatestMinorReleaseBinaryFromMain()
	})
}

func ExecuteWithLatestReleaseFromTestBranch() CommandOpt {
	return executeWithBinaryCommandOpt(func() (string, error) {
		return GetLatestMinorReleaseBinaryFromTestBranch()
	})
}

type binaryFetcher func() (binaryPath string, err error)

func executeWithBinaryCommandOpt(fetcher binaryFetcher) CommandOpt {
	return func(binaryPath *string, args *[]string) (err error) {
		b, err := fetcher()
		if err != nil {
			return err
		}
		*binaryPath = b
		if err = setEksctlVersionEnvVar(); err != nil {
			return err
		}

		// When bundles override is present, the manifest belongs to the current
		// build of the CLI and it's intended to be used only with that version
		removeFlag("--bundles-override", args)

		return nil
	}
}

func removeFlag(flag string, args *[]string) {
	for i, a := range *args {
		if a == flag {
			elementsToDelete := 1
			// If it's not the last arg and next arg is not a flag,
			// that means it's the value for the current flag, remove it as well
			if i < len(*args)-1 && !strings.HasPrefix((*args)[i+1], "-") {
				elementsToDelete = 2
			}

			*args = append((*args)[:i], (*args)[i+elementsToDelete:]...)
			break
		}
	}
}
