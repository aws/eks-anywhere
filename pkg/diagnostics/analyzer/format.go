package analyzer

import (
	"fmt"
	"regexp"
)

func resourceStatusMessage(sev Severity, resourceType, name, namespace, status, message string) string {
	m := fmt.Sprintf("%s %s %s %s", blue(resourceType), bold(fmt.Sprintf("%s/%s", namespace, name)), verb(status), formatStatus(sev, status))
	if message != "" {
		m = fmt.Sprintf("%s: %s", m, message)
	}
	return m
}

func actionStatusMessage(sev Severity, resourceType, name, status, image string) string {
	return fmt.Sprintf("%s %s %s %s (image=%s)", blue(resourceType), bold(name), verb(status), formatStatus(sev, status), green(image))
}

type logHighlighter struct {
	regex *regexp.Regexp
	color func(string) string
}

var logHighlighters = []logHighlighter{
	{regex: regexp.MustCompile(`\b-?\d+(\.\d+)?\b`), color: magenta},            // numbers not surrounded by letters
	{regex: regexp.MustCompile(`^I\d{4}`), color: blue},                         // Kubernetes info logs header
	{regex: regexp.MustCompile(`^W\d{4}`), color: yellow},                       // Kubernetes warning logs header
	{regex: regexp.MustCompile(`^E\d{4}`), color: red},                          // Kubernetes error logs
	{regex: regexp.MustCompile(`^F\d{4}`), color: red},                          // Kubernetes fatal logs header
	{regex: regexp.MustCompile(`"([^"]*)"`), color: green},                      // everything in quotes
	{regex: regexp.MustCompile(`\b\d{2}:\d{2}:\d{2}\.\d{6}\b`), color: magenta}, // Date in format 11:39:04.577952
}

var colorSequence = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func formatLog(log string) string {
	for _, h := range logHighlighters {
		log = h.regex.ReplaceAllStringFunc(log, func(s string) string {
			return h.color(colorSequence.ReplaceAllString(s, ""))
		})
	}

	return log
}

func formatStatus(sev Severity, status string) string {
	var c func(string) string
	if sev == SeverityError {
		c = red
	} else {
		c = yellow
	}

	return bold(c(status))
}

func verb(status string) string {
	if status == "failed" {
		return "has"
	}
	return "is"
}
