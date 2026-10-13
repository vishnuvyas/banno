package healthgorilla

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/vishnuvyas/banno/internal/fhir"
)

// patient360Operation triggers Health Gorilla's federated document
// retrieval across a patient's connected HIE networks (CommonWell,
// Carequality, eHealth Exchange, and QHINs). It is a type-level operation on
// DocumentReference, invoked with the patient's Health Gorilla ID as a
// query parameter, and always runs asynchronously - see
// https://developer.healthgorilla.com/docs/p360-query-api and
// https://developer.healthgorilla.com/reference/async-job-handling.
const patient360Operation = "p360-retrieve"

// expiredOrNotFoundStatus is Health Gorilla's non-standard status code for a
// RequestResult that has expired or no longer exists.
const expiredOrNotFoundStatus = 454

// Patient360 triggers Health Gorilla's Patient 360 retrieval for the given
// patient and blocks (polling with exponential backoff) until it completes,
// returning how many documents were found/imported. The retrieved documents
// themselves are then available via a normal Search, e.g.
// Search(ctx, "DocumentReference", url.Values{"patient": {patientID}}).
func (c *Client) Patient360(ctx context.Context, patientID string) (*fhir.Patient360Result, error) {
	if patientID == "" {
		return nil, fmt.Errorf("healthgorilla: patient360: patientID is required")
	}

	u := c.resourceURL("DocumentReference", "")
	u.Path = path.Join(u.Path, "$"+patient360Operation)

	req, err := c.newRequest(ctx, http.MethodGet, u, url.Values{"patient": {patientID}}, nil)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: patient360: %w", err)
	}
	// Mandatory per Health Gorilla's docs: without it the server may still
	// respond synchronously, but the documented contract requires this.
	req.Header.Set("Prefer", "respond-async")

	body, err := c.executeAsync(req)
	if err != nil {
		return nil, fmt.Errorf("healthgorilla: patient360: %w", err)
	}

	var result fhir.Patient360Result
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("healthgorilla: patient360: decoding result: %w", err)
	}
	return &result, nil
}

// executeAsync runs req and, if the server responds 202 Accepted, polls the
// Location/Content-Location URL with exponential backoff until the
// operation completes (200 OK) or fails.
func (c *Client) executeAsync(req *http.Request) ([]byte, error) {
	body, status, header, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}
	if status != http.StatusAccepted {
		if err := checkStatus(status, body); err != nil {
			return nil, err
		}
		return body, nil
	}

	location := header.Get("Location")
	if location == "" {
		location = header.Get("Content-Location")
	}
	if location == "" {
		return nil, fmt.Errorf("202 Accepted response missing Location/Content-Location header to poll")
	}

	return c.pollAsyncResult(req.Context(), location)
}

func (c *Client) pollAsyncResult(ctx context.Context, location string) ([]byte, error) {
	interval := c.pollInterval
	if interval <= 0 {
		interval = defaultAsyncPollInterval
	}

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, location, nil)
		if err != nil {
			return nil, fmt.Errorf("building poll request: %w", err)
		}
		req.Header.Set("Accept", fhirContentType)

		body, status, _, err := c.doRequest(req)
		if err != nil {
			return nil, err
		}

		switch status {
		case http.StatusOK:
			return body, nil
		case http.StatusAccepted:
			// still in progress; fall through to backoff and retry.
		case expiredOrNotFoundStatus:
			return nil, fmt.Errorf("async result at %s expired or was not found", location)
		default:
			return nil, checkStatus(status, body)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(interval):
		}
		interval *= 2
		if interval > asyncPollMaxInterval {
			interval = asyncPollMaxInterval
		}
	}
}
