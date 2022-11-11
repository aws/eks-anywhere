package fileutils

import (
	"strings"
	"time"
)

func GenOutputDirName(suffix string) string {
	now := time.Now().Format(time.RFC3339)
	// Replace : characters with _ for easier double-click selection in a
	// terminal.
	prefix := strings.ReplaceAll(now, ":", "_")
	return prefix + "-" + suffix
}
