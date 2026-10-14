package dropbox

// signature request types
type (
	CustomField struct {
		Name     string `json:"name"`
		Type     string `json:"type"`
		Required *bool  `json:"required"`
		ApiID    string `json:"api_id"`
		Value    string `json:"string"`
	}

	SignatureRequest struct {
		SignatureRequestID    string        `json:"signature_request_id"`
		TestMode              bool          `json:"test_mode"`
		Title                 string        `json:"title"`
		OriginalTitle         string        `json:"original_title"`
		Subject               string        `json:"subject"`
		Message               string        `json:"message"`
		CreatedAt             uint64        `json:"created_at"`
		ExpiresAt             uint64        `json:"expires_at"`
		IsComplete            bool          `json:"is_complete"`
		IsDeclined            bool          `json:"is_declined"`
		HasError              bool          `json:"has_error"`
		CustomFields          []CustomField `json:"custom_fields"`
		RequestorEmailAddress string        `json:"requester_email_address"`
		FinalCopyURI          string        `json:"final_copy_uri"`
		FilesURL              string        `json:"files_url"`
	}

	ListInfo struct {
		Page       int `json:"page"`
		NumPages   int `json:"num_pages"`
		NumResults int `json:"num_results"`
		PageSize   int `json:"page_size"`
	}

	SignatureRequestListResponse struct {
		ListInfo          ListInfo           `json:"list_info"`
		SignatureRequests []SignatureRequest `json:"signature_requests"`
	}
)
