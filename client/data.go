package client

const (
	numOfErrorsToExtractBulkResponse = 5
)

// Item defines the structure of an item from a bulk response
type Item struct {
	Index  string `json:"_index"`
	ID     string `json:"_id"`
	Status int    `json:"status"`
	Result string `json:"result"`
	Error  struct {
		Type   string `json:"type"`
		Reason string `json:"reason"`
		Cause  struct {
			Type        string   `json:"type"`
			Reason      string   `json:"reason"`
			ScriptStack []string `json:"script_stack"`
			Script      string   `json:"script"`
		} `json:"caused_by"`
	} `json:"error"`
}
