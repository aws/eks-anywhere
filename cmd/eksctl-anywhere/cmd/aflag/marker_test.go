package aflag_test

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aws/eks-anywhere/cmd/eksctl-anywhere/cmd/aflag"
)

func TestMarkRequired(t *testing.T) {
	type flag struct {
		Pflag   pflag.Flag
		Require bool
	}

	for _, tc := range []struct {
		Name        string
		Flags       []flag
		Args        []string
		ExpectError string
	}{
		{
			Name:  "RequiredFlagPresent",
			Flags: []flag{{Pflag: nopPflag("foo"), Require: true}},
			Args:  []string{"--foo=."},
		},
		{
			Name:        "RequiredFlagNotPresent",
			Flags:       []flag{{Pflag: nopPflag("foo"), Require: true}},
			ExpectError: "required flag(s) \"foo\" not set",
		},
		{
			Name: "MultipleRequiredFlagsPresent",
			Flags: []flag{
				{Pflag: nopPflag("foo"), Require: true},
				{Pflag: nopPflag("bar"), Require: true},
			},
			Args: []string{"--foo=.", "--bar=."},
		},
		{
			Name: "MultipleRequiredFlagsNotPresent",
			Flags: []flag{
				{Pflag: nopPflag("foo"), Require: true},
				{Pflag: nopPflag("bar"), Require: true},
			},
			// A bug in cobra causes
			ExpectError: "required flag(s) \"bar\", \"foo\" not set",
		},
		{
			Name: "NoRequiredFlags",
			Flags: []flag{
				{Pflag: nopPflag("foo")},
				{Pflag: nopPflag("bar")},
			},
			Args: []string{"--foo=."},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			cmd := &cobra.Command{}

			// Cache required flags so we can call the function under test.
			var required []string

			// Add all flags.
			for _, flg := range tc.Flags {
				flg := flg
				cmd.Flags().AddFlag(&flg.Pflag)

				if flg.Require {
					required = append(required, flg.Pflag.Name)
				}
			}

			// The cmd.ValidateRequiredFlags() operates on the flag set as opposed to cmd internals
			// so we need to parse the args using the flag set before calling it.
			_ = cmd.Flags().Parse(tc.Args)

			aflag.MarkRequired(cmd.Flags(), required...)

			err := cmd.ValidateRequiredFlags()

			// We expect an error but we received nothing.
			if tc.ExpectError != "" && err == nil {
				t.Error("Expected error but received nil")
			}

			// We don't expect an error but we received something.
			if tc.ExpectError == "" && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// We expect an error but its not what we expected.
			if tc.ExpectError != "" && tc.ExpectError != err.Error() {
				t.Errorf("Expected: %v\nReceived: %v\n", tc.ExpectError, err)
			}
		})
	}
}

func TestMarkRequired_FlagDoesNotExist(t *testing.T) {
	defer func() {
		// We expect the panic called in flags.MarkRequired() to pass a non-nil value so we can
		// test for that here.
		if recover() == nil {
			t.Error("no panic received")
		}
	}()

	flgs := pflag.NewFlagSet("", pflag.ContinueOnError)
	aflag.MarkRequired(flgs, "does-not-exist")
}

func nopPflag(name string) pflag.Flag {
	return pflag.Flag{
		Name:        name,
		Value:       nopValue{},
		Annotations: map[string][]string{},
	}
}

type nopValue struct{}

func (nopValue) String() string   { return "" }
func (nopValue) Set(string) error { return nil }
func (nopValue) Type() string     { return "" }

func TestMarkHidden(t *testing.T) {
	tests := map[string]struct {
		flags       []string
		hidden      []string
		shouldPanic bool
	}{
		"success":             {flags: []string{"foo", "bar"}, hidden: []string{"foo", "bar"}},
		"flag does not exist": {flags: []string{"foo", "bar"}, hidden: []string{"foo", "bar", "baz"}, shouldPanic: true},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.shouldPanic {
				defer func() {
					if recover() == nil {
						t.Error("no panic received")
					}
				}()
			}
			cmd := &cobra.Command{}
			for _, flag := range tc.flags {
				cmd.Flags().AddFlag(&pflag.Flag{Name: flag})
			}
			aflag.MarkHidden(cmd.Flags(), tc.hidden...)
			for _, flag := range tc.flags {
				if !cmd.Flags().Lookup(flag).Hidden {
					t.Errorf("flag %s should be hidden", flag)
				}
			}
		})
	}
}
