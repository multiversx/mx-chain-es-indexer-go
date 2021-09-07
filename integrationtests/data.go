package integrationtests

import "encoding/json"

// GenericResponse is the structure for the generic response
type GenericResponse struct {
	Docs []GenericResponseDB `json:"docs"`
}

// GenericResponseDB is the structure for the generic response database
type GenericResponseDB struct {
	Found  bool            `json:"found"`
	ID     string          `json:"_id"`
	Source json.RawMessage `json:"_source"`
}
