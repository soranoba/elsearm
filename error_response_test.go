package elsearm

import (
	"encoding/json"
	"testing"
)

func TestErrorResponse_Error(t *testing.T) {
	var _ error = &ErrorResponse{}
}

func TestErrorResponse(t *testing.T) {
	str := `{
		"error": {
			"root_cause": [
				{
					"type": "x_content_parse_exception",
					"reason": "[24:30] [date_histogram] failed to parse field [calendar_interval]"
				}
			],
			"type": "x_content_parse_exception",
			"reason": "[24:30] [date_histogram] failed to parse field [calendar_interval]",
			"caused_by": {
				"type": "illegal_argument_exception",
				"reason": "The supplied interval [10d] could not be parsed as a calendar interval."
			}
		},
		"status": 400
	}`
	var res ErrorResponse
	if err := json.Unmarshal([]byte(str), &res); err != nil {
		t.Error(err)
	}
}
