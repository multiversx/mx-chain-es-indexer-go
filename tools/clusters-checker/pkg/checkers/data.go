package checkers

import "encoding/json"

type generalElasticResponse struct {
	Hits struct {
		Hits []struct {
			ID     string          `json:"_id"`
			Source json.RawMessage `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}

type Interval struct {
	start int64
	stop  int64
}
