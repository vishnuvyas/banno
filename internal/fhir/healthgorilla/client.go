// Package healthgorilla implements fhir.Client against Health Gorilla's FHIR
// R4 API, plus Health Gorilla's HIE-specific Patient 360 aggregation feature
// (see patient360.go).
package healthgorilla

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"time"

	"golang.org/x/oauth2"

	"github.com/vishnuvyas/banno/internal/fhir"
)

const fhirContentType = "application/fhir+json"

// Documented FHIR R4 API bases, per
// https://developer.healthgorilla.com/reference/fhir-restful-api-design.
const (
	ProductionBaseURL = "https://api.healthgorilla.com/fhir/R4"
	SandboxBaseURL    = "https://sandbox.healthgorilla.com/fhir/R4"
)

// defaultAsyncPollInterval is the initial backoff between polls of an
// asynchronous operation's RequestResult URL (see patient360.go). It doubles
// on each attempt up to asyncPollMaxInterval, per Health Gorilla's guidance
// to poll with exponential backoff.
const (
	defaultAsyncPollInterval = 2 * time.Second
	asyncPollMaxInterval     = 30 * time.Second
)

// Config holds the settings needed to talk to Health Gorilla's FHIR R4 API.
type Config struct {
	// BaseURL is the FHIR API base - use ProductionBaseURL or SandboxBaseURL,
	// or an account-specific base if Health Gorilla issued one.
	BaseURL string
	// TokenURL is Health Gorilla's OAuth2 token endpoint, e.g.
	// "https://api.healthgorilla.com/oauth/token".
	TokenURL string
	// ClientID and ClientSecret are the OAuth2 client credentials issued by
	// Health Gorilla during onboarding. Health Gorilla's token endpoint does
	// not support the client_credentials grant; the client authenticates
	// using the JWT Bearer grant instead, signing a short-lived assertion
	// with ClientSecret (HS256). See auth.go and
	// https://developer.healthgorilla.com/reference/oauth-20-authentication.
	ClientID     string
	ClientSecret string
	// Scopes requested when obtaining an access token. If omitted, Health
	// Gorilla applies the scopes assigned to the client during registration.
	Scopes []string
	// HTTPClient is the base HTTP client used both for token requests and,
	// wrapped with OAuth2 token handling, for FHIR API requests. Defaults to
	// http.DefaultClient. Mainly useful for tests.
	HTTPClient *http.Client
}

// Client is a Health Gorilla FHIR R4 API client. It implements fhir.Client
// and fhir.Patient360Provider.
type Client struct {
	baseURL *url.URL
	http    *http.Client
	// pollInterval is the initial async poll backoff; overridden in tests to
	// avoid slow test runs.
	pollInterval time.Duration
}

var (
	_ fhir.Client             = (*Client)(nil)
	_ fhir.Patient360Provider = (*Client)(nil)
)

// NewClient builds a Client that authenticates with Health Gorilla using the
// JWT Bearer grant (see auth.go). The returned client manages its own token
// acquisition and refresh.
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	base, err := url.Parse(cfg.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: invalid base URL: %w", err)
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	tokenSource := oauth2.ReuseTokenSource(nil, &jwtBearerTokenSource{
		tokenURL:     cfg.TokenURL,
		clientID:     cfg.ClientID,
		clientSecret: cfg.ClientSecret,
		scopes:       cfg.Scopes,
		httpClient:   httpClient,
	})

	return &Client{
		baseURL:      base,
		http:         oauth2.NewClient(context.WithValue(ctx, oauth2.HTTPClient, httpClient), tokenSource),
		pollInterval: defaultAsyncPollInterval,
	}, nil
}

func (c *Client) resourceURL(resourceType, id string) *url.URL {
	u := *c.baseURL
	u.Path = path.Join(u.Path, resourceType, id)
	return &u
}

// newRequest builds a FHIR API request against u, with query merged in and
// body (if any) sent as application/fhir+json. It does not execute the
// request, so callers that need to set extra headers (e.g. patient360.go's
// Prefer: respond-async) can do so before calling c.http.Do.
func (c *Client) newRequest(ctx context.Context, method string, u *url.URL, query url.Values, body []byte) (*http.Request, error) {
	if len(query) > 0 {
		cloned := *u
		cloned.RawQuery = query.Encode()
		u = &cloned
	}

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), reader)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: building request: %w", err)
	}
	req.Header.Set("Accept", fhirContentType)
	if body != nil {
		req.Header.Set("Content-Type", fhirContentType)
	}
	return req, nil
}

func (c *Client) do(ctx context.Context, method string, u *url.URL, query url.Values, body []byte) ([]byte, int, error) {
	req, err := c.newRequest(ctx, method, u, query, body)
	if err != nil {
		return nil, 0, err
	}
	respBody, status, _, err := c.doRequest(req)
	return respBody, status, err
}

// doRequest executes req and also returns the response header, needed by
// patient360.go to read the Location/Content-Location header of a 202
// Accepted async response.
func (c *Client) doRequest(req *http.Request) ([]byte, int, http.Header, error) {
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, nil, fmt.Errorf("healthgorilla: request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, resp.Header, fmt.Errorf("healthgorilla: reading response: %w", err)
	}

	return respBody, resp.StatusCode, resp.Header, nil
}

func checkStatus(statusCode int, body []byte) error {
	if statusCode >= 200 && statusCode < 300 {
		return nil
	}
	fhirErr := &fhir.Error{StatusCode: statusCode, Body: body}
	var outcome fhir.OperationOutcome
	if json.Unmarshal(body, &outcome) == nil && outcome.ResourceType == "OperationOutcome" {
		fhirErr.Outcome = &outcome
	}
	return fhirErr
}

func (c *Client) Metadata(ctx context.Context) (fhir.Resource, error) {
	u := *c.baseURL
	u.Path = path.Join(u.Path, "metadata")
	body, status, err := c.do(ctx, http.MethodGet, &u, nil, nil)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	return fhir.Resource(body), nil
}

func (c *Client) Read(ctx context.Context, resourceType, id string) (fhir.Resource, error) {
	body, status, err := c.do(ctx, http.MethodGet, c.resourceURL(resourceType, id), nil, nil)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	return fhir.Resource(body), nil
}

func (c *Client) Create(ctx context.Context, resourceType string, resource fhir.Resource) (fhir.Resource, error) {
	body, status, err := c.do(ctx, http.MethodPost, c.resourceURL(resourceType, ""), nil, []byte(resource))
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	return fhir.Resource(body), nil
}

func (c *Client) Update(ctx context.Context, resourceType, id string, resource fhir.Resource) (fhir.Resource, error) {
	body, status, err := c.do(ctx, http.MethodPut, c.resourceURL(resourceType, id), nil, []byte(resource))
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	return fhir.Resource(body), nil
}

func (c *Client) Delete(ctx context.Context, resourceType, id string) error {
	body, status, err := c.do(ctx, http.MethodDelete, c.resourceURL(resourceType, id), nil, nil)
	if err != nil {
		return err
	}
	return checkStatus(status, body)
}

func (c *Client) Search(ctx context.Context, resourceType string, params fhir.SearchParams) (*fhir.Bundle, error) {
	body, status, err := c.do(ctx, http.MethodGet, c.resourceURL(resourceType, ""), params, nil)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	var bundle fhir.Bundle
	if err := json.Unmarshal(body, &bundle); err != nil {
		return nil, fmt.Errorf("healthgorilla: decoding search bundle: %w", err)
	}
	return &bundle, nil
}

func (c *Client) Batch(ctx context.Context, bundle *fhir.Bundle) (*fhir.Bundle, error) {
	payload, err := json.Marshal(bundle)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: marshaling bundle: %w", err)
	}
	body, status, err := c.do(ctx, http.MethodPost, c.baseURL, nil, payload)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, body); err != nil {
		return nil, err
	}
	var result fhir.Bundle
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("healthgorilla: decoding batch response: %w", err)
	}
	return &result, nil
}

func (c *Client) Operation(ctx context.Context, resourceType, id, name string, params fhir.SearchParams, body fhir.Resource) (fhir.Resource, error) {
	u := c.resourceURL(resourceType, id)
	u.Path = path.Join(u.Path, "$"+name)

	method := http.MethodGet
	var payload []byte
	if body != nil {
		method = http.MethodPost
		payload = []byte(body)
	}

	respBody, status, err := c.do(ctx, method, u, params, payload)
	if err != nil {
		return nil, err
	}
	if err := checkStatus(status, respBody); err != nil {
		return nil, err
	}
	return fhir.Resource(respBody), nil
}
