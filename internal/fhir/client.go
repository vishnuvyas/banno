package fhir

import (
	"context"
	"net/url"
)

// SearchParams are FHIR search parameters, e.g. {"name": {"Smith"}, "_count": {"50"}}.
type SearchParams = url.Values

// Client is the set of operations common to any FHIR server, regardless of
// vendor. Vendor-specific behavior - auth, proprietary endpoints, HIE
// aggregation features - is layered on top via the capability interfaces in
// capabilities.go rather than growing this interface per vendor.
type Client interface {
	// Metadata fetches the server's CapabilityStatement from [base]/metadata.
	Metadata(ctx context.Context) (Resource, error)

	// Read fetches a single resource by type and id.
	Read(ctx context.Context, resourceType, id string) (Resource, error)

	// Create posts a new resource of the given type, returning the
	// server-assigned resource (including its id).
	Create(ctx context.Context, resourceType string, resource Resource) (Resource, error)

	// Update replaces the resource at resourceType/id.
	Update(ctx context.Context, resourceType, id string, resource Resource) (Resource, error)

	// Delete removes the resource at resourceType/id.
	Delete(ctx context.Context, resourceType, id string) error

	// Search executes a search against a resource type, returning the result
	// Bundle.
	Search(ctx context.Context, resourceType string, params SearchParams) (*Bundle, error)

	// Batch submits a batch or transaction bundle - which one depends on
	// bundle.Type ("batch" or "transaction") per the FHIR spec, the request
	// shape is identical - and returns the response bundle.
	Batch(ctx context.Context, bundle *Bundle) (*Bundle, error)

	// Operation invokes a FHIR "$operation". If id is empty the operation is
	// invoked at the type level ([base]/resourceType/$name); otherwise at the
	// instance level ([base]/resourceType/id/$name). A nil body issues a GET,
	// a non-nil body issues a POST with that body. The result can be any
	// resource type (Bundle, Parameters, ...) - decode it into whatever the
	// operation is documented to return.
	Operation(ctx context.Context, resourceType, id, name string, params SearchParams, body Resource) (Resource, error)
}
