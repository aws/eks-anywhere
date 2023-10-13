package aflag

import (
	"testing"

	"github.com/spf13/pflag"
)

func TestString(t *testing.T) {
	type args struct {
		f   Flag[string]
		dst *string
		fs  *pflag.FlagSet
	}
	tests := map[string]struct {
		name string
		args args
	}{
		"success long form": {
			args: args{
				f: Flag[string]{
					Name:    "test",
					Usage:   "test usage",
					Default: "test default",
				},
				dst: new(string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
		"success short form": {
			args: args{
				f: Flag[string]{
					Name:    "test",
					Short:   "t",
					Usage:   "test usage",
					Default: "test default",
				},
				dst: new(string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			String(tt.args.f, tt.args.dst, tt.args.fs)
		})
	}
}

func TestBool(t *testing.T) {
	type args struct {
		f   Flag[bool]
		dst *bool
		fs  *pflag.FlagSet
	}
	tests := map[string]struct {
		name string
		args args
	}{
		"success long form": {
			args: args{
				f: Flag[bool]{
					Name:    "test",
					Usage:   "test usage",
					Default: true,
				},
				dst: new(bool),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
		"success short form": {
			args: args{
				f: Flag[bool]{
					Name:    "test",
					Short:   "t",
					Usage:   "test usage",
					Default: true,
				},
				dst: new(bool),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Bool(tt.args.f, tt.args.dst, tt.args.fs)
		})
	}
}

func TestStringSlice(t *testing.T) {
	type args struct {
		f   Flag[[]string]
		dst *[]string
		fs  *pflag.FlagSet
	}
	tests := map[string]struct {
		name string
		args args
	}{
		"success long form": {
			args: args{
				f: Flag[[]string]{
					Name:    "test",
					Usage:   "test usage",
					Default: []string{"test"},
				},
				dst: new([]string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
		"success short form": {
			args: args{
				f: Flag[[]string]{
					Name:    "test",
					Short:   "t",
					Usage:   "test usage",
					Default: []string{"test"},
				},
				dst: new([]string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StringSlice(tt.args.f, tt.args.dst, tt.args.fs)
		})
	}
}

func TestStringString(t *testing.T) {
	type args struct {
		f   Flag[map[string]string]
		dst *map[string]string
		fs  *pflag.FlagSet
	}
	tests := map[string]struct {
		name string
		args args
	}{
		"success long form": {
			args: args{
				f: Flag[map[string]string]{
					Name:    "test",
					Usage:   "test usage",
					Default: map[string]string{"test": "test"},
				},
				dst: new(map[string]string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
		"success short form": {
			args: args{
				f: Flag[map[string]string]{
					Name:    "test",
					Short:   "t",
					Usage:   "test usage",
					Default: map[string]string{"test": "test"},
				},
				dst: new(map[string]string),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			StringString(tt.args.f, tt.args.dst, tt.args.fs)
		})
	}
}

func TestHTTPHeader(t *testing.T) {
	type args struct {
		f   Flag[Header]
		dst *Header
		fs  *pflag.FlagSet
	}
	tests := map[string]struct {
		name string
		args args
	}{
		"success long form": {
			args: args{
				f: Flag[Header]{
					Name:    "test",
					Usage:   "test usage",
					Default: Header{},
				},
				dst: new(Header),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
		"success short form": {
			args: args{
				f: Flag[Header]{
					Name:    "test",
					Short:   "t",
					Usage:   "test usage",
					Default: Header{},
				},
				dst: new(Header),
				fs:  pflag.NewFlagSet("test", pflag.ContinueOnError),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			HTTPHeader(tt.args.f, tt.args.dst, tt.args.fs)
		})
	}
}
