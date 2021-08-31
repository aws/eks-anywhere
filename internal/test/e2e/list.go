package e2e

import (
	"bufio"
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/aws/eks-anywhere/pkg/executables"
)

func listTests(regex string) ([]string, error) {
	e := executables.NewExecutable(filepath.Join("bin", e2eBinary))
	ctx := context.Background()
	testReponse, err := e.Execute(ctx, "-test.list", regex)
	if err != nil {
		return nil, fmt.Errorf("failed listing test from e2e binary: %v", err)
	}

	tests := make([]string, 0)

	scanner := bufio.NewScanner(&testReponse)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Test") {
			tests = append(tests, line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed reading e2e list response: %v", err)
	}

	return tests, nil
}
