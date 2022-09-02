package diagnostics

import (
	"fmt"
	"github.com/aws/eks-anywhere/pkg/executables"
	"log"

	ui "github.com/replicatedhq/termui/v3"
	"github.com/replicatedhq/termui/v3/widgets"
)

var (
	selectedResult = 0
	table          = widgets.NewTable()
	isShowingSaved = false
)

func printOutput(supportBundleName string, analyzers []*executables.SupportBundleAnalysis) {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()
	drawUI(supportBundleName, analyzers)

	uiEvents := ui.PollEvents()
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "<C-c>":
				return
			case "q":
				return
			case "<Down>":
				if selectedResult < len(analyzers)-1 {
					selectedResult++
				} else {
					selectedResult = 0
					//table.SelectedRow = 0
				}
				//table.ScrollDown()
				ui.Clear()
				drawUI(supportBundleName, analyzers)
			case "<Up>":
				if selectedResult > 0 {
					selectedResult--
				} else {
					selectedResult = len(analyzers) - 1
					//table.SelectedRow = len(analyzeResults)
				}
				//table.ScrollUp()
				ui.Clear()
				drawUI(supportBundleName, analyzers)
			}
		}
	}

}

func drawUI(supportBundleName string, analyzeResults []*executables.SupportBundleAnalysis) {
	drawGrid(analyzeResults)
	drawHeader(supportBundleName)
	drawFooter()
}

func drawHeader(supportBundleName string) {
	termWidth, _ := ui.TerminalDimensions()

	title := widgets.NewParagraph()
	title.Text = fmt.Sprintf("%s Support Bundle Analysis", supportBundleName)
	title.TextStyle.Fg = ui.ColorWhite
	title.TextStyle.Bg = ui.ColorClear
	title.TextStyle.Modifier = ui.ModifierBold
	title.Border = false

	left := termWidth/2 - 2*len(title.Text)/3
	right := termWidth/2 + (termWidth/2 - left)

	title.SetRect(left, 0, right, 1)
	ui.Render(title)
}
func drawFooter() {
	termWidth, termHeight := ui.TerminalDimensions()

	instructions := widgets.NewParagraph()
	instructions.Text = "[q] quit    [s] save    [↑][↓] scroll"
	instructions.Border = false

	left := 0
	right := termWidth
	top := termHeight - 1
	bottom := termHeight

	instructions.SetRect(left, top, right, bottom)
	ui.Render(instructions)
}

func drawGrid(analyzeResults []*executables.SupportBundleAnalysis) {
	drawAnalyzersTable(analyzeResults)
	drawDetails(analyzeResults[selectedResult])
}
func drawAnalyzersTable(analyzeResults []*executables.SupportBundleAnalysis) {
	termWidth, termHeight := ui.TerminalDimensions()

	table.SetRect(0, 3, termWidth/2, termHeight-6)
	table.FillRow = true
	table.Border = true
	table.Rows = [][]string{}
	table.ColumnWidths = []int{termWidth}

	for i, analyzeResult := range analyzeResults {
		title := analyzeResult.Title
		if analyzeResult.IsPass {
			title = fmt.Sprintf("✔  %s", title)
		} else if analyzeResult.IsWarn {
			title = fmt.Sprintf("⚠️  %s", title)
		} else if analyzeResult.IsFail {
			title = fmt.Sprintf("✘  %s", title)
		}
		table.Rows = append(table.Rows, []string{
			title,
		})

		if analyzeResult.IsPass {
			if i == selectedResult {
				table.RowStyles[i] = ui.NewStyle(ui.ColorGreen, ui.ColorClear, ui.ModifierReverse)
			} else {
				table.RowStyles[i] = ui.NewStyle(ui.ColorGreen, ui.ColorClear)
			}
		} else if analyzeResult.IsWarn {
			if i == selectedResult {
				table.RowStyles[i] = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierReverse)
			} else {
				table.RowStyles[i] = ui.NewStyle(ui.ColorYellow, ui.ColorClear)
			}
		} else if analyzeResult.IsFail {
			if i == selectedResult {
				table.RowStyles[i] = ui.NewStyle(ui.ColorRed, ui.ColorClear, ui.ModifierReverse)
			} else {
				table.RowStyles[i] = ui.NewStyle(ui.ColorRed, ui.ColorClear)
			}
		}
	}

	ui.Render(table)
}
func drawDetails(analysisResult *executables.SupportBundleAnalysis) {
	termWidth, _ := ui.TerminalDimensions()

	currentTop := 4
	title := widgets.NewParagraph()
	title.Text = analysisResult.Title
	title.Border = false
	if analysisResult.IsPass {
		title.TextStyle = ui.NewStyle(ui.ColorGreen, ui.ColorClear, ui.ModifierBold)
	} else if analysisResult.IsWarn {
		title.TextStyle = ui.NewStyle(ui.ColorYellow, ui.ColorClear, ui.ModifierBold)
	} else if analysisResult.IsFail {
		title.TextStyle = ui.NewStyle(ui.ColorRed, ui.ColorClear, ui.ModifierBold)
	}
	height := estimateNumberOfLines(title.Text, termWidth/2)
	title.SetRect(termWidth/2, currentTop, termWidth, currentTop+height)
	ui.Render(title)
	currentTop = currentTop + height + 1

	message := widgets.NewParagraph()
	message.Text = analysisResult.Message
	message.Border = false
	height = estimateNumberOfLines(message.Text, termWidth/2) + 2
	message.SetRect(termWidth/2, currentTop, termWidth, currentTop+height)
	ui.Render(message)
	currentTop = currentTop + height + 1

	if analysisResult.Uri != "" {
		uri := widgets.NewParagraph()
		uri.Text = fmt.Sprintf("For more information: %s", analysisResult.Uri)
		uri.Border = false
		height = estimateNumberOfLines(uri.Text, termWidth/2)
		uri.SetRect(termWidth/2, currentTop, termWidth, currentTop+height)
		ui.Render(uri)
		currentTop = currentTop + height + 1
	}
}
func estimateNumberOfLines(text string, width int) int {
	lines := len(text)/width + 1
	return lines
}
