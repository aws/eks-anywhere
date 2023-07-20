package files

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"golang.org/x/net/http/httpproxy"
)

const (
	httpsScheme = "https"
	embedScheme = "embed"
)

type Reader struct {
	embedFS    embed.FS
	httpClient *http.Client
	userAgent  string
}

type ReaderOpt func(*Reader)

func WithEmbedFS(embedFS embed.FS) ReaderOpt {
	return func(s *Reader) {
		s.embedFS = embedFS
	}
}

func WithUserAgent(userAgent string) ReaderOpt {
	return func(s *Reader) {
		s.userAgent = userAgent
	}
}

// WithEKSAUserAgent sets the user agent for a particular eks-a component and version.
// component should be something like "cli", "controller", "e2e", etc.
// version should generally be a semver, but when not available, any string is valid.
func WithEKSAUserAgent(eksAComponent, version string) ReaderOpt {
	return WithUserAgent(eksaUserAgent(eksAComponent, version))
}

// WithRootCACerts configures the HTTP client's trusted CAs. Note that this will overwrite
// the defaults so the host's trust will be ignored. This option is only for testing.
func WithRootCACerts(certs []*x509.Certificate) ReaderOpt {
	return func(r *Reader) {
		t := r.httpClient.Transport.(*http.Transport)
		if t.TLSClientConfig == nil {
			t.TLSClientConfig = &tls.Config{}
		}

		if t.TLSClientConfig.RootCAs == nil {
			t.TLSClientConfig.RootCAs = x509.NewCertPool()
		}

		for _, c := range certs {
			t.TLSClientConfig.RootCAs.AddCert(c)
		}
	}
}

// WithNonCachedProxyConfig configures the HTTP client to read the Proxy configuration
// from the environment on every request instead of relying on the default package
// level cache (implemented in the http package with envProxyFuncValue), which is only
// read once. If Proxy is not configured in the client's transport, nothing is changed.
// This is only for testing.
func WithNonCachedProxyConfig() ReaderOpt {
	return func(r *Reader) {
		t := r.httpClient.Transport.(*http.Transport)
		if t.Proxy == nil {
			return
		}

		t.Proxy = func(r *http.Request) (*url.URL, error) {
			return httpproxy.FromEnvironment().ProxyFunc()(r.URL)
		}
	}
}

func NewReader(opts ...ReaderOpt) *Reader {
	// In order to modify the TLSHandshakeTimeout we first clone the default transport.
	// It has some defaults that we want to preserve. In particular Proxy, which is set
	// to http.ProxyFromEnvironment. This will make the client honor the HTTP_PROXY,
	// HTTPS_PROXY and NO_PROXY env variables.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSHandshakeTimeout = 60 * time.Second
	client := &http.Client{
		Transport: transport,
	}

	r := &Reader{
		embedFS:    embedFS,
		httpClient: client,
		userAgent:  eksaUserAgent("unknown", "no-version"),
	}

	for _, o := range opts {
		o(r)
	}

	return r
}

func (r *Reader) ReadFile(uri string) ([]byte, error) {
	url, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("can't build cluster spec, invalid release manifest url: %v", err)
	}

	switch url.Scheme {
	case httpsScheme:
		return r.readHttpFile(uri)
	case embedScheme:
		return r.readEmbedFile(url)
	default:
		return readLocalFile(uri)
	}
}

func (r *Reader) readHttpFile(uri string) ([]byte, error) {
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, fmt.Errorf("failed creating http GET request for downloading file: %v", err)
	}

	request.Header.Set("User-Agent", r.userAgent)
	resp, err := r.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed reading file from url [%s]: %v", uri, err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed reading file from url [%s]: %v", uri, err)
	}

	return data, nil
}

func (r *Reader) readEmbedFile(url *url.URL) ([]byte, error) {
	data, err := r.embedFS.ReadFile(strings.TrimPrefix(url.Path, "/"))
	if err != nil {
		return nil, fmt.Errorf("failed reading embed file [%s] for cluster spec: %v", url.Path, err)
	}

	return data, nil
}

func readLocalFile(filename string) ([]byte, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed reading local file [%s]: %v", filename, err)
	}

	return data, nil
}

func eksaUserAgent(eksAComponent, version string) string {
	return fmt.Sprintf("eks-a-%s/%s", eksAComponent, version)
}
