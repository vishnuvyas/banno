package healthgorilla

import (
	"context"
	"fmt"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

func TestPatient360(t *testing.T) {
	var pollCount int32
	srv := newTestServer(t, map[string]http.HandlerFunc{
		"/fhir/R4/DocumentReference/$p360-retrieve": func(w http.ResponseWriter, r *http.Request) {
			requireBearerToken(t, r)
			if r.Method != http.MethodGet {
				t.Errorf("method = %s, want GET", r.Method)
			}
			if got := r.Header.Get("Prefer"); got != "respond-async" {
				t.Errorf("Prefer header = %q, want respond-async", got)
			}
			if got := r.URL.Query().Get("patient"); got != "123" {
				t.Errorf("patient param = %q, want 123", got)
			}
			w.Header().Set("Location", "http://"+r.Host+"/poll/abc123")
			w.WriteHeader(http.StatusAccepted)
		},
		"/poll/abc123": func(w http.ResponseWriter, r *http.Request) {
			if atomic.AddInt32(&pollCount, 1) == 1 {
				w.WriteHeader(http.StatusAccepted)
				return
			}
			fmt.Fprint(w, `{
				"organizationsTotal": 5,
				"documentsFound": 10,
				"documentsImported": 8,
				"documentsSkipped": 1,
				"documentsFailed": 1
			}`)
		},
	})
	c := newTestClient(t, srv)
	c.pollInterval = time.Millisecond

	result, err := c.Patient360(context.Background(), "123")
	if err != nil {
		t.Fatalf("Patient360: %v", err)
	}
	if result.OrganizationsTotal != 5 || result.DocumentsFound != 10 || result.DocumentsImported != 8 {
		t.Fatalf("unexpected result: %+v", result)
	}
	if got := atomic.LoadInt32(&pollCount); got < 2 {
		t.Fatalf("expected at least 2 polls, got %d", got)
	}
}

func TestPatient360RequiresPatientID(t *testing.T) {
	c := &Client{}
	if _, err := c.Patient360(context.Background(), ""); err == nil {
		t.Fatal("expected error for empty patientID")
	}
}
