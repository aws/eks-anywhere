package supportbundle_test

import (
	"testing"

	support "github.com/aws/eks-anywhere/pkg/support"
)

func TestParseTimeOptions(t *testing.T) {
	type args struct {
		since     string
		sinceTime string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Without time options",
			args: args{
				since:     "",
				sinceTime: "",
			},
			wantErr: false,
		},
		{
			name: "Good since options",
			args: args{
				since:     "8h",
				sinceTime: "",
			},
			wantErr: false,
		},
		{
			name: "Good since time options",
			args: args{
				since:     "",
				sinceTime: "2021-06-28T15:04:05Z",
			},
			wantErr: false,
		},
		{
			name: "Duplicate time options",
			args: args{
				since:     "8m",
				sinceTime: "2021-06-28T15:04:05Z",
			},
			wantErr: true,
		},
		{
			name: "Wrong since time options",
			args: args{
				since:     "",
				sinceTime: "2021-06-28T15:04:05Z07:00",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := support.ParseTimeOptions(tt.args.since, tt.args.sinceTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseTimeOptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestParseBundleFromDoc(t *testing.T) {
	type args struct {
		bundleConfig string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Good bundle config",
			args: args{
				bundleConfig: "testdata/support-bundle-test1.yaml",
			},
			wantErr: false,
		},
		{
			name: "Wrong bundle config",
			args: args{
				bundleConfig: "testdata/support-bundle-test2.yaml",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := support.ParseBundleFromDoc(tt.args.bundleConfig)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBundleFromDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
