package fhir

import "context"

// HIE (health information exchange) providers such as Health Gorilla layer
// proprietary aggregation features on top of the FHIR API that have no
// standard FHIR shape and aren't offered by every vendor. Rather than
// growing Client with vendor-specific methods, each such feature gets its
// own small "capability" interface - the same pattern as
// http.Pusher/http.Flusher - so callers type-assert a Client to the
// capability they need:
//
//	if p360, ok := client.(fhir.Patient360Provider); ok {
//		result, err := p360.Patient360(ctx, patientID)
//	}
//
// This keeps Client stable while letting individual vendor clients opt into
// whatever HIE-specific operations they actually support. Add further
// capability interfaces here as more HIE-specific features are needed (e.g.
// patient discovery/record location, document query).

// Patient360Provider is implemented by clients that can trigger an
// aggregated clinical record retrieval for a patient from a connected health
// information exchange (patient discovery plus document aggregation across
// participating organizations/networks).
//
// The retrieval itself is typically asynchronous: implementations query
// connected networks, import matching documents into the patient's chart,
// and report how many were found/imported rather than returning the
// clinical data directly. Fetch the imported data afterwards with a normal
// Client.Search (e.g. for DocumentReference).
type Patient360Provider interface {
	Patient360(ctx context.Context, patientID string) (*Patient360Result, error)
}

// Patient360Result reports the outcome of a Patient360 retrieval.
type Patient360Result struct {
	OrganizationsTotal int `json:"organizationsTotal"`
	DocumentsFound     int `json:"documentsFound"`
	DocumentsImported  int `json:"documentsImported"`
	DocumentsSkipped   int `json:"documentsSkipped"`
	DocumentsFailed    int `json:"documentsFailed"`
}
