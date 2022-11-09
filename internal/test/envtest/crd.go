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

type moduleWithCRD struct {
	pkg          string
	crdPaths     []string
	requireRegex *regexp.Regexp
	replaceRegex *regexp.Regexp
}

func mustBuildModuleWithCRDs(p string, opts ...moduleOpt) moduleWithCRD {
	pkgCRD, err := buildModuleWithCRD(p, opts...)
	if err != nil {
		panic(err)
	}

	return *pkgCRD
}

func withAdditionalCustomCRDPath(customCRDPath string) moduleOpt {
	return func(m *moduleWithCRD) {
		m.crdPaths = append(m.crdPaths, customCRDPath)
	}
}

func withMainCustomCRDPath(customCRDPath string) moduleOpt {
	return func(m *moduleWithCRD) {
		m.crdPaths[0] = customCRDPath
	}
}

type moduleOpt func(*moduleWithCRD)

func buildModuleWithCRD(pkg string, opts ...moduleOpt) (*moduleWithCRD, error) {
	requireRegex, err := regexp.Compile(fmt.Sprintf("%s%s v(.+)", `^(\W)`, pkg))
	if err != nil {
		return nil, errors.Wrapf(err, "failed building regex for package with CRD")
	}

	replaceRegex, err := regexp.Compile(fmt.Sprintf("%s%s => (.+) v(.+)", `^(\W)`, pkg))
	if err != nil {
		return nil, errors.Wrapf(err, "failed building regex for package with CRD")
	}

	m := &moduleWithCRD{
		pkg:          pkg,
		requireRegex: requireRegex,
		replaceRegex: replaceRegex,
		crdPaths:     []string{"config/crd/bases"},
	}

	for _, opt := range opts {
		opt(m)
	}

	return m, nil
}

type moduleInDisk struct {
	moduleWithCRD
	path, name, version string
}

func (m moduleInDisk) pathsToCRDs() []string {
	paths := make([]string, 0, len(m.crdPaths))
	for _, crdPath := range m.crdPaths {
		paths = append(paths, pathToCRDs(m.path, m.name, m.version, crdPath))
	}
	return paths
}

func pathToCRDs(path, name, version, crdPath string) string {
	gopath := envOrDefault("GOPATH", build.Default.GOPATH)
	return filepath.Join(gopath, "pkg", "mod", path, fmt.Sprintf("%s@v%s", name, version), crdPath)
}

func getPathsToPackagesCRDs(rootFolder string, packages ...moduleWithCRD) ([]string, error) {
	goModFile, err := os.Open(filepath.Join(rootFolder, "go.mod"))
	if err != nil {
		return nil, err
	}
	defer goModFile.Close()

	modulesMappedToDisk := buildModulesMappedToDisk(packages)

	scanner := bufio.NewScanner(goModFile)
	for scanner.Scan() {
		moduleLine := scanner.Text()
		for _, p := range packages {
			matches := p.requireRegex.FindStringSubmatch(moduleLine)
			if len(matches) == 3 {
				version := matches[2]
				moduleInDisk := modulesMappedToDisk[p.pkg]
				if moduleInDisk.version != "" {
					// If the package has already been mapped to disk, it was
					// probably by a replace, don't overwrite
					continue
				}

				moduleInDisk.path = filepath.Dir(p.pkg)
				moduleInDisk.name = filepath.Base(p.pkg)
				moduleInDisk.version = version
				continue
			}

			matches = p.replaceRegex.FindStringSubmatch(moduleLine)
			if len(matches) == 4 {
				replaceModule := matches[2]
				replaceVersion := matches[3]
				modulesMappedToDisk[p.pkg] = &moduleInDisk{
					moduleWithCRD: p,
					path:          filepath.Dir(replaceModule),
					name:          filepath.Base(replaceModule),
					version:       replaceVersion,
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	paths := make([]string, 0, len(modulesMappedToDisk))
	for _, m := range modulesMappedToDisk {
		if m.version == "" {
			return nil, fmt.Errorf("couldn't find module in disk for %s", m.pkg)
		}
		paths = append(paths, m.pathsToCRDs()...)
	}

	return paths, nil
}

func envOrDefault(envKey, defaultValue string) string {
	if value, ok := os.LookupEnv(envKey); ok {
		return value
	}
	return defaultValue
}

func buildModulesMappedToDisk(modules []moduleWithCRD) map[string]*moduleInDisk {
	modulesMappedToDisk := make(map[string]*moduleInDisk, len(packages))
	for _, m := range modules {
		modulesMappedToDisk[m.pkg] = &moduleInDisk{
			moduleWithCRD: m,
		}
	}

	return modulesMappedToDisk
}
