package e2e

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/executables"
	"github.com/aws/eks-anywhere/pkg/types"
)

func listTests(regex string, testsToSkip []string) (tests, skippedTests []string, err error) {
	e := executables.NewExecutable(filepath.Join("bin", e2eBinary))
	ctx := context.Background()
	testResponse, err := e.Execute(ctx, "-test.list", regex)
	if err != nil {
		return nil, nil, fmt.Errorf("failed listing test from e2e binary: %v", err)
	}

	skipLookup := types.SliceToLookup(testsToSkip)
	scanner := bufio.NewScanner(&testResponse)
	for scanner.Scan() {
		line := scanner.Text()
		if skipLookup.IsPresent(line) {
			skippedTests = append(skippedTests, line)
			continue
		}

		if strings.HasPrefix(line, "Test") {
			tests = append(tests, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("failed reading e2e list response: %v", err)
	}

	return tests, skippedTests, nil
}
