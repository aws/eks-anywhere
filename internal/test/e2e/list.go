package e2e

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/types"
	e2etest "github.com/aws/eks-anywhere/test/e2e"
)

// regex will filter out the tests that match the regex, but it does't see subtests.
// TestsToSelect and testsToSkip will filter out all test cases including subtests.
func listTests(regex string, testsToSelect []string, testsToSkip []string) (tests, skippedTests []string, err error) {
	e := executables.NewExecutable(filepath.Join("bin", e2eBinary))
	ctx := context.Background()
	testResponse, err := e.Execute(ctx, "-test.list", regex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed listing test from e2e binary: %v", err)
	}

	skipLookup := types.SliceToLookup(testsToSkip)
	scanner := bufio.NewScanner(&testResponse)
	rawTestList := processTestListOutput(testResponse)

	for _, t := range rawTestList {
		selected := len(testsToSelect) == 0
		for _, s := range testsToSelect {
			re := regexp.MustCompile(s)
			if re.MatchString(t) {
				selected = true
				break
			}
		}
		if !selected {
			continue
		}
		if skipLookup.IsPresent(t) {
			skippedTests = append(skippedTests, t)
			continue
		}

		tests = append(tests, t)
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed reading e2e list response: %v", err)
	}

	return tests, skippedTests, nil
}

// parse the output of -test.list, and add subtests.
func processTestListOutput(b bytes.Buffer) (tests []string) {
	s := bufio.NewScanner(&b)
	for s.Scan() {
		line := s.Text()
		if !strings.HasPrefix(line, "Test") {
			continue
		}

		if !strings.HasSuffix(line, "Suite") {
			tests = append(tests, line)
			continue
		}

		if strings.HasSuffix(line, "Suite") {
			for k, s := range e2etest.Suites {
				if strings.HasSuffix(line, k) {
					for _, st := range s {
						tests = append(tests, line+"/"+st.GetName())
					}
					break
				}
			}
		}
	}
	return tests
}
