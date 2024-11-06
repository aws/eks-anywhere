package analyzer

import (
	"fmt"
)

// NewPrinter returns a new Printer.
func NewPrinter() Printer {
	return Printer{}
}

// Printer prints the analysis result.
type Printer struct{}

// Process prints the analysis result.
func (Printer) Process(analysis ClusterAnalysisResult) error {
	if len(analysis.Findings) == 0 {
		print("%s Cluster %s is %s", blue(k8s), bold(analysis.Cluster.Name), green("healthy"))
		return nil
	}

	print("%s Cluster %s is %s", blue(k8s), bold(analysis.Cluster.Name), red("unhealthy"))

	for _, finding := range analysis.Findings {
		print("%s %s", fromArrow, finding.Message)
		for _, f := range unNestFindings(finding.Findings) {
			print("    %s", downArrow)
			print("    %s", f.Message)
			if f.Recommendation != "" {
				print("    %s %s %s", shortArrow, yellow("[Tip]"), f.Recommendation)
			}
			for _, log := range f.Logs {
				print("    %s %s", shortArrow, bold(log.Source))
				for _, line := range log.Lines {
					print("      %s", formatLog(line))
				}
			}
		}
	}

	return nil
}

func print(msg string, args ...any) {
	fmt.Printf(msg+"\n", args...)
}

func unNestFindings(findings []Finding) []Finding {
	var unnestedFindings []Finding
	for _, f := range findings {
		unnestedFindings = append(unnestedFindings, f)
		unnestedFindings = append(unnestedFindings, unNestFindings(f.Findings)...)
	}

	return unnestedFindings
}
