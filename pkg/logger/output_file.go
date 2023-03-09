package logger

import "sync"

// This source file uses package state to configure output file paths for the bundle to retrieve.
// The relationship between the 2 components is undesirable and will be refactored to pass the
// path to bundle generation code explicitly.
//
// Please avoid using the GetOutputFilePath() function in new code.

var (
	outputFilePath    string
	outputFilePathMtx sync.Mutex
)

func setOutputFilePath(path string) {
	outputFilePathMtx.Lock()
	defer outputFilePathMtx.Unlock()
	outputFilePath = path
}

// GetOutputFilePath returns the path to the file where high verbosity logs are written to.
// If the logger hasn't been configured to output to a file, it returns an empty string.
//
// Deprecated: The function will be removed to avoid using package state.
func GetOutputFilePath() string {
	outputFilePathMtx.Lock()
	defer outputFilePathMtx.Unlock()
	return outputFilePath
}
