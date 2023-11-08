package files_test

import (
	"crypto/x509"
	"embed"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	"github.com/aws/eks-anywhere/internal/test"
	"github.com/aws/eks-anywhere/pkg/files"
)

//go:embed testdata
var testdataFS embed.FS

func TestReaderReadFileError(t *testing.T) {
	tests := []struct {
		testName string
		uri      string
		filePath string
	}{
		{
			testName: "missing local file",
			uri:      "fake-local-file.yaml",
		},
		{
			testName: "missing embed file",
			uri:      "embed:///fake-local-file.yaml",
		},
		{
			testName: "invalid uri",
			uri:      ":domain.com/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			r := files.NewReader()
			_, err := r.ReadFile(tt.uri)
			g.Expect(err).NotTo(BeNil())
		})
	}
}

func TestReaderReadFileSuccess(t *testing.T) {
	tests := []struct {
		testName string
		uri      string
		filePath string
	}{
		{
			testName: "local file",
			uri:      "testdata/file.yaml",
			filePath: "testdata/file.yaml",
		},
		{
			testName: "embed file",
			uri:      "embed:///testdata/file.yaml",
			filePath: "testdata/file.yaml",
		},
	}

	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			g := NewWithT(t)
			r := files.NewReader(files.WithEmbedFS(testdataFS))
			got, err := r.ReadFile(tt.uri)
			g.Expect(err).To(BeNil())
			test.AssertContentToFile(t, string(got), tt.filePath)
		})
	}
}

func TestReaderReadFileHTTPSSuccess(t *testing.T) {
	g := NewWithT(t)
	filePath := "testdata/file.yaml"

	server := test.NewHTTPSServerForFile(t, filePath)
	uri := server.URL + "/" + filePath

	r := files.NewReader(files.WithRootCACerts(serverCerts(g, server)))
	got, err := r.ReadFile(uri)
	g.Expect(err).To(BeNil())
	test.AssertContentToFile(t, string(got), filePath)
}

func TestReaderReadFileHTTPSProxySuccess(t *testing.T) {
	t.Skip("Flaky (https://github.com/aws/eks-anywhere/issues/5775)")

	g := NewWithT(t)
	filePath := "testdata/file.yaml"
	// It's important to use example.com because the certificate created for
	// the TLS server is only valid for this domain and 127.0.0.1.
	fakeServerHost := "example.com:443"
	fileURL := "https://" + fakeServerHost + "/" + filePath

	server := test.NewHTTPSServerForFile(t, filePath)
	serverHost := serverHost(g, server)
	// We need to use the proxy server to do a host "swapping".
	// The test server created by NewHTTPSServerForFile will be listening in
	// 127.0.0.1. However, the Go documentation for the transport.Proxy states that:
	// > if req.URL.Host is "localhost" or a loopback address (with or without
	// > a port number), then a nil URL and nil error will be returned.
	// https://pkg.go.dev/golang.org/x/net/http/httpproxy#Config.ProxyFunc
	// Which means that it will never honor the HTTPS_PROXY env var since our
	// request will be pointing to a loopback address.
	// In order to make it work, we pass example.com in our request and use the
	// proxy to map this domain to 127.0.0.1, where our file server is listening.
	hostMappings := map[string]string{fakeServerHost: serverHost}
	proxy := newProxyServer(t, hostMappings)

	os.Unsetenv("HTTP_PROXY")
	t.Setenv("HTTPS_PROXY", proxy.URL)

	r := files.NewReader(
		files.WithRootCACerts(serverCerts(g, server)),
		files.WithNonCachedProxyConfig(),
	)

	got, err := r.ReadFile(fileURL)
	g.Expect(err).To(BeNil())
	test.AssertContentToFile(t, string(got), filePath)

	g.Expect(proxy.countForHost(serverHost)).To(
		Equal(1), "Host %s should have been proxied exactly once", serverHost,
	)
}

func serverCerts(g Gomega, server *httptest.Server) []*x509.Certificate {
	certs := []*x509.Certificate{}
	for _, c := range server.TLS.Certificates {
		roots, err := x509.ParseCertificates(c.Certificate[len(c.Certificate)-1])
		g.Expect(err).NotTo(HaveOccurred())
		certs = append(certs, roots...)
	}

	return certs
}

func serverHost(g Gomega, server *httptest.Server) string {
	u, err := url.Parse(server.URL)
	g.Expect(err).NotTo(HaveOccurred())
	return u.Host
}

type proxyServer struct {
	*httptest.Server
	*proxy
}

func newProxyServer(tb testing.TB, hostMappings map[string]string) *proxyServer {
	proxyServer := &proxyServer{
		proxy: newProxy(hostMappings, tb),
	}
	proxyServer.Server = httptest.NewServer(http.HandlerFunc(proxyServer.handleProxy))

	tb.Cleanup(func() {
		proxyServer.Close()
	})

	return proxyServer
}

type proxy struct {
	sync.Mutex
	// proxied maintains a count of how many proxied requests
	// have been completed per host.
	proxied map[string]int
	// hostMappings allows to map the dst host in the CONNECT
	// request to a different host.
	hostMappings map[string]string
	tb           testing.TB
}

func newProxy(hostMappings map[string]string, tb testing.TB) *proxy {
	return &proxy{
		proxied:      map[string]int{},
		hostMappings: hostMappings,
		tb:           tb,
	}
}

func (p *proxy) handleProxy(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodConnect {
		p.tunnelConnection(w, r)
	} else {
		http.Error(w, "Only supports CONNECT", http.StatusMethodNotAllowed)
	}
}

func (p *proxy) tunnelConnection(w http.ResponseWriter, r *http.Request) {
	host := r.Host
	if mappedDstHost, ok := p.hostMappings[host]; ok {
		host = mappedDstHost
	}
	destConn, err := net.DialTimeout("tcp", host, 2*time.Second)
	if err != nil {
		http.Error(w, fmt.Sprintf("opening TCP connection to proxied host %s: %s", host, err), http.StatusServiceUnavailable)
		return
	}
	defer destConn.Close()

	h, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking is not supported", http.StatusInternalServerError)
		return
	}

	// This might be too early if we can't hijack the connection
	// But we can't use ResponseWriter after the hijack
	// and doing it manually is too difficult
	w.WriteHeader(http.StatusOK)

	clientConn, hijackedConnData, err := h.Hijack()
	if err != nil {
		http.Error(w, fmt.Sprintf("hijacking the connection in proxy: %s", err), http.StatusServiceUnavailable)
		return
	}
	defer clientConn.Close()

	// Read more from the client. Include the connection buffer if
	// it contains data.
	orgReader := io.Reader(clientConn)
	if hijackedConnData.Reader.Buffered() > 0 {
		p.tb.Logf("Looks like there is unread data in the buffer: %d bytes", hijackedConnData.Reader.Buffered())
		orgReader = io.MultiReader(hijackedConnData, clientConn)
	}

	if hijackedConnData.Reader.Buffered() > 0 {
		if _, err := io.Copy(destConn, hijackedConnData); err != nil {
			p.tb.Errorf("Error writing client unprocessed data: %s", err)
			return
		}
	}

	p.countRequest(host)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		p.pipe(destConn, orgReader)
		wg.Done()
	}()
	go func() {
		p.pipe(clientConn, destConn)
		wg.Done()
	}()

	wg.Wait()
}

// countRequest increases the proxied counter for the given host.
func (p *proxy) countRequest(host string) {
	p.Lock()
	defer p.Unlock()

	p.proxied[host] = p.proxied[host] + 1
}

// countForHost returns the number of time a particular host has been proxied.
func (p *proxy) countForHost(host string) int {
	p.Lock()
	defer p.Unlock()

	return p.proxied[host]
}

func (p *proxy) pipe(destination io.Writer, source io.Reader) {
	if _, err := io.Copy(destination, source); err != nil {
		p.tb.Errorf("Error piping: %s", err)
	}
}
