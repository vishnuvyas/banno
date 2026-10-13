package fhir

import "fmt"

// OperationOutcome is the FHIR resource servers use to describe errors (and
// sometimes warnings or informational results).
type OperationOutcome struct {
	ResourceType string                  `json:"resourceType"`
	Issue        []OperationOutcomeIssue `json:"issue"`
}

type OperationOutcomeIssue struct {
	Severity    string `json:"severity"`
	Code        string `json:"code"`
	Diagnostics string `json:"diagnostics,omitempty"`
}

// Error is returned by Client methods when a FHIR server responds with a
// non-success status. It carries the raw response so callers who need
// server-specific detail beyond the OperationOutcome can inspect it.
type Error struct {
	StatusCode int
	Outcome    *OperationOutcome
	Body       []byte
}

func (e *Error) Error() string {
	if e.Outcome != nil && len(e.Outcome.Issue) > 0 {
		issue := e.Outcome.Issue[0]
		if issue.Diagnostics != "" {
			return fmt.Sprintf("fhir: status %d: %s: %s", e.StatusCode, issue.Code, issue.Diagnostics)
		}
		return fmt.Sprintf("fhir: status %d: %s", e.StatusCode, issue.Code)
	}
	return fmt.Sprintf("fhir: status %d", e.StatusCode)
}
