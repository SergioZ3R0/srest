// Package api provides a pure HTTP client for the Slurm REST API.
//
// This package must not contain any UI logic: it only builds and sends HTTP
// requests to slurmrestd.
//
// The client discovers the available data_parser version (or uses an
// explicitly pinned one) and uses it to build request paths, enabling
// version-gating of fields in future versions.
package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// Query is a single HTTP request/response recorded by the client. It is used
// by the UI's query inspector to show what srest sends to slurmrestd.
type Query struct {
	Method     string
	URL        string
	StatusCode int
	Duration   time.Duration
	Error      error
	Warnings   []Warning
	Errors     []Error
	Body       []byte
}

// Client is an HTTP client for interacting with slurmrestd.
type Client struct {
	baseURL    string
	jwt        string
	username   string
	httpClient *http.Client

	// version is the data_parser version we talk to. A zero Version means it
	// has not been determined yet.
	version Version

	// mu guards the query log, which is written from request goroutines.
	mu      sync.RWMutex
	queries []Query
}

// New creates a new API client with the given configuration.
func New(baseURL, jwt, username string) *Client {
	return &Client{
		baseURL:  baseURL,
		jwt:      jwt,
		username: username,
		httpClient: &http.Client{
			// Global timeout to avoid hanging requests.
			Timeout: 10 * time.Second,
		},
	}
}

// SetVersion pins the API version explicitly (e.g. from SLURM_API_VERSION),
// skipping auto-detection.
func (c *Client) SetVersion(v Version) {
	c.version = v
}

// Version returns the currently configured version and whether it is set.
func (c *Client) Version() (Version, bool) {
	return c.version, c.version != Version{}
}

// StatusError represents an HTTP response whose status code was not the
// expected one. It allows distinguishing, for example, a 404 during version
// auto-detection.
type StatusError struct {
	Code int
	Body string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected status %s: %s", http.StatusText(e.Code), e.Body)
}

// Detect tries the supported versions (highest first) and pins the first one
// that answers the /ping endpoint correctly. It returns the chosen version.
func (c *Client) Detect(ctx context.Context) (Version, error) {
	var lastErr error
	for _, v := range SupportedVersions {
		if _, err := c.ping(ctx, v); err != nil {
			// A 404 means that version does not exist on the cluster: try the
			// next one. Any other error (network, auth, etc.) is decisive and
			// is returned as-is.
			var se *StatusError
			if errors.As(err, &se) && se.Code == http.StatusNotFound {
				lastErr = err
				continue
			}
			return Version{}, err
		}
		c.version = v
		return v, nil
	}
	if lastErr != nil {
		return Version{}, fmt.Errorf("no supported API version found: %w", lastErr)
	}
	return Version{}, fmt.Errorf("no supported API version found")
}

// Partitions returns the names of all partitions visible to the user. Used by
// the query builder to offer partition values as selectable options.
func (c *Client) Partitions(ctx context.Context) ([]string, error) {
	v, ok := c.Version()
	if !ok {
		return nil, fmt.Errorf("API version not determined; call Detect or SetVersion first")
	}

	var resp struct {
		Partitions []struct {
			Name string `json:"name"`
		} `json:"partitions"`
	}
	if err := c.get(ctx, v, "/partitions", &resp); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(resp.Partitions))
	for _, p := range resp.Partitions {
		if p.Name != "" {
			names = append(names, p.Name)
		}
	}
	return names, nil
}

// Ping checks connectivity with slurmrestd and returns the relevant
// information (API and Slurm versions, and warnings).
func (c *Client) Ping(ctx context.Context) (PingInfo, error) {
	v, ok := c.Version()
	if !ok {
		return PingInfo{}, fmt.Errorf("API version not determined; call Detect or SetVersion first")
	}
	info, err := c.ping(ctx, v)
	if err != nil {
		return PingInfo{}, err
	}
	info.API = v
	return info, nil
}

// ping performs GET /slurm/{version}/ping and decodes the response.
func (c *Client) ping(ctx context.Context, version Version) (PingInfo, error) {
	var resp Response
	if err := c.get(ctx, version, "/ping", &resp); err != nil {
		return PingInfo{}, err
	}
	return PingInfo{
		Slurm:    resp.Meta.Slurm,
		Warnings: resp.Warnings,
	}, nil
}

// recordQuery appends a query to the internal log.
func (c *Client) recordQuery(q Query) {
	c.mu.Lock()
	c.queries = append(c.queries, q)
	c.mu.Unlock()
}

// Queries returns a snapshot of all recorded queries.
func (c *Client) Queries() []Query {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]Query, len(c.queries))
	copy(out, c.queries)
	return out
}

// Get issues a GET request to the given API path (e.g. "/jobs?state=RUNNING")
// using the detected API version. The response is recorded in the query log.
// If out is non-nil, the JSON body is decoded into it.
func (c *Client) Get(ctx context.Context, path string, out any) error {
	v, ok := c.Version()
	if !ok {
		return fmt.Errorf("API version not determined; call Detect or SetVersion first")
	}
	return c.get(ctx, v, path, out)
}

// get executes a GET request and decodes the resulting JSON into out.
func (c *Client) get(ctx context.Context, version Version, path string, out any) error {
	var endpoint string
	if strings.HasPrefix(path, "/slurmdb") {
		endpoint = c.baseURL + path
	} else {
		endpoint = fmt.Sprintf("%s/slurm/%s%s", c.baseURL, version, path)
	}
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("building request: %w", err)
	}

	// Authentication headers required by slurmrestd.
	req.Header.Set("X-SLURM-USER-TOKEN", c.jwt)
	req.Header.Set("X-SLURM-USER-NAME", c.username)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.recordQuery(Query{Method: http.MethodGet, URL: endpoint, Duration: time.Since(start), Error: err})
		return fmt.Errorf("contacting %s: %w", endpoint, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.recordQuery(Query{Method: http.MethodGet, URL: endpoint, StatusCode: resp.StatusCode, Duration: time.Since(start), Error: err})
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		se := &StatusError{Code: resp.StatusCode, Body: string(body)}
		c.recordQuery(Query{Method: http.MethodGet, URL: endpoint, StatusCode: resp.StatusCode, Duration: time.Since(start), Error: se, Body: body})
		return se
	}

	if out != nil {
		if err := json.Unmarshal(body, out); err != nil {
			c.recordQuery(Query{Method: http.MethodGet, URL: endpoint, StatusCode: resp.StatusCode, Duration: time.Since(start), Error: err})
			return fmt.Errorf("decoding response: %w", err)
		}
	}

	// Decode the common envelope to capture warnings/errors regardless of the
	// type requested in out.
	var envelope Response
	var warnings []Warning
	var qerrs []Error
	if err := json.Unmarshal(body, &envelope); err == nil {
		warnings = envelope.Warnings
		qerrs = envelope.Errors
	}

	c.recordQuery(Query{
		Method:     http.MethodGet,
		URL:        endpoint,
		StatusCode: resp.StatusCode,
		Duration:   time.Since(start),
		Warnings:   warnings,
		Errors:     qerrs,
		Body:       body,
	})
	return nil
}
