package files

import (
	"embed"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	httpsScheme = "https"
	httpScheme  = "http"
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

func NewReader(opts ...ReaderOpt) *Reader {
	r := &Reader{
		embedFS:    embed.FS{},
		httpClient: &http.Client{},
		userAgent:  "eks-a/unknown",
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
	case httpScheme:
		return r.readHttpFile(uri)
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

	data, err := ioutil.ReadAll(resp.Body)
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
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed reading local file [%s] for cluster spec: %v", filename, err)
	}

	return data, nil
}
