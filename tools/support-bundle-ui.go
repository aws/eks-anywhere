//go:generate fyne bundle -o icons.go eks-anywhere-logo.png

package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"sigs.k8s.io/yaml"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/replicatedhq/troubleshoot/pkg/convert"
)

type appTheme struct{}

var _ fyne.Theme = (*appTheme)(nil)

func (t appTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	if name == theme.ColorNameBackground {
		if variant == theme.VariantLight {
			return color.White
		}
		return color.Black
	}

	return theme.DefaultTheme().Color(name, variant)
}

func (m appTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m appTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m appTheme) Size(name fyne.ThemeSizeName) float32 {
	return theme.DefaultTheme().Size(name)
}

type stringList []string

func (sl stringList) contains(str string) bool {
	for _, elem := range sl {
		if elem == str {
			return true
		}
	}
	return false
}

// Display the following metadata about the cluster
// - eks-a version
// - k8s version
// - cp and worker node group replicas
// - os family
// - provider
// - etcd stacked / unstacked with replicas
// - registry mirror true/false
func addClusterMetadata(w fyne.Window, container *fyne.Container, supportBundleFolderPath string) {
	// Set label for cluster meta data
	clusterMetaDataLabel := canvas.NewText("CLUSTER METADATA", color.NRGBA{R: 255, G: 165, A: 255})
	clusterMetaDataLabel.TextStyle = fyne.TextStyle{Bold: true}
	clusterMetaDataLabel.TextSize = 35
	container.Add(clusterMetaDataLabel)

	// Parse cluster metadata from support bundle files
	crdPath := filepath.Join(supportBundleFolderPath, "cluster-resources", "custom-resources")
	clusterSpecFilePath := filepath.Join(crdPath, "clusters.anywhere.eks.amazonaws.com", "default.yaml")
	var clusterSpec []map[string]interface{}
	contents, err := os.ReadFile(clusterSpecFilePath)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	err = yaml.Unmarshal(contents, &clusterSpec)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}

	// Cards for metadata
	spec := clusterSpec[0]["spec"].(map[string]interface{})
	version := spec["eksaVersion"]
	eksaVersion := "Unknown"
	if version != nil {
		eksaVersion = version.(string)
	}
	container.Add(getCard("EKS-A Version", eksaVersion, 18, color.NRGBA{R: 255, G: 165, A: 255}))

	// k8s version
	k8sVersion := spec["kubernetesVersion"].(string)
	container.Add(getCard("K8s Version", k8sVersion, 18, color.NRGBA{R: 255, G: 165, A: 255}))

	// Provider
	dcRef := spec["datacenterRef"].(map[string]interface{})
	dcKind := dcRef["kind"].(string)
	var provider string
	var machineConfigFile string
	//var machineConfigKind string
	switch dcKind {
	case "VSphereDatacenterConfig":
		provider = "vSphere"
		machineConfigFile = "vspheremachineconfigs.anywhere.eks.amazonaws.com"
		//machineConfigKind = "VSphereMachineConfig"
	case "TinkerbellDatacenterConfig":
		provider = "Tinkerbell"
		machineConfigFile = "tinkerbellmachineconfigs.anywhere.eks.amazonaws.com"
		//machineConfigKind = "TinkerbellMachineConfig"
	case "CloudStackDatacenterConfig":
		provider = "CloudStack"
		machineConfigFile = "cloudstackmachineconfigs.anywhere.eks.amazonaws.com"
		//machineConfigKind = "CloudStackMachineConfig"
	default:
		provider = "Snow/Nutanix/Docker"
		machineConfigFile = "vspheremachineconfigs.anywhere.eks.amazonaws.com"
		//machineConfigKind = "VSphereMachineConfig"
	}
	container.Add(getCard("Provider", provider, 18, color.NRGBA{R: 255, G: 165, A: 255}))

	extEtcd := spec["externalEtcdConfiguration"]
	isExtEtcd := false
	if extEtcd != nil {
		isExtEtcd = true
	}
	container.Add(getCard("External Etcd", fmt.Sprintf("%t", isExtEtcd), 18, color.NRGBA{R: 255, G: 165, A: 255}))

	// CP count
	cpConf := spec["controlPlaneConfiguration"].(map[string]interface{})
	cpCount := cpConf["count"].(float64)
	container.Add(getCard("CP Replica", strconv.FormatFloat(cpCount, 'f', -1, 64), 18, color.NRGBA{R: 255, G: 165, A: 255}))

	// OS Family
	// We need cp machine config for OS family
	machineConfigFilePath := filepath.Join(filepath.Join(crdPath, machineConfigFile, "default.yaml"))
	var machineSpec []map[string]interface{}
	contents, err = os.ReadFile(machineConfigFilePath)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	err = yaml.Unmarshal(contents, &machineSpec)
	if err != nil {
		dialog.ShowError(err, w)
		return
	}
	machineSpecObj := machineSpec[0]["spec"].(map[string]interface{})
	osFamily := machineSpecObj["osFamily"].(string)
	container.Add(getCard("OS Family", osFamily, 18, color.NRGBA{R: 255, G: 165, A: 255}))

	// Registry mirror
	registryMirror := spec["registryMirrorConfiguration"]
	mirror := false
	if registryMirror != nil {
		mirror = true
	}
	container.Add(getCard("Registry Mirror", fmt.Sprintf("%t", mirror), 18, color.NRGBA{R: 255, G: 165, A: 255}))
}

func getCard(title, content string, textsize float32, textColor color.Color) *widget.Card {
	contentText := canvas.NewText(content, textColor)
	contentText.Alignment = fyne.TextAlignCenter
	contentText.TextSize = textsize
	card := &widget.Card{
		BaseWidget: widget.BaseWidget{},
		Title:      title,
		Subtitle:   "",
		Image:      nil,
		Content:    contentText,
	}
	return card
}

func getDeploymentCard(deploymentName string, gradientColor color.Color) *fyne.Container {
	gradient := canvas.NewHorizontalGradient(gradientColor, color.White)
	contentText := canvas.NewText(deploymentName, color.Black)
	contentText.TextStyle = fyne.TextStyle{Bold: true}
	contentText.Alignment = fyne.TextAlignCenter

	statusSymbol := canvas.NewText("✅", nil)
	statusSymbol.TextSize = 35
	deploymentStatusText := container.NewHBox(layout.NewSpacer(), contentText, layout.NewSpacer(), statusSymbol)
	ctr := container.NewStack(gradient, deploymentStatusText)
	return ctr
}

func extractInfoFromSupportBundle(w fyne.Window, metadataContainer, deploymentsContainer, crdStatusHeader, crdStatusContainer, logsAnalysisHeader, logsAnalysisContainer *fyne.Container) {
	dialog.ShowFileOpen(func(dir fyne.URIReadCloser, err error) {
		fileSelected := "No path selected"
		if err != nil {
			dialog.ShowError(err, w)
			return
		}
		if dir != nil {
			fileSelected = dir.URI().Path()
		}

		tarballDirectory := filepath.Dir(fileSelected)
		supportBundleFolderName := strings.Split(filepath.Base(fileSelected), ".")[0]
		supportBundleFolderPath := filepath.Join(tarballDirectory, supportBundleFolderName)

		err = os.RemoveAll(supportBundleFolderPath)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		fmt.Printf("Untaring tar file: %s\n", fileSelected)
		err = extractTarGz(fileSelected)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		var results []*convert.Result
		analysisFile := filepath.Join(supportBundleFolderPath, "analysis.json")
		contents, err := os.ReadFile(analysisFile)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		reported := stringList([]string{})

		fmt.Printf("Unmarshaling analysis file: %s\n", analysisFile)
		err = json.Unmarshal(contents, &results)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		fmt.Println("Adding cluster metadata")
		addClusterMetadata(w, metadataContainer, supportBundleFolderPath)

		fmt.Println("Adding Deployment statuses")
		deploymentStatusLabel := canvas.NewText("DEPLOYMENT STATUS", color.NRGBA{R: 255, G: 165, A: 255})
		deploymentStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
		deploymentStatusLabel.TextSize = 35
		deploymentsContainer.Add(deploymentStatusLabel)
		deploymentsContainer.Add(widget.NewLabel(""))
		for _, result := range results {
			if result.Severity == convert.SeverityError {
				if result.Labels["iconKey"] == "kubernetes_deployment_status" {
					deploymentName := strings.Join(strings.Split(result.Name, ".")[0:len(strings.Split(result.Name, "."))-1], "-")
					deploymentsContainer.Add(getDeploymentCard(deploymentName, color.NRGBA{R: 255, A: 255}))
				}
			} else if result.Severity == convert.SeverityDebug {
				if result.Labels["iconKey"] == "kubernetes_deployment_status" {
					deploymentName := strings.Join(strings.Split(result.Name, ".")[0:len(strings.Split(result.Name, "."))-1], "-")
					deploymentsContainer.Add(getDeploymentCard(deploymentName, color.NRGBA{G: 200, A: 255}))
				}
			}
		}
		if deploymentsContainer.Hidden {
			deploymentsContainer.Show()
		}

		fmt.Println("Adding CRD statuses")
		crdStatusLabel := canvas.NewText("CRD STATUS", color.NRGBA{R: 255, G: 165, A: 255})
		crdStatusLabel.TextStyle = fyne.TextStyle{Bold: true}
		crdStatusLabel.TextSize = 35
		crdStatusHeader.Add(crdStatusLabel)
		crdStatusHeader.Add(widget.NewLabel(""))
		for _, result := range results {
			reportCard := &widget.Card{}
			reportAccordion := &widget.Accordion{}
			if result.Severity == convert.SeverityError {
				if result.Labels["iconKey"] == "kubernetes_custom_resource_definition" {
					if !reported.contains(result.Name) {
						reportCard.Title = result.Name
						decoratedError := fmt.Sprintf("❌ %s", result.Error)
						errorText := canvas.NewText(decoratedError, color.NRGBA{R: 255, A: 255})
						reportCard.Content = errorText
						crdStatusContainer.Add(reportCard)
						reported = append(reported, result.Name)
					}
				}
			} else if result.Severity == convert.SeverityDebug {
				if result.Labels["iconKey"] == "kubernetes_custom_resource_definition" {
					if !reported.contains(result.Name) {
						decoratedError := fmt.Sprintf("✅ %s", result.Insight.Detail)
						errorText := canvas.NewText(decoratedError, color.NRGBA{G: 255, A: 255})
						errorText.TextSize = 18
						errorText.Alignment = fyne.TextAlignCenter
						reportAccordion.Items = append(reportAccordion.Items, &widget.AccordionItem{Title: result.Name, Detail: errorText, Open: false})
						crdStatusContainer.Add(reportAccordion)
						reported = append(reported, result.Name)
					}
				}
			}
		}

		fmt.Println("Adding logs analysis")
		logsAnalysisLabel := canvas.NewText("LOGS ANALYSIS", color.NRGBA{R: 255, G: 165, A: 255})
		logsAnalysisLabel.TextStyle = fyne.TextStyle{Bold: true}
		logsAnalysisLabel.TextSize = 35
		logsAnalysisHeader.Add(logsAnalysisLabel)
		logsAnalysisHeader.Add(widget.NewLabel(""))
		for _, result := range results {
			reportCard := &widget.Card{}
			reportAccordion := &widget.Accordion{}
			if result.Severity == convert.SeverityError {
				if result.Labels["iconKey"] == "kubernetes_text_analyze" {
					if !reported.contains(result.Name) {
						reportCard.Title = result.Name
						decoratedError := fmt.Sprintf("❌ %s", result.Error)
						errorText := canvas.NewText(decoratedError, color.NRGBA{R: 255, A: 255})
						reportCard.Content = errorText
						logsAnalysisContainer.Add(reportCard)
						reported = append(reported, result.Name)
					}
				}
			} else if result.Severity == convert.SeverityDebug {
				if result.Labels["iconKey"] == "kubernetes_text_analyze" {
					if !reported.contains(result.Name) {
						decoratedError := fmt.Sprintf("✅ %s", result.Insight.Detail)
						errorText := canvas.NewText(decoratedError, color.NRGBA{G: 255, A: 255})
						errorText.TextSize = 18
						errorText.Alignment = fyne.TextAlignCenter
						reportAccordion.Items = append(reportAccordion.Items, &widget.AccordionItem{Title: result.Name, Detail: errorText, Open: false})
						logsAnalysisContainer.Add(reportAccordion)
						reported = append(reported, result.Name)
					}
				}
			}
		}

		hidden := os.Getenv("CONTAINERS_HIDDEN")
		if hidden == "true" {
			metadataContainer.Show()
			deploymentsContainer.Show()
			crdStatusHeader.Show()
			crdStatusContainer.Show()
			logsAnalysisHeader.Show()
			logsAnalysisContainer.Show()
			_ = os.Setenv("CONTAINERS_HIDDEN", "false")
		}
	}, w)
}

func extractTarGz(supportBundlePath string) error {
	supportBundleFile, err := os.Open(supportBundlePath)
	if err != nil {
		return err
	}

	tarballDirectory := filepath.Dir(supportBundlePath)

	uncompressedStream, err := gzip.NewReader(supportBundleFile)
	if err != nil {
		return err
	}

	tarReader := tar.NewReader(uncompressedStream)
	var header *tar.Header
	for header, err = tarReader.Next(); err == nil; header, err = tarReader.Next() {
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.Mkdir(filepath.Join(tarballDirectory, header.Name), 0o755); err != nil {
				return fmt.Errorf("creating directory from archive: %v", err)
			}
		case tar.TypeReg:
			enclosingDirectory := filepath.Dir(header.Name)
			err = os.MkdirAll(filepath.Join(tarballDirectory, enclosingDirectory), 0o755)
			if err != nil {
				return err
			}
			outFile, err := os.Create(filepath.Join(tarballDirectory, header.Name))
			if err != nil {
				return fmt.Errorf("creating file from archive: %v", err)
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				outFile.Close()
				return fmt.Errorf("copying file to output destination: %v", err)
			}
			if err := outFile.Close(); err != nil {
				return fmt.Errorf("closing output destination file descriptor: %v", err)
			}
		case tar.TypeSymlink:
			enclosingDirectory := filepath.Dir(header.Name)
			err = os.MkdirAll(filepath.Join(tarballDirectory, enclosingDirectory), 0o755)
			if err != nil {
				return err
			}
			if err := os.Symlink(filepath.Join(tarballDirectory, filepath.Dir(header.Name), header.Linkname), filepath.Join(tarballDirectory, header.Name)); err != nil {
				return fmt.Errorf("writing symbolic link: %s", err)
			}
		default:
			return fmt.Errorf("unknown type in tar header: %b in %s", header.Typeflag, header.Name)
		}
	}
	if err != io.EOF {
		return fmt.Errorf("advancing to next entry in archive: %v", err)
	}
	return nil
}

func main() {
	a := app.New()
	a.Settings().SetTheme(&appTheme{})
	w := a.NewWindow("Support Bundle Analyzer")
	windowSize := fyne.NewSize(500, 500)
	w.Resize(windowSize)

	logo := canvas.NewImageFromResource(resourceEksAnywhereLogoPng)
	logo.FillMode = canvas.ImageFillOriginal
	logo.Translucency = 0.8
	logo.SetMinSize(windowSize)

	appHeader := canvas.NewText("SUPPORT BUNDLE ANALYZER", color.NRGBA{R: 255, G: 165, A: 255})
	appHeader.TextStyle = fyne.TextStyle{Bold: true}
	appHeader.Alignment = fyne.TextAlignCenter
	appHeader.TextSize = 50

	metadataContainer := container.New(layout.NewAdaptiveGridLayout(4))
	deploymentsContainer := container.New(layout.NewAdaptiveGridLayout(2))
	crdStatusHeader := container.New(layout.NewAdaptiveGridLayout(2))
	crdStatusContainer := container.NewVBox()
	logsAnalysisHeader := container.New(layout.NewAdaptiveGridLayout(2))
	logsAnalysisContainer := container.NewVBox()

	displayContainers := []fyne.CanvasObject{metadataContainer, deploymentsContainer, crdStatusHeader, crdStatusContainer, logsAnalysisHeader, logsAnalysisContainer}

	toolbar := &widget.Toolbar{
		Items: []widget.ToolbarItem{
			widget.NewToolbarAction(theme.FolderOpenIcon(), func() {
				extractInfoFromSupportBundle(w, metadataContainer, deploymentsContainer, crdStatusHeader, crdStatusContainer, logsAnalysisHeader, logsAnalysisContainer)
			}),
			widget.NewToolbarAction(theme.LogoutIcon(), func() {
				fmt.Println("Exiting")
				os.Exit(0)
			}),
			widget.NewToolbarAction(theme.ViewRefreshIcon(), func() {
				hidden := os.Getenv("CONTAINERS_HIDDEN")
				if hidden != "true" {
					for _, container := range displayContainers {
						container.Hide()
					}
					_ = os.Setenv("CONTAINERS_HIDDEN", "true")
				}
			}),
		},
	}

	windowElements := append([]fyne.CanvasObject{
		toolbar,
		appHeader,
	}, displayContainers...)

	windowContainer := container.NewVBox(windowElements...)

	scroller := container.NewVScroll(windowContainer)

	fullContainer := container.NewPadded(logo, scroller)

	themes := container.NewGridWithColumns(2,
		widget.NewButton("Dark", func() {
			a.Settings().SetTheme(theme.DarkTheme())
		}),
		widget.NewButton("Light", func() {
			a.Settings().SetTheme(theme.LightTheme())
		}),
	)

	w.SetContent(container.NewBorder(nil, themes, nil, nil, fullContainer))

	w.ShowAndRun()
}
