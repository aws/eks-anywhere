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
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/replicatedhq/troubleshoot/pkg/convert"
)

func extractErrorsFromSupportBundle(w fyne.Window, errorContainer *fyne.Container) {
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

		fmt.Printf("Unmarshaling analysis file: %s\n", analysisFile)
		err = json.Unmarshal(contents, &results)
		if err != nil {
			dialog.ShowError(err, w)
			return
		}

		for _, result := range results {
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
	errorContainer := container.NewVBox()
	selectButton := widget.NewButtonWithIcon("Select Support Bundle file", theme.FolderOpenIcon(), func() {
		extractErrorsFromSupportBundle(w, errorContainer)
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
		errorContainer,
	)

	scroller := container.NewVScroll(windowContainer)

	fullContainer := container.NewPadded(logo, scroller)
	w.SetContent(fullContainer)

	w.ShowAndRun()
}
