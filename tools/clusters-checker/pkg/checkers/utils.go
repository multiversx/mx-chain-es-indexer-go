package checkers

import (
	"encoding/json"
	"fmt"
	"reflect"
)

func areEqualJSON(s1, s2 json.RawMessage) (bool, error) {
	var o1 interface{}
	var o2 interface{}

	var err error
	err = json.Unmarshal(s1, &o1)
	if err != nil {
		return false, fmt.Errorf("error unmashalling s1: %s", err.Error())
	}
	err = json.Unmarshal(s2, &o2)
	if err != nil {
		return false, fmt.Errorf("error unmashalling s2: %s", err.Error())
	}

	return reflect.DeepEqual(o1, o2), nil
}
