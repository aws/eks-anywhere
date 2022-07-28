package curatedpackages

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// cpTabwriter is a modified tabwriter for curated packages CLI duty.
type cpTabwriter struct {
	*tabwriter.Writer
}

// newCPTabwriter instantiates a curated packages custom tabwriter.
//
// If p is nil, cpTabwriterDefaultParams will be used. The caller should call
// Flush just as they would with an unmodified tabwriter.Writer.
func newCPTabwriter(w io.Writer, customParams *cpTabwriterParams) *cpTabwriter {
	p := cpTabwriterDefaultParams
	if customParams != nil {
		p = *customParams
	}
	tw := tabwriter.NewWriter(w, p.minWidth, p.tabWidth, p.padding, p.padChar, p.flags)
	return &cpTabwriter{Writer: tw}
}

// WriteTable from a 2-D slice of strings, joining every string with tabs.
//
// Tab characters and newlines will be added to the end of each rank.
func (w *cpTabwriter) WriteTable(lines [][]string) (int, error) {
	sum := 0
	var err error

	for _, line := range lines {
		// A final "\t" is added, as tabwriter is tab-terminated, not the more
		// common tab-separated. See https://pkg.go.dev/text/tabwriter#Writer
		// for details. There are cases where one might not want this trailing
		// tab, but it hasn't come up yet, and is easily worked around when
		// the time comes.
		sum, err = fmt.Fprintln(w, strings.Join(line, "\t")+"\t")
		if err != nil {
			return sum, err
		}
	}
	return sum, nil
}

// cpTabwriterParams makes it easier to reuse common tabwriter parameters.
type cpTabwriterParams struct {
	minWidth int
	tabWidth int
	padding  int
	padChar  byte
	flags    uint
}

// cpTabwriterDefaultParams is just a convenience.
var cpTabwriterDefaultParams = cpTabwriterParams{16, 8, 0, '\t', 0}
