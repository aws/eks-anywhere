package e2e

import (
	"testing"

	. "github.com/onsi/gomega"
)

func TestCopyCommand(t *testing.T) {
	tests := []struct {
		name string
		c    copyCommand
		want string
	}{
		{
			name: "simple upload",
			c: newCopyCommand().from(
				"/home/path", "user",
			).to("s3://bucket", "path"),
			want: "aws s3 cp /home/path/user s3://bucket/path",
		},
		{
			name: "simple download",
			c: newCopyCommand().from(
				"s3://bucket", "path",
			).to("/home/path", "user"),
			want: "aws s3 cp s3://bucket/path /home/path/user",
		},
		{
			name: "recursive upload exclude everything include file pattern",
			c: newCopyCommand().from("/home/e2e/").to(
				"s3://bucket", "path", "path2/",
			).recursive().exclude("*").include("support-bundle-*.tar.gz"),
			want: "aws s3 cp /home/e2e/ s3://bucket/path/path2/ --recursive --exclude \"*\" --include \"support-bundle-*.tar.gz\"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(tt.c.String()).To(Equal(tt.want), "String() should return the proper built s3 copy command")
		})
	}
}
