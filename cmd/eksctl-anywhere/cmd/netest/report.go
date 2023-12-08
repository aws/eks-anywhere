package netest

import "github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/netest/invoker"

// Outcome is the outcome of executing a netest test suite.
type Outcome int

const (
	// Fail indicates a test suite failed.
	Fail Outcome = iota

	// Pass indicates a test suite passed.
	Pass
)

// Report is a collect of test atomic results.
type Report []TestResult

func (report *Report) Result(r TestResult) {
	*report = append(*report, r)
}

// TestResult provides detail on the result of executing an atomic test.
type TestResult struct {
	Cmd     string
	Outcome Outcome
	Error   string
}

// Passed indicates t represents a test who's outcome was Pass.
func (t TestResult) Passed() bool {
	return t.Outcome == Pass
}

func toResult(r invoker.Outcome) TestResult {
	res := TestResult{
		Cmd:     r.Cmd,
		Outcome: Pass,
	}

	if r.Error != nil {
		res.Outcome = Fail
		res.Error = r.Stderr.String()
	}

	return res
}
