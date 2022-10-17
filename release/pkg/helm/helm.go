// Copyright Amazon.com Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package helm

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/helm/pkg/chartutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"

	"github.com/aws/eks-anywhere/release/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/pkg/types"
)

var BundleLog = ctrl.Log.WithName("BundleGenerator")

// helmDriver implements PackageDriver to install packages from Helm charts.
type helmDriver struct {
	cfg      *action.Configuration
	log      logr.Logger
	settings *cli.EnvSettings
}

func NewHelm() (*helmDriver, error) {
	settings := cli.New()
	client, err := registry.NewClient()
	if err != nil {
		return nil, fmt.Errorf("creating registry client while initializing helm driver: %w", err)
	}
	cfg := &action.Configuration{RegistryClient: client}
	err = cfg.Init(settings.RESTClientGetter(), settings.Namespace(),
		os.Getenv("HELM_DRIVER"), helmLog(BundleLog))
	if err != nil {
		return nil, fmt.Errorf("initializing helm driver: %w", err)
	}
	return &helmDriver{
		cfg:      cfg,
		log:      BundleLog,
		settings: settings,
	}, nil
}

func GetHelmDest(d *helmDriver, ReleaseImageURI, assetName string) (string, error) {
	var chartPath string
	var err error
	helmChart := strings.Split(ReleaseImageURI, ":")
	chartPath, err = d.PullHelmChart(helmChart[0], helmChart[1])
	if err != nil {
		return "", fmt.Errorf("pulling the helm chart: %w", err)
	}
	pwd, err := os.Getwd()
	dest := filepath.Join(pwd, assetName)
	if err != nil {
		return "", fmt.Errorf("getting current working dir: %w", err)
	}
	err = UnTarHelmChart(chartPath, assetName, dest)
	if err != nil {
		return "", fmt.Errorf("untar the helm chart: %w", err)
	}
	helmDest := filepath.Join(pwd, assetName, assetName)
	return helmDest, nil
}

func GetChartImageTags(d *helmDriver, helmDest string) (*Requires, error) {
	f, err := HasRequires(helmDest)
	if err != nil {
		return &Requires{}, fmt.Errorf("finding the requires.yaml: %w", err)
	}
	helmRequires, err := ValidateHelmRequires(f)
	if err != nil {
		return &Requires{}, fmt.Errorf("turning requires.yaml to struct: %w", err)
	}
	return helmRequires, nil
}

func ModifyAndPushChartYaml(i releasetypes.ImageArtifact, r *releasetypes.ReleaseConfig, d *helmDriver, helmDest string) error {
	helmChart := strings.Split(i.ReleaseImageURI, ":")
	helmtag := helmChart[1]

	// Overwrite Chart.yaml
	chart, err := HasChart(helmDest)
	if err != nil {
		return fmt.Errorf("finding the Chart.yaml: %w", err)
	}
	chartYaml, err := ValidateHelmChart(chart)
	if err != nil {
		return fmt.Errorf("turning Chart.yaml to struct: %w", err)
	}
	chartYaml.Version = helmtag
	err = OverwriteChartYaml(fmt.Sprintf("%s/%s", helmDest, "Chart.yaml"), chartYaml)
	if err != nil {
		return fmt.Errorf("overwriting the Chart.yaml version: %w", err)
	}
	packaged, err := PackageHelmChart(helmDest)
	if err != nil {
		return fmt.Errorf("packaging the helm chart: %w", err)
	}
	err = d.PushHelmChart(packaged, filepath.Dir(helmChart[0]))
	if err != nil {
		return fmt.Errorf("pushing the helm chart: %w", err)
	}
	return nil
}

// PullHelmChart will take in a a remote Helm URI and attempt to pull down the chart if it exists.
func (d *helmDriver) PullHelmChart(name, version string) (string, error) {
	if name == "" || version == "" {
		return "", fmt.Errorf("empty input for PullHelmChart, check flags")
	}
	install := action.NewInstall(d.cfg)
	install.ChartPathOptions.Version = version
	if !strings.HasPrefix(name, "oci://") {
		name = fmt.Sprintf("oci://%s", name)
	}
	chartPath, err := install.LocateChart(name, d.settings)
	if err != nil || chartPath == "" {
		return "", fmt.Errorf("running the Helm LocateChart command, you might need run an AWS ECR Login: %w", err)
	}
	return chartPath, nil
}

// PushHelmChart will take in packaged helm chart and push to a remote URI
func (d *helmDriver) PushHelmChart(packaged, URI string) error {
	config := action.WithPushConfig(d.cfg)
	p := action.NewPushWithOpts(config)
	if !strings.HasPrefix(URI, "oci://") {
		URI = fmt.Sprintf("oci://%s", URI)
	}
	_, err := p.Run(packaged, URI)
	if err != nil {
		return fmt.Errorf("running Helm push command on URI %s: %w", URI, err)
	}
	return nil
}

// PackageHelmChart will package a dir into a helm chart
func PackageHelmChart(dir string) (string, error) {
	if dir == "" {
		return "", fmt.Errorf("empty input for PackageHelmChart, check flags")
	}
	p := action.NewPackage()
	vals := new(map[string]interface{})
	packaged, err := p.Run(dir, *vals)
	if err != nil || packaged == "" {
		return "", fmt.Errorf("running the Helm Package command %w", err)
	}
	return packaged, nil
}

// helmLog wraps logr.Logger to make it compatible with helm's DebugLog.
func helmLog(log logr.Logger) action.DebugLog {
	return func(template string, args ...interface{}) {
		log.Info(fmt.Sprintf(template, args...))
	}
}

// UnTarHelmChart will attempt to move the helm chart out of the helm cache, by untaring it to the pwd and creating the filesystem to unpack it into.
func UnTarHelmChart(chartRef, chartPath, dest string) error {
	if chartRef == "" || chartPath == "" || dest == "" {
		return fmt.Errorf("Empty input value given for UnTarHelmChart")
	}
	_, err := os.Stat(dest)
	if os.IsNotExist(err) {
		if _, err := os.Stat(chartPath); err != nil {
			if err := os.MkdirAll(chartPath, 0o755); err != nil {
				return errors.Wrap(err, "failed to untar (mkdir)")
			}
		} else {
			return errors.Errorf("failed to untar: a file or directory with the name %s already exists", dest)
		}
	} else {
		if err != nil { // Checks directory check errors such as permission issues to read
			return errors.Errorf("failed UnTarHelmChart: %s", err)
		}
	}
	// Untar the files, and create the directory structure
	return chartutil.ExpandFile(dest, chartRef)
}

// HasRequires checks for the existance of the requires.yaml within the helm directory
func HasRequires(helmdir string) (string, error) {
	requires := filepath.Join(helmdir, "requires.yaml")
	info, err := os.Stat(requires)
	if os.IsNotExist(err) {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("found Dir, not requires.yaml file")
	}
	return requires, nil
}

// ValidateHelmRequires runs the parse file into struct function, and validations
func ValidateHelmRequires(fileName string) (*Requires, error) {
	helmrequires := &Requires{}
	err := parseHelmRequires(fileName, helmrequires)
	if err != nil {
		return nil, err
	}
	err = validateHelmRequiresContent(helmrequires)
	if err != nil {
		return nil, err
	}
	return helmrequires, err
}

// validateHelmRequiresContent loops over the validation tests
func validateHelmRequiresContent(helmrequires *Requires) error {
	for _, v := range helmRequiresValidations {
		if err := v(helmrequires); err != nil {
			return err
		}
	}
	return nil
}

var helmRequiresValidations = []func(*Requires) error{
	validateHelmRequiresName,
}

func validateHelmRequiresName(helmrequires *Requires) error {
	err := helmrequires.validateHelmRequiresNotEmpty()
	if err != nil {
		return err
	}
	return nil
}

// validateHelmRequiresNotEmpty checks that it has at least one image in the spec
func (helmrequires *Requires) validateHelmRequiresNotEmpty() error {
	// Check if Projects are listed
	if len(helmrequires.Spec.Images) < 1 {
		return fmt.Errorf("should use non-empty list of images for requires")
	}
	return nil
}

// parseHelmRequires will attempt to unpack the requires.yaml into the Go struct `Requires`
func parseHelmRequires(fileName string, helmrequires *Requires) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), constants.YamlSeparator) {
		if err = yaml.Unmarshal([]byte(c), helmrequires); err != nil {
			return fmt.Errorf("unable to parse %s\nyaml: %s\n %v", fileName, string(c), err)
		}
		err = yaml.UnmarshalStrict([]byte(c), helmrequires)
		if err != nil {
			return fmt.Errorf("unable to UnmarshalStrict %v\nyaml: %s\n %v", helmrequires, string(c), err)
		}
		return nil
	}
	return fmt.Errorf("requires.yaml file [%s] is invalid or does not contain kind %v", fileName, helmrequires)
}

// Chart yaml functions

// HasChart checks for the existance of the Chart.yaml within the helm directory
func HasChart(helmdir string) (string, error) {
	requires := filepath.Join(helmdir, "Chart.yaml")
	info, err := os.Stat(requires)
	if os.IsNotExist(err) {
		return "", err
	}
	if info.IsDir() {
		return "", fmt.Errorf("found Dir, not Chart.yaml file")
	}
	return requires, nil
}

// ValidateHelmChart runs the parse file into struct function, and validations
func ValidateHelmChart(fileName string) (*chart.Metadata, error) {
	helmChart := &chart.Metadata{}
	err := parseHelmChart(fileName, helmChart)
	if err != nil {
		return nil, err
	}
	return helmChart, err
}

// parseHelmChart will attempt to unpack the Chart.yaml into the Go struct `Chart`
func parseHelmChart(fileName string, helmChart *chart.Metadata) error {
	content, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("unable to read file due to: %v", err)
	}
	for _, c := range strings.Split(string(content), constants.YamlSeparator) {
		if err = yaml.Unmarshal([]byte(c), helmChart); err != nil {
			return fmt.Errorf("unable to parse %s\nyaml: %s\n %v", fileName, string(c), err)
		}
		err = yaml.UnmarshalStrict([]byte(c), helmChart)
		if err != nil {
			return fmt.Errorf("unable to UnmarshalStrict %v\nyaml: %s\n %v", helmChart, string(c), err)
		}
		return nil
	}
	return fmt.Errorf("Chart.yaml file [%s] is invalid or does not contain kind %v", fileName, helmChart)
}

func OverwriteChartYaml(filename string, helmChart *chart.Metadata) error {
	yamlData, err := yaml.Marshal(&helmChart)
	if err != nil {
		return fmt.Errorf("unable to Marshal %v\nyamlData: %s\n %v", helmChart, yamlData, err)
	}

	err = ioutil.WriteFile(filename, yamlData, 0o644)
	if err != nil {
		return err
	}
	return nil
}
