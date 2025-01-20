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

// UnmarshalJSON will unmarshall and remove uuid field from json object
func (gr *GenericResponseDB) UnmarshalJSON(data []byte) error {
	type Alias GenericResponseDB
	aux := &struct {
		Source json.RawMessage `json:"_source"`
		*Alias
	}{
		Alias: (*Alias)(gr),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if gr.Source == nil {
		gr.Source = aux.Source
		return nil
	}

	var sourceMap map[string]interface{}
	if err := json.Unmarshal(aux.Source, &sourceMap); err != nil {
		return err
	}

	delete(sourceMap, "uuid")

	modifiedSource, err := json.Marshal(sourceMap)
	if err != nil {
		return err
	}

	gr.Source = modifiedSource
	return nil
}
