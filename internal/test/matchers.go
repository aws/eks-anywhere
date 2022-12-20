package test

import (
	"context"
	"fmt"
	"reflect"

	"github.com/golang/mock/gomock"
)

type ofType struct{ t string }

func OfType(t string) gomock.Matcher {
	return &ofType{t}
}

func (o *ofType) Matches(x interface{}) bool {
	return reflect.TypeOf(x).String() == o.t
}

func (o *ofType) String() string {
	return "is of type " + o.t
}

// AContext returns a gomock matchers that evaluates if the receive value can
// fullfills the context.Context interface.
func AContext() gomock.Matcher {
	ctxInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	return gomock.AssignableToTypeOf(ctxInterface)
}

type bytesMatchFile struct{ file string }

// MatchFile returns a gomock matcher that compares []byte input to the content
// of the given file.
func MatchFile(file string) gomock.Matcher {
	return &bytesMatchFile{file: file}
}

func (o *bytesMatchFile) Matches(x interface{}) bool {
	content, ok := x.([]byte)
	if !ok {
		return false
	}

	equal, err := contentEqualToFile(content, o.file)
	if err != nil {
		return false
	}

	return equal
}

func (o *bytesMatchFile) String() string {
	return "matches content of " + o.file
}

func (o *bytesMatchFile) Got(got interface{}) string {
	content, ok := got.([]byte)
	if !ok {
		return "got is not a []byte"
	}

	diff, err := computeDiffBetweenContentAndFile(content, o.file)
	if err != nil {
		return fmt.Sprintf("failed computing diff with file %s: %s", o.file, err)
	}

	return fmt.Sprintf("is not equal to file %s:\n%s", o.file, diff)
}
