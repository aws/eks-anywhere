package flags

import "github.com/spf13/pflag"

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
