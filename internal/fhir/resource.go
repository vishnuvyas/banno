// Package fhir defines a vendor-neutral FHIR client interface plus the
// shared resource/bundle types needed to use it. Vendor-specific clients
// (e.g. internal/fhir/healthgorilla) implement Client and, where they offer
// proprietary HIE features, one or more of the capability interfaces in
// capabilities.go.
package fhir

import (
	"encoding/json"
	"fmt"
)

// Resource is a FHIR resource represented as raw JSON. Resources are kept as
// raw JSON rather than fully typed per-resource-type structs so Client stays
// generic across every resource type (and across R4/R4B/R5 differences);
// callers Decode into whatever shape they need.
type Resource json.RawMessage

func (r Resource) MarshalJSON() ([]byte, error) {
	return json.RawMessage(r).MarshalJSON()
}

func (r *Resource) UnmarshalJSON(data []byte) error {
	*r = append((*r)[0:0], data...)
	return nil
}

// ResourceType returns the resourceType field of the underlying JSON, or ""
// if it isn't set or the resource can't be parsed.
func (r Resource) ResourceType() string {
	var v struct {
		ResourceType string `json:"resourceType"`
	}
	if err := json.Unmarshal(r, &v); err != nil {
		return ""
	}
	return v.ResourceType
}

// ID returns the id field of the underlying JSON, or "" if it isn't set.
func (r Resource) ID() string {
	var v struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(r, &v); err != nil {
		return ""
	}
	return v.ID
}

// Decode unmarshals the resource into v.
func (r Resource) Decode(v any) error {
	if len(r) == 0 {
		return fmt.Errorf("fhir: empty resource")
	}
	return json.Unmarshal(r, v)
}

// NewResource marshals v into a Resource.
func NewResource(v any) (Resource, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("fhir: marshaling resource: %w", err)
	}
	return Resource(data), nil
}
