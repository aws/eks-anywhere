package ecr

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	. "github.com/onsi/gomega"
)

func TestGetLatestOCIShaTag(t *testing.T) {
	now := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	image := func(tag string, pushed time.Time) *ecr.ImageDetail {
		return &ecr.ImageDetail{
			ImageTags:     []*string{aws.String(tag)},
			ImageDigest:   aws.String("sha256:fixture"),
			ImagePushedAt: aws.Time(pushed),
		}
	}
	tests := []struct {
		name    string
		images  []*ecr.ImageDetail
		branch  string
		helm    bool
		wantTag string
	}{
		{name: "empty repository", branch: "release-0.27", helm: true},
		{
			name: "no chart for new release branch",
			images: []*ecr.ImageDetail{
				image("0.0.0-dev-main", now),
				image("0.0.0-dev-release-0.26", now.Add(time.Minute)),
			},
			branch: "release-0.27", helm: true,
		},
		{
			name:   "main rejects release charts",
			images: []*ecr.ImageDetail{image("0.0.0-dev-release-0.26", now)},
			branch: "main", helm: true,
		},
		{
			name:   "untagged image",
			images: []*ecr.ImageDetail{{ImageDigest: aws.String("sha256:fixture"), ImagePushedAt: aws.Time(now)}},
			branch: "release-0.27", helm: true,
		},
		{
			name:   "missing digest",
			images: []*ecr.ImageDetail{{ImageTags: []*string{aws.String("0.0.0-dev-release-0.27")}, ImagePushedAt: aws.Time(now)}},
			branch: "release-0.27", helm: true,
		},
		{
			name:   "missing push time",
			images: []*ecr.ImageDetail{{ImageTags: []*string{aws.String("0.0.0-dev-release-0.27")}, ImageDigest: aws.String("sha256:fixture")}},
			branch: "release-0.27", helm: true,
		},
		{
			name: "newest matching chart ignores newer other branches",
			images: []*ecr.ImageDetail{
				image("0.0.0-dev-release-0.27-old", now),
				image("0.0.0-dev-release-0.27-new", now.Add(time.Minute)),
				image("0.0.0-dev-main", now.Add(2*time.Minute)),
			},
			branch: "release-0.27", helm: true, wantTag: "0.0.0-dev-release-0.27-new",
		},
		{
			name: "main selects newest non-release chart",
			images: []*ecr.ImageDetail{
				image("0.0.0-dev-main", now),
				image("0.0.0-dev-release-0.26", now.Add(time.Minute)),
			},
			branch: "main", helm: true, wantTag: "0.0.0-dev-main",
		},
		{
			name: "container image selection remains branch independent",
			images: []*ecr.ImageDetail{
				image("v0.0.0-dev-main", now),
				image("v0.0.0-dev-release-0.26", now.Add(time.Minute)),
			},
			branch: "release-0.27", wantTag: "v0.0.0-dev-release-0.26",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			tag, digest, err := getLatestOCIShaTag(tt.images, tt.branch, tt.helm)
			if tt.wantTag == "" {
				g.Expect(err).To(MatchError(ContainSubstring("no images found")))
				g.Expect(tag).To(BeEmpty())
				g.Expect(digest).To(BeEmpty())
				return
			}
			g.Expect(err).NotTo(HaveOccurred())
			g.Expect(tag).To(Equal(tt.wantTag))
			g.Expect(digest).To(Equal("sha256:fixture"))
		})
	}
}

func TestFilterECRRepoByTagPrefixNoMatchingBranch(t *testing.T) {
	g := NewWithT(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		g.Expect(r.Header.Get("X-Amz-Target")).To(HaveSuffix(".DescribeImages"))
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		_, err := w.Write([]byte(`{"imageDetails":[{"imageDigest":"sha256:fixture","imagePushedAt":1704067200,"imageTags":["0.0.0-dev-main","v0.0.0-dev-main"]}]}`))
		g.Expect(err).NotTo(HaveOccurred())
	}))
	defer server.Close()
	s := session.Must(session.NewSession(aws.NewConfig().
		WithRegion("us-west-2").
		WithEndpoint(server.URL).
		WithCredentials(credentials.NewStaticCredentials("fixture", "fixture", ""))))
	tag, digest, err := FilterECRRepoByTagPrefix(ecr.New(s), "fixture-packages", "0.0.0", "release-0.27", true)
	g.Expect(err).To(MatchError(`no images found for branch "release-0.27"`))
	g.Expect(tag).To(BeEmpty())
	g.Expect(digest).To(BeEmpty())
}
