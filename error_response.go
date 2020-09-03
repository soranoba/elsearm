package elsearm

// ErrorResponse is an error response format of Elasticsearch.
type ErrorResponse struct {
	Status uint `json:"status"`
	Error  struct {
		Type      string `json:"type"`
		Reason    string `json:"reason"`
		RootCause []struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"root_cause"`
		CausedBy struct {
			Type   string `json:"type"`
			Reason string `json:"reason"`
		} `json:"caused_by"`
	} `json:"error"`
}
