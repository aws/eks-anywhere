package analyzer

// Severity represents the severity of a finding.
type Severity string

const (
	// SeverityError indicates that the finding is an error.
	SeverityError Severity = "error"
	// SeverityWarning indicates that the finding is a warning.
	SeverityWarning Severity = "warning"
)

// Finding is something that is not totally right in the cluster.
type Finding struct {
	// Severity indicates if the finding is an error or not.
	// Even if it's not an error, something could be broken.
	// This just means that at this level the finding cannot determine it.
	Severity Severity
	Message  string
	Findings []Finding
	// Logs contains logs that are relevant to the finding. They might not explain the source of the issue
	// but they can either help troubleshoot or they were a source of information for the finding.
	Logs []Log
	// Recommendation gives a recommendation on how to fix the finding.
	Recommendation string
}

// Log contains log lines from a particular source.
type Log struct {
	Source string
	Lines  []string
}
