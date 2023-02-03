package aws_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/pkg/aws"
)

type metadataTest struct {
	*WithT
	client *aws.Client
	ts     *httptest.Server
}

func newMetadataTest(t *testing.T) *metadataTest {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(func() { ts.Close() })
	return &metadataTest{
		WithT:  NewWithT(t),
		client: aws.NewClient(),
		ts:     ts,
	}
}

func TestGenerateIMDSv2SessionToken(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.GenerateIMDSv2SessionToken(g.ts.URL)
	g.Expect(err).To(Succeed())
}

func TestGenerateIMDSv2SessionTokenCreatePostError(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.GenerateIMDSv2SessionToken(strings.TrimPrefix(g.ts.URL, "http://"))
	g.Expect(err).To(MatchError(ContainSubstring("creating http POST request for generating IMDSv2 session")))
}

func TestGenerateIMDSv2SessionTokenPostRunError(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.GenerateIMDSv2SessionToken("invalid url")
	g.Expect(err).To(MatchError(ContainSubstring("generating IMDSv2 session token from IMDSv2 url")))
}

func TestInstanceIP(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.InstanceIP(g.ts.URL, "")
	g.Expect(err).To(Succeed())
}

func TestInstanceIPCreateGetError(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.InstanceIP(strings.TrimPrefix(g.ts.URL, "http://"), "")
	g.Expect(err).To(MatchError(ContainSubstring("creating http GET request for fetching instance IP through IMDSv2")))
}

func TestInstanceIPGetRunError(t *testing.T) {
	g := newMetadataTest(t)
	_, err := g.client.InstanceIP("invalid url", "")
	g.Expect(err).To(MatchError(ContainSubstring("fetching instance IP from IMDSv2 url")))
}
