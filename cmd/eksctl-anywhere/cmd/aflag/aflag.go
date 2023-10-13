// Package aflag is the eks anywhere flag handling package.
package aflag

import (
	"github.com/spf13/pflag"
)

// Flag defines a CLI flag.
type Flag[T any] struct {
	Name    string
	Short   string
	Usage   string
	Default T
}

// String applies f to fs and writes the value to dst.
func String(f Flag[string], dst *string, fs *pflag.FlagSet) {
	switch {
	// With short form
	case f.Short != "":
		fs.StringVarP(dst, f.Name, f.Short, f.Default, f.Usage)
	// Without short form
	default:
		fs.StringVar(dst, f.Name, f.Default, f.Usage)
	}
}

// Bool applies f to fs and writes the value to dst.
func Bool(f Flag[bool], dst *bool, fs *pflag.FlagSet) {
	switch {
	case f.Short != "":
		fs.BoolVarP(dst, f.Name, f.Short, f.Default, f.Usage)
	default:
		fs.BoolVar(dst, f.Name, f.Default, f.Usage)
	}
}

// StringSlice applies f to fs and writes the value to dst.
func StringSlice(f Flag[[]string], dst *[]string, fs *pflag.FlagSet) {
	switch {
	case f.Short != "":
		fs.StringSliceVarP(dst, f.Name, f.Short, f.Default, f.Usage)
	default:
		fs.StringSliceVar(dst, f.Name, f.Default, f.Usage)
	}
}

// StringString applies f to fs and writes the value to dst.
func StringString(f Flag[map[string]string], dst *map[string]string, fs *pflag.FlagSet) {
	switch {
	case f.Short != "":
		fs.StringToStringVarP(dst, f.Name, f.Short, f.Default, f.Usage)
	default:
		fs.StringToStringVar(dst, f.Name, f.Default, f.Usage)
	}
}

// HTTPHeader applies f to fs and writes the value to dst.
func HTTPHeader(f Flag[Header], dst *Header, fs *pflag.FlagSet) {
	switch {
	case f.Short != "":
		fs.VarP(dst, f.Name, f.Short, f.Usage)
	default:
		fs.Var(dst, f.Name, f.Usage)
	}
}
