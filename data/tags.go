package data

// ResponseTags is the structure for the tags response
type ResponseTags struct {
	Docs []ResponseTagDB `json:"docs"`
}

// ResponseTagDB is the structure for the tag response
type ResponseTagDB struct {
	Found   bool      `json:"found"`
	TagName string    `json:"_id"`
	Source  SourceTag `json:"_source"`
}

// SourceTag is the structure for the source body of a tag
type SourceTag struct {
	Count int `json:"count"`
}
