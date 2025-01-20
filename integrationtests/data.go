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

func (gr *GenericResponse) UnmarshalJSON(data []byte) error {
	type Alias GenericResponse
	aux := &struct {
		Docs []json.RawMessage `json:"docs"`
		*Alias
	}{
		Alias: (*Alias)(gr),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	gr.Docs = make([]GenericResponseDB, len(aux.Docs))
	for i, doc := range aux.Docs {
		var docObj GenericResponseDB
		if err := json.Unmarshal(doc, &docObj); err != nil {
			return err
		}
		gr.Docs[i] = docObj
	}

	return nil
}
