package templatesConfig

import (
	"bytes"
	"encoding/json"
)

type Array []interface{}

type Object map[string]interface{}

func (o *Object) ToBuffer() *bytes.Buffer {
	objectBytes, _ := json.Marshal(o)

	buff := &bytes.Buffer{}
	_, _ = buff.Write(objectBytes)

	return buff
}
