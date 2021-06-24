package tags

type responseTags struct {
	Docs []struct {
		Found  bool   `json:"found"`
		Tag    string `json:"_id"`
		Source struct {
			Count int `json:"count"`
		} `json:"_source"`
	} `json:"docs"`
}
