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
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/registry"
	"k8s.io/helm/pkg/chartutil"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/yaml"

	anywherev1alpha1 "github.com/aws/eks-anywhere/release/api/v1alpha1"
	"github.com/aws/eks-anywhere/release/cli/pkg/constants"
	releasetypes "github.com/aws/eks-anywhere/release/cli/pkg/types"
	commandutils "github.com/aws/eks-anywhere/release/cli/pkg/util/command"
)

var HelmLog = ctrl.Log.WithName("HelmLog")

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
		os.Getenv("HELM_DRIVER"), helmLog(HelmLog))
	if err != nil {
		return nil, fmt.Errorf("initializing helm driver: %w", err)
	}
	return &helmDriver{
		cfg:      cfg,
		log:      HelmLog,
		settings: settings,
	}, nil
}

func GetHelmDest(d *helmDriver, r *releasetypes.ReleaseConfig, ReleaseImageURI, assetName string) (string, error) {
	var chartPath string
	var err error

	err = d.HelmRegistryLogin(r, "source")
	if err != nil {
		return "", fmt.Errorf("logging into the source registry: %w", err)
	}

	helmChart := strings.Split(ReleaseImageURI, ":")
	fmt.Printf("Starting to modifying helm chart %s\n", helmChart[0])
	fmt.Printf("Pulling helm chart %s\n", ReleaseImageURI)
	chartPath, err = d.PullHelmChart(helmChart[0], helmChart[1])
	if err != nil {
		return "", fmt.Errorf("pulling the helm chart: %w", err)
	}

	err = d.HelmRegistryLogout(r, "source")
	if err != nil {
		return "", fmt.Errorf("logging out of the source registry: %w", err)
	}

	pwd, err := os.Getwd()
	dest := filepath.Join(pwd, assetName)
	if err != nil {
		return "", fmt.Errorf("getting current working dir: %w", err)
	}
	fmt.Printf("Untar helm chart %s into %s\n", chartPath, dest)
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

func ModifyAndPushChartYaml(i releasetypes.ImageArtifact, r *releasetypes.ReleaseConfig, d *helmDriver, helmDest string, eksaArtifacts map[string][]releasetypes.Artifact, shaMap map[string]anywherev1alpha1.Image) error {
	helmChart := strings.Split(i.ReleaseImageURI, ":")
	helmtag := helmChart[1]

	// Overwrite Chart.yaml
	fmt.Printf("Checking inside helm chart for Chart.yaml %s\n", helmDest)
	chart, err := HasChart(helmDest)
	if err != nil {
		return fmt.Errorf("finding the Chart.yaml: %w", err)
	}

	chartYaml, err := ValidateHelmChart(chart)
	if err != nil {
		return fmt.Errorf("turning Chart.yaml to struct: %w", err)
	}
	chartYaml.Version = helmtag

	fmt.Printf("Overwriting helm chart.yaml version to new tag %s\n", chartYaml.Version)
	err = OverwriteChartYaml(fmt.Sprintf("%s/%s", helmDest, "Chart.yaml"), chartYaml)
	if err != nil {
		return fmt.Errorf("overwriting the Chart.yaml version: %w", err)
	}

	// If the chart is packages, we find the image tag values and overide them in the values.yaml.
	if strings.Contains(helmDest, "eks-anywhere-packages") {
		imageTagMap, err := GetPackagesImageTags(eksaArtifacts)
		fmt.Printf("Overwriting helm values.yaml version to new image tags %v\n", imageTagMap)
		err = OverWriteChartValuesImageTag(helmDest, imageTagMap)
		if err != nil {
			return fmt.Errorf("overwriting the values.yaml version: %w", err)
		}
		if shaMap != nil {
			fmt.Printf("Overwriting helm values.yaml image shas to new image shas for Dev Release %v\n", shaMap)
			err = OverWriteChartValuesImageSha(helmDest, shaMap)
			if err != nil {
				return fmt.Errorf("overwriting the values.yaml version: %w", err)
			}
		}
	}

	fmt.Printf("Re-Packaging modified helm chart %s\n", helmDest)
	packaged, err := PackageHelmChart(helmDest)
	if err != nil {
		return fmt.Errorf("packaging the helm chart: %w", err)
	}

	fmt.Printf("Pushing modified helm chart %s to %s\n", packaged, r.ReleaseContainerRegistry)
	err = d.HelmRegistryLogin(r, "destination")
	if err != nil {
		return fmt.Errorf("logging into the destination registry: %w", err)
	}

	err = PushHelmChart(packaged, filepath.Dir(helmChart[0]))
	if err != nil {
		return fmt.Errorf("pushing the helm chart: %w", err)
	}

	err = d.HelmRegistryLogout(r, "destination")
	if err != nil {
		return fmt.Errorf("logging out of the destination registry: %w", err)
	}

	return nil
}

func (d *helmDriver) HelmRegistryLogin(r *releasetypes.ReleaseConfig, remoteType string) error {
	var authConfig *docker.AuthConfiguration
	var remote string
	if remoteType == "source" {
		authConfig = r.SourceClients.ECR.AuthConfig
		remote = r.SourceContainerRegistry
	} else if remoteType == "destination" {
		authConfig = r.ReleaseClients.ECRPublic.AuthConfig
		remote = r.ReleaseContainerRegistry
	}
	login := action.NewRegistryLogin(d.cfg)
	err := login.Run(os.Stdout, remote, authConfig.Username, authConfig.Password, false)
	if err != nil {
		return fmt.Errorf("running the Helm registry login command: %w", err)
	}

	return nil
}

func (d *helmDriver) HelmRegistryLogout(r *releasetypes.ReleaseConfig, remoteType string) error {
	var remote string
	if remoteType == "source" {
		remote = r.SourceContainerRegistry
	} else if remoteType == "destination" {
		remote = r.ReleaseContainerRegistry
	}
	logout := action.NewRegistryLogout(d.cfg)
	err := logout.Run(os.Stdout, remote)
	if err != nil {
		return fmt.Errorf("running the Helm registry logout command: %w", err)
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

// PushHelmChart will take in packaged helm chart and push to a remote URI.
func PushHelmChart(packaged, URI string) error {
	if !strings.HasPrefix(URI, "oci://") {
		URI = fmt.Sprintf("oci://%s", URI)
	}

	cmd := exec.Command("helm", "push", packaged, URI)
	out, err := commandutils.ExecCommand(cmd)
	fmt.Println(out)

	if err != nil {
		return fmt.Errorf("running Helm push command on URI %s: %v", URI, err)
	}
	return nil
}

// PackageHelmChart will package a dir into a helm chart.
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

// HasRequires checks for the existance of the requires.yaml within the helm directory.
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

// ValidateHelmRequires runs the parse file into struct function, and validations.
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

// validateHelmRequiresContent loops over the validation tests.
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

// validateHelmRequiresNotEmpty checks that it has at least one image in the spec.
func (helmrequires *Requires) validateHelmRequiresNotEmpty() error {
	// Check if Projects are listed
	if len(helmrequires.Spec.Images) < 1 {
		return fmt.Errorf("should use non-empty list of images for requires")
	}
	return nil
}

// parseHelmRequires will attempt to unpack the requires.yaml into the Go struct `Requires`.
func parseHelmRequires(fileName string, helmrequires *Requires) error {
	content, err := os.ReadFile(fileName)
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

// HasChart checks for the existance of the Chart.yaml within the helm directory.
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

// ValidateHelmChart runs the parse file into struct function, and validations.
func ValidateHelmChart(fileName string) (*chart.Metadata, error) {
	helmChart := &chart.Metadata{}
	err := parseHelmChart(fileName, helmChart)
	if err != nil {
		return nil, err
	}
	return helmChart, err
}

// parseHelmChart will attempt to unpack the Chart.yaml into the Go struct `Chart`.
func parseHelmChart(fileName string, helmChart *chart.Metadata) error {
	content, err := os.ReadFile(fileName)
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

	err = os.WriteFile(filename, yamlData, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func OverWriteChartValuesImageTag(filename string, tagMap map[string]string) error {
	packagesURI := strings.Split(tagMap["eks-anywhere-packages"], ":")
	refresherURI := strings.Split(tagMap["ecr-token-refresher"], ":")

	valuesFile := filepath.Join(filename, "values.yaml")
	values, err := chartutil.ReadValuesFile(valuesFile)
	if err != nil {
		return err
	}
	values["controller"].(map[string]interface{})["tag"] = packagesURI[len(packagesURI)-1]
	values["cronjob"].(map[string]interface{})["tag"] = refresherURI[len(refresherURI)-1]
	yamlData, err := yaml.Marshal(&values)
	if err != nil {
		return fmt.Errorf("unable to Marshal %v\nyamlData: %s\n %v", values, yamlData, err)
	}
	err = os.WriteFile(valuesFile, yamlData, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func OverWriteChartValuesImageSha(filename string, shaMap map[string]anywherev1alpha1.Image) error {
	valuesFile := filepath.Join(filename, "values.yaml")
	values, err := chartutil.ReadValuesFile(valuesFile)
	if err != nil {
		return err
	}
	values["controller"].(map[string]interface{})["digest"] = shaMap["eks-anywhere-packages"].ImageDigest
	values["cronjob"].(map[string]interface{})["digest"] = shaMap["ecr-token-refresher"].ImageDigest
	yamlData, err := yaml.Marshal(&values)
	if err != nil {
		return fmt.Errorf("unable to Marshal %v\nyamlData: %s\n %v", values, yamlData, err)
	}
	err = os.WriteFile(valuesFile, yamlData, 0o644)
	if err != nil {
		return err
	}
	return nil
}

func GetPackagesImageTags(packagesArtifacts map[string][]releasetypes.Artifact) (map[string]string, error) {
	m := make(map[string]string)
	for _, artifacts := range packagesArtifacts {
		for _, artifact := range artifacts {
			if artifact.Image != nil {
				m[artifact.Image.AssetName] = artifact.Image.ReleaseImageURI
			}
		}
	}
	if len(m) == 0 {
		return nil, fmt.Errorf("No assets found for eks-anywhere-packages, or ecr-token-refresher in packagesArtifacts")
	}
	return m, nil
}
