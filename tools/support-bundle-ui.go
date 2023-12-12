package main

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strconv"
	"strings"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/replicatedhq/troubleshoot/pkg/convert"
)

// Display the following metadata about the cluster
// - eks-a version
// - k8s version
// - cp and worker node group replicas
// - os family
// - provider
// - etcd stacked / unstacked with replicas
// - registry mirror true/false
func addClusterMetadata(w fyne.Window, container *fyne.Container, supportBundleFolderPath string, errorResults []*convert.Result) {
	// Set label for cluster meta data
	clusterMetaDataLabel := &widget.Label{
		BaseWidget: widget.BaseWidget{},
		Text:       "Cluster Metadata",
		Alignment:  0,
		Wrapping:   0,
		TextStyle: fyne.TextStyle{
			Bold: true,
		},
		Truncation: 0,
		Importance: 0,
	}
	container.Add(clusterMetaDataLabel)

	// Parse cluster metadata from support bundle files
	crdPath := filepath.Join(supportBundleFolderPath, "cluster-resources", "custom-resource-definitions")
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
	eksaVersion := ""
	if version != nil {
		eksaVersion = version.(string)
	}
	container.Add(getCard("EKS-A Version", eksaVersion, 18, color.Black))

	// k8s version
	k8sVersion := spec["kubernetesVersion"].(string)
	container.Add(getCard("K8s Version", k8sVersion, 18, color.Black))

	// Provider
	dcRef := spec["datacenterRef"].(map[string]interface{})
	dcKind := dcRef["kind"].(string)
	var provider string
	switch dcKind {
	case "VSphereDatacenterConfig":
		provider = "vSphere"
	case "TinkerbellDatacenterConfig":
		provider = "Tinkerbell"
	case "CloudStackDatacenterConfig":
		provider = "CloudStack"
	default:
		provider = "Snow/Nutanix/Docker"
	}
	container.Add(getCard("Provider", provider, 18, color.Black))

	extEtcd := spec["externalEtcdConfiguration"]
	isExtEtcd := false
	if extEtcd != nil {
		isExtEtcd = true
	}
	container.Add(getCard("External Etcd", fmt.Sprintf("%t", isExtEtcd), 18, color.Black))

	// CP count
	cpConf := spec["controlPlaneConfiguration"].(map[string]interface{})
	cpCount := cpConf["count"].(float64)
	container.Add(getCard("CP Replica", strconv.FormatFloat(cpCount, 'f', -1, 64), 18, color.Black))

	// OS Family
}

func getCard(title, content string, textsize float32, color color.Color) *widget.Card {
	contentText := canvas.NewText(content, color)
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

func extractErrorsFromSupportBundle(w fyne.Window, metadataContainer, errorContainer *fyne.Container) {
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
		contents, _ := os.ReadFile(analysisFile)
		//if err != nil {
		//	dialog.ShowError(err, w)
		//	return
		//}

		fmt.Printf("Unmarshaling analysis file: %s\n", analysisFile)
		err = json.Unmarshal(contents, &results)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		addClusterMetadata(w, metadataContainer, supportBundleFolderPath, results)
		for _, result := range results {
			reportCard := &widget.Card{}
			if result.Severity == convert.SeverityError {
				reportCard.Title = result.Name
				decoratedError := fmt.Sprintf("❌ %s", result.Error)
				errorText := canvas.NewText(decoratedError, color.NRGBA{R: 255, A: 255})
				reportCard.Content = errorText
				errorContainer.Add(reportCard)
			}
			reportAccordion := &widget.Accordion{}
			if result.Severity == convert.SeverityError {
				decoratedError := fmt.Sprintf("❌ %s", result.Error)
				errorText := canvas.NewText(decoratedError, color.NRGBA{R: 255, A: 255})
				errorText.TextSize = 30
				errorText.Alignment = fyne.TextAlignCenter
				reportAccordion.Items = append(reportAccordion.Items, &widget.AccordionItem{Title: result.Name, Detail: errorText, Open: false})
				errorContainer.Add(reportAccordion)
			} else if result.Severity == convert.SeverityDebug {
				decoratedError := fmt.Sprintf("✅ %s", result.Insight.Detail)
				errorText := canvas.NewText(decoratedError, color.NRGBA{G: 255, A: 255})
				errorText.TextSize = 30
				errorText.Alignment = fyne.TextAlignCenter
				reportAccordion.Items = append(reportAccordion.Items, &widget.AccordionItem{Title: result.Name, Detail: errorText, Open: false})
				errorContainer.Add(reportAccordion)
			}
		}
		if errorContainer.Hidden {
			errorContainer.Show()
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
	w := a.NewWindow("Support Bundle Analyzer")
	windowSize := fyne.NewSize(500, 500)
	w.Resize(windowSize)

	logo := canvas.NewImageFromFile("/Users/arnchlm/Downloads/eks-anywhere-logo.png")
	logo.FillMode = canvas.ImageFillOriginal
	logo.Translucency = 0.8
	logo.SetMinSize(windowSize)

	hello := widget.NewLabelWithStyle("Support Bundle Analyzer", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	metadataContainer := container.New(layout.NewAdaptiveGridLayout(4))
	errorContainer := container.NewVBox()
	selectButton := widget.NewButtonWithIcon("Select Support Bundle file", theme.FolderOpenIcon(), func() {
		extractErrorsFromSupportBundle(w, metadataContainer, errorContainer)
	})
	exitButton := widget.NewButtonWithIcon("Exit", theme.LogoutIcon(), func() {
		fmt.Println("Exiting")
		os.Exit(0)
	})
	startOverButton := widget.NewButtonWithIcon("Start Over", theme.ViewRefreshIcon(), func() {
		errorContainer.Hide()
	})

	windowContainer := container.NewVBox(
		hello,
		selectButton,
		startOverButton,
		exitButton,
		metadataContainer,
		errorContainer,
	)

	scroller := container.NewVScroll(windowContainer)

	fullContainer := container.NewPadded(logo, scroller)
	w.SetContent(fullContainer)

	w.ShowAndRun()
}
