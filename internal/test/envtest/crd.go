package envtest

import (
	"bufio"
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"regexp"

	"github.com/pkg/errors"
)

type packageWithCRD struct {
	pkg   string
	regex *regexp.Regexp
	name  string
	path  string
}

func (p *packageWithCRD) pathToCRDs(version string) string {
	gopath := envOrDefault("GOPATH", build.Default.GOPATH)
	return filepath.Join(gopath, "pkg", "mod", p.path, fmt.Sprintf("%s@v%s", p.name, version), "config", "crd", "bases")
}

func mustBuildPackagesWithCRDs(packages ...string) []packageWithCRD {
	pkgs, err := buildPackagesWithCRD(packages...)
	if err != nil {
		panic(err)
	}

	return pkgs
}

func buildPackagesWithCRD(packages ...string) ([]packageWithCRD, error) {
	pkgs := make([]packageWithCRD, 0, len(packages))
	for _, p := range packages {
		pkgCRD, err := buildPackageWithCRD(p)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, *pkgCRD)
	}

	return pkgs, nil
}

func buildPackageWithCRD(pkg string) (*packageWithCRD, error) {
	r, err := regexp.Compile(fmt.Sprintf("%s%s v(.+)", `^(\W)`, pkg))
	if err != nil {
		return nil, errors.Wrapf(err, "failed building regex for package with CRD")
	}

	return &packageWithCRD{
		pkg:   pkg,
		regex: r,
		name:  filepath.Base(pkg),
		path:  filepath.Dir(pkg),
	}, nil
}

func getPathsToPackagesCRDs(rootFolder string, packages ...packageWithCRD) ([]string, error) {
	goModFile, err := os.Open(filepath.Join(rootFolder, "go.mod"))
	if err != nil {
		return nil, err
	}
	defer goModFile.Close()

	paths := make([]string, 0, len(packages))

	scanner := bufio.NewScanner(goModFile)
	for scanner.Scan() {
		moduleLine := scanner.Text()
		for _, p := range packages {
			matches := p.regex.FindStringSubmatch(moduleLine)
			if len(matches) == 3 {
				version := matches[2]
				paths = append(paths, p.pathToCRDs(version))
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return paths, nil
}

func envOrDefault(envKey, defaultValue string) string {
	if value, ok := os.LookupEnv(envKey); ok {
		return value
	}
	return defaultValue
}
