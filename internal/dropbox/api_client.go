// Package dropbox - api client to get dropbox sign related stuff
// primarily focusing on documents and downloading them
package dropbox

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/vishnuvyas/banno/internal"
)

type DropboxAPIClient interface {
	GetSignatureRequests(ctx context.Context, accountID string, page int, pageSize int, query string) (SignatureRequestListResponse, error)
}

type apiClient struct {
	apiKey     string
	baseURL    *url.URL
	httpClient *http.Client
}

func (c *apiClient) do(ctx context.Context, method string, path string, queryParams map[string]string, body string) (*http.Response, error) {
	params := url.Values{}
	for key, val := range queryParams {
		params.Add(key, val)
	}
	requestURL := c.baseURL.Clone()
	requestURL = requestURL.JoinPath(path)
	if len(queryParams) > 0 {
		requestURL.RawQuery = params.Encode()
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL.String(), strings.NewReader(body))
	request.SetBasicAuth(c.apiKey, "")
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	return c.httpClient.Do(request)
}

func (c *apiClient) GetSignatureRequests(ctx context.Context, accountID string, page int, pageSize int, query string) (SignatureRequestListResponse, error) {
	if accountID == "" {
		accountID = "all"
	}

	resp, err := c.do(ctx, "GET", "/signature_request/list", map[string]string{
		"account_id": accountID,
		"page":       strconv.Itoa(page),
		"page_size":  strconv.Itoa(pageSize),
		"query":      query,
	}, "")
	if err != nil {
		return SignatureRequestListResponse{}, fmt.Errorf("unable to do the request: %w", err)
	}

	body := resp.Body
	defer body.Close()
	var signatureListResponse SignatureRequestListResponse
	decoder := json.NewDecoder(body)
	if err := decoder.Decode(&signatureListResponse); err != nil {
		return SignatureRequestListResponse{}, fmt.Errorf("unable to parse json response: %w", err)
	}

	return signatureListResponse, nil
}

func NewAPIClient(config internal.BannoConfig) (*apiClient, error) {
	baseUrl, _ := url.Parse("https://api.hellosign.com/v3")
	return &apiClient{
		apiKey:  config.GetDropboxAPIKey(),
		baseURL: baseUrl,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}
