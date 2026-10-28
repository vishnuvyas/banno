// Package athenaone - API Client for accessing athenaone
package athenaone

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/vishnuvyas/banno/internal"
)

type AthenaOneClient interface {
	// Logs in and gets an access token
	Login(ctx context.Context) bool
}

type athenaOneClient struct {
	clientID     string
	clientSecret string
	baseURL      *url.URL
	practiceID   string
	accessToken  string
	client       *http.Client
}

func (a *athenaOneClient) do(ctx context.Context, method string, path string, body io.Reader, params map[string]string, headers map[string]string) (*http.Response, error) {
	requestUrl := a.baseURL.JoinPath(path)
	if len(params) > 0 {
		queryParams := requestUrl.Query()
		for k, v := range params {
			queryParams.Add(k, v)
		}
		requestUrl.RawQuery = queryParams.Encode()
	}

	request, err := http.NewRequestWithContext(ctx, method, requestUrl.RequestURI(), body)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %w", err)
	}

	if a.accessToken != "" {
		request.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.accessToken))
	}
	for k, v := range headers {
		request.Header.Add(k, v)
	}

	resp, err := a.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("unable to get response from client: %w", err)
	}
	return resp, nil
}

func (a *athenaOneClient) Login(ctx context.Context) bool {
	return false
}

func New(config internal.BannoConfig) (*athenaOneClient, error) {
	return nil, nil
}
