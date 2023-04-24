package console_test

import (
	"bytes"
	"testing"

	"github.com/aws/eks-anywhere/pkg/console"
)

func TestConfirm(t *testing.T) {
	tests := []struct {
		Name   string
		Input  string
		Result bool
	}{
		{
			Name:   "UpperY",
			Input:  "Y",
			Result: true,
		},
		{
			Name:  "LowerY",
			Input: "y",
		},
		{
			Name:  "UpperN",
			Input: "y",
		},
		{
			Name:  "WordStartingWithUpperY",
			Input: "Yellow",
		},
		{
			Name:  "WordStartingWithLowerY",
			Input: "yellow",
		},
		{
			Name:  "Word",
			Input: "foobar",
		},
		{
			Name:  "HighUTF8",
			Input: "√",
		},
		{
			Name:  "WordHighUTF8",
			Input: "√¥†˙˚œ˚",
		},
	}

	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			var input, output bytes.Buffer

			input.WriteString(tc.Input)

			result := console.Confirm("My foo bar question?", &output, &input)

			switch {
			case tc.Result && !result:
				t.Fatal("Expected true, received false")
			case !tc.Result && result:
				t.Fatal("Expected false, but got true")
			}
		})
	}
}
