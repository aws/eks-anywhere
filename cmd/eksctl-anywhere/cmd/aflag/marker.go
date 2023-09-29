package aflag

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// MarkRequired is a helper to mark flags required on cmd. If a flag does not exist, it panics.
func MarkRequired(set *pflag.FlagSet, flags ...string) {
	for _, flag := range flags {
		if err := cobra.MarkFlagRequired(set, flag); err != nil {
			panic(err)
		}
	}
}

// MarkHidden is a helper to mark flags hidden on cmd. If a flag does not exist, it panics.
func MarkHidden(set *pflag.FlagSet, flags ...string) {
	for _, flag := range flags {
		if err := set.MarkHidden(flag); err != nil {
			panic(err)
		}
	}
}
