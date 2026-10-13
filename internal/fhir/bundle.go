package fhir

// Bundle covers every use of the FHIR Bundle resource this package needs:
// search results, batch/transaction requests and responses, and the result
// of operations that return a Bundle (e.g. $everything, HIE aggregation
// operations).
type Bundle struct {
	ResourceType string        `json:"resourceType"`
	Type         string        `json:"type"`
	Total        *int          `json:"total,omitempty"`
	Link         []BundleLink  `json:"link,omitempty"`
	Entry        []BundleEntry `json:"entry,omitempty"`
}

type BundleLink struct {
	Relation string `json:"relation"`
	URL      string `json:"url"`
}

type BundleEntry struct {
	FullURL  string               `json:"fullUrl,omitempty"`
	Resource Resource             `json:"resource,omitempty"`
	Search   *BundleEntrySearch   `json:"search,omitempty"`
	Request  *BundleEntryRequest  `json:"request,omitempty"`
	Response *BundleEntryResponse `json:"response,omitempty"`
}

type BundleEntrySearch struct {
	Mode string `json:"mode,omitempty"`
}

type BundleEntryRequest struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type BundleEntryResponse struct {
	Status   string `json:"status"`
	Location string `json:"location,omitempty"`
}

// NewBundle constructs an empty bundle of the given type ("batch" or
// "transaction") ready to have entries appended.
func NewBundle(bundleType string) *Bundle {
	return &Bundle{ResourceType: "Bundle", Type: bundleType}
}
