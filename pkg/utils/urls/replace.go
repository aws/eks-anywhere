package urls

import (
	"net/url"
	"path/filepath"
	"strings"
)

// ReplaceHost replaces the host in a url
// It supports full URLs and container image URLs
// If the provided original url is malformed, the are no guarantees
// that the returned value will be valid
// If host is empty, it will return the original URL.
func ReplaceHost(orgURL, replace string) string {
	if replace == "" {
		return orgURL
	}

	o := tryParse(orgURL)
	if o == nil {
		return orgURL
	}

	r := tryParse(replace)
	if r == nil {
		return orgURL
	}

	o.Host = r.Host

	if r.Path != "" {
		o.Path = filepath.Join(r.Path, o.Path)
	}

	if r.Scheme != "" {
		o.Scheme = r.Scheme
	}

	return strings.TrimPrefix(o.String(), "//")
}

func tryParse(in string) *url.URL {
	u, err := url.Parse(in)
	if err != nil || u.Scheme == "" {
		u, err = url.Parse("//" + in)
		if err != nil {
			return nil
		}

		u.Scheme = ""
	}

	return u
}
