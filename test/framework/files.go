package framework

import "os"

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return err == nil
}
