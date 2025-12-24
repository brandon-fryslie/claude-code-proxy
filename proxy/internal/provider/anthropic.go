package provider

import (
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/seifghazi/claude-code-monitor/internal/config"
)

type AnthropicProvider struct {
	name   string
	client *http.Client
	config *config.ProviderConfig
}

func NewAnthropicProvider(name string, cfg *config.ProviderConfig) Provider {
	return &AnthropicProvider{
		name: name,
		client: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes timeout
		},
		config: cfg,
	}
}

func (p *AnthropicProvider) Name() string {
	return p.name
}

func (p *AnthropicProvider) ForwardRequest(ctx context.Context, originalReq *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	proxyReq := originalReq.Clone(ctx)

	// Parse the configured base URL
	baseURL, err := url.Parse(p.config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL '%s': %w", p.config.BaseURL, err)
	}

	if baseURL.Scheme == "" || baseURL.Host == "" {
		return nil, fmt.Errorf("invalid base URL, scheme and host are required: %s", p.config.BaseURL)
	}

	// Update the destination URL
	proxyReq.URL.Scheme = baseURL.Scheme
	proxyReq.URL.Host = baseURL.Host
	proxyReq.URL.Path = path.Join(baseURL.Path, originalReq.URL.Path)

	// Preserve query parameters
	proxyReq.URL.RawQuery = originalReq.URL.RawQuery

	// Update request headers
	proxyReq.RequestURI = ""
	proxyReq.Host = baseURL.Host

	// Remove hop-by-hop headers
	removeHopByHopHeaders(proxyReq.Header)

	// Add required headers if not present
	version := p.config.Version
	if version == "" {
		version = "2023-06-01" // Default Anthropic API version
	}
	if proxyReq.Header.Get("anthropic-version") == "" {
		proxyReq.Header.Set("anthropic-version", version)
	}

	// If this provider has its own API key, use it (override the original)
	if p.config.APIKey != "" {
		proxyReq.Header.Set("x-api-key", p.config.APIKey)
	}

	// Support gzip encoding
	proxyReq.Header.Set("Accept-Encoding", "gzip")

	// Forward the request
	resp, err := p.client.Do(proxyReq)
	if err != nil {
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}

	// Handle gzip-encoded responses
	if resp.Header.Get("Content-Encoding") == "gzip" {
		resp.Header.Del("Content-Encoding")
		resp.Header.Del("Content-Length")
		gzipReader, err := gzip.NewReader(resp.Body)
		if err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to create gzip reader: %w", err)
		}
		resp.Body = &gzipResponseBody{
			Reader: gzipReader,
			closer: resp.Body,
		}
	}

	return resp, nil
}

type gzipResponseBody struct {
	io.Reader
	closer io.Closer
}

func (g *gzipResponseBody) Close() error {
	if gzReader, ok := g.Reader.(*gzip.Reader); ok {
		gzReader.Close()
	}
	return g.closer.Close()
}

func removeHopByHopHeaders(header http.Header) {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"TE",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, h := range hopByHopHeaders {
		header.Del(h)
	}

	// Remove any headers specified in the Connection header
	if connection := header.Get("Connection"); connection != "" {
		for _, h := range strings.Split(connection, ",") {
			header.Del(strings.TrimSpace(h))
		}
		header.Del("Connection")
	}
}
