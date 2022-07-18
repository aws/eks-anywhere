package test

import (
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func NewNullLogger() logr.Logger {
	return logr.New(log.NullLogSink{})
}
