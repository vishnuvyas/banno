package healthgorilla

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vishnuvyas/banno/internal/fhir"
)

const (
	testClientID     = "id"
	testClientSecret = "secret"
)

// newTestServer wires up a fake token endpoint that validates the JWT Bearer
// assertion our client sends (grant type, claims, and HMAC signature) the
// same way Health Gorilla's real token endpoint would, plus whatever
// resource handlers the test supplies.
func newTestServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parsing token request form: %v", err)
		}
		if got := r.PostForm.Get("grant_type"); got != "urn:ietf:params:oauth:grant-type:jwt-bearer" {
			t.Errorf("grant_type = %q, want JWT bearer", got)
		}
		if got := r.PostForm.Get("client_id"); got != testClientID {
			t.Errorf("client_id = %q, want %q", got, testClientID)
		}
		if err := verifyTestAssertion(r.PostForm.Get("assertion")); err != nil {
			t.Errorf("invalid assertion: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": "test-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		})
	})
	for pattern, h := range handlers {
		mux.HandleFunc(pattern, h)
	}
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// verifyTestAssertion re-derives the HMAC signature over the assertion's
// header+payload using testClientSecret and checks it matches, exercising
// the real signing path in auth.go end-to-end.
func verifyTestAssertion(assertion string) error {
	parts := strings.Split(assertion, ".")
	if len(parts) != 3 {
		return fmt.Errorf("expected 3 JWT segments, got %d", len(parts))
	}

	var claims jwtClaims
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return fmt.Errorf("decoding claims: %w", err)
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return fmt.Errorf("unmarshaling claims: %w", err)
	}
	if claims.Iss != testClientID || claims.Sub != testClientID {
		return fmt.Errorf("unexpected iss/sub: %+v", claims)
	}

	mac := hmac.New(sha256.New, []byte(testClientSecret))
	mac.Write([]byte(parts[0] + "." + parts[1]))
	wantSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if parts[2] != wantSig {
		return fmt.Errorf("signature mismatch")
	}
	return nil
}

func newTestClient(t *testing.T, srv *httptest.Server) *Client {
	t.Helper()
	c, err := NewClient(context.Background(), Config{
		BaseURL:      srv.URL + "/fhir/R4",
		TokenURL:     srv.URL + "/oauth/token",
		ClientID:     testClientID,
		ClientSecret: testClientSecret,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	return c
}

func requireBearerToken(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
		t.Errorf("Authorization header = %q, want %q", got, "Bearer test-token")
	}
}

func TestMetadata(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/metadata": func(w http.ResponseWriter, r *http.Request) {
			requireBearerToken(t, r)
			fmt.Fprint(w, `{"resourceType":"CapabilityStatement","status":"active"}`)
		},
	})
	c := newTestClient(t, srv)

	res, err := c.Metadata(context.Background())
	if err != nil {
		t.Fatalf("Metadata: %v", err)
	}
	if res.ResourceType() != "CapabilityStatement" {
		t.Fatalf("unexpected resource: %s", res)
	}
}

func TestRead(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/Patient/123": func(w http.ResponseWriter, r *http.Request) {
			requireBearerToken(t, r)
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			fmt.Fprint(w, `{"resourceType":"Patient","id":"123","name":[{"family":"Smith"}]}`)
		},
	})
	c := newTestClient(t, srv)

	res, err := c.Read(context.Background(), "Patient", "123")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if res.ResourceType() != "Patient" || res.ID() != "123" {
		t.Fatalf("unexpected resource: %s", res)
	}
}

func TestReadNotFound(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/Patient/missing": func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"resourceType":"OperationOutcome","issue":[{"severity":"error","code":"not-found","diagnostics":"no such patient"}]}`)
		},
	})
	c := newTestClient(t, srv)

	_, err := c.Read(context.Background(), "Patient", "missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var fhirErr *fhir.Error
	if !errors.As(err, &fhirErr) {
		t.Fatalf("expected *fhir.Error, got %T: %v", err, err)
	}
	if fhirErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want 404", fhirErr.StatusCode)
	}
	if fhirErr.Outcome == nil || fhirErr.Outcome.Issue[0].Diagnostics != "no such patient" {
		t.Errorf("unexpected outcome: %+v", fhirErr.Outcome)
	}
}

func TestCreate(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/Patient": func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("method = %s, want POST", r.Method)
			}
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"resourceType":"Patient","id":"new-id"}`)
		},
	})
	c := newTestClient(t, srv)

	res, err := c.Create(context.Background(), "Patient", fhir.Resource(`{"resourceType":"Patient"}`))
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if res.ID() != "new-id" {
		t.Fatalf("unexpected resource: %s", res)
	}
}

func TestSearch(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/Patient": func(w http.ResponseWriter, r *http.Request) {
			if got := r.URL.Query().Get("family"); got != "Smith" {
				t.Errorf("family param = %q, want Smith", got)
			}
			fmt.Fprint(w, `{"resourceType":"Bundle","type":"searchset","entry":[{"resource":{"resourceType":"Patient","id":"1"}}]}`)
		},
	})
	c := newTestClient(t, srv)

	bundle, err := c.Search(context.Background(), "Patient", fhir.SearchParams{"family": []string{"Smith"}})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(bundle.Entry) != 1 || bundle.Entry[0].Resource.ID() != "1" {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}
}

func TestBatch(t *testing.T) {
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4": func(w http.ResponseWriter, r *http.Request) {
			var req fhir.Bundle
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("decoding request body: %v", err)
			}
			if req.Type != "transaction" {
				t.Errorf("bundle type = %q, want transaction", req.Type)
			}
			fmt.Fprint(w, `{"resourceType":"Bundle","type":"transaction-response","entry":[{"response":{"status":"201 Created"}}]}`)
		},
	})
	c := newTestClient(t, srv)

	req := fhir.NewBundle("transaction")
	req.Entry = append(req.Entry, fhir.BundleEntry{
		Resource: fhir.Resource(`{"resourceType":"Patient"}`),
		Request:  &fhir.BundleEntryRequest{Method: "POST", URL: "Patient"},
	})

	resp, err := c.Batch(context.Background(), req)
	if err != nil {
		t.Fatalf("Batch: %v", err)
	}
	if resp.Type != "transaction-response" || len(resp.Entry) != 1 {
		t.Fatalf("unexpected response: %+v", resp)
	}
}
