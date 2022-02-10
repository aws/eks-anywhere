package logfetcher

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/aws/eks-anywhere/pkg/logger"
)

const (
	Reset  = "\033[0m"
	Black  = "\033[30m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Grey   = "\033[37m"
	White  = "\033[97m"
)

type colorer func(string) string

var colorsForRegexp = []struct {
	regex   *regexp.Regexp
	colorer colorer
}{
	{
		// not very relevant CLI logs
		regex:   regexp.MustCompile(`^.*\s*V(4|5|6|7|8|9)\s{1}[^e]`),
		colorer: black,
	},
	{
		// e2e test logs
		regex:   regexp.MustCompile(`^.*\s*V\d\s*e2e`),
		colorer: blue,
	},
	{
		// Go test logs
		regex:   regexp.MustCompile(`^.*\.go:\d*:`),
		colorer: blue,
	},
	{
		// Go test start
		regex:   regexp.MustCompile("^=== RUN "),
		colorer: green,
	},
	{
		// CLI warning
		regex:   regexp.MustCompile(`^.*\s*V\d\s*Warning:`),
		colorer: yellow,
	},
	{
		// CLI error
		regex:   regexp.MustCompile("^Error:"),
		colorer: red,
	},
	{
		// Go test failure
		regex:   regexp.MustCompile("^--- FAIL:|^FAIL"),
		colorer: red,
	},
}

func logTest(testName string, logs []*cloudwatchlogs.OutputLogEvent) error {
	logger.Info("Test logs", "testName", testName)
	for _, e := range logs {
		m := *e.Message
		for _, line := range strings.Split(m, "\n") {
			for _, rc := range colorsForRegexp {
				if rc.regex.Match([]byte(line)) {
					line = rc.colorer(line)
					break
				}
			}

			fmt.Println(line)
		}
	}

	return nil
}

func color(m, c string) string {
	return c + m + Reset
}

func blue(m string) string {
	return color(m, Blue)
}

func red(m string) string {
	return color(m, Red)
}

func green(m string) string {
	return color(m, Green)
}

func yellow(m string) string {
	return color(m, Yellow)
}

func black(m string) string {
	return color(m, Black)
}
