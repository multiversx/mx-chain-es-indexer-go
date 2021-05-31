package data

import "strings"

const (
	attributesSeparator = ";"
	keyValuesSeparator  = ":"
	valuesSeparator     = ","
)

// Attributes hold data about token attributes
type Attributes map[string][]string

// NewAttributesDTO will create a new instance of Attributes base of the input attributes
func NewAttributesDTO(attributes []byte) *Attributes {
	if len(attributes) == 0 {
		return nil
	}

	attrs := make(Attributes)
	sAttributes := strings.Split(string(attributes), attributesSeparator)

	for _, keValuesPair := range sAttributes {
		sKeyValuesPair := strings.Split(keValuesPair, keyValuesSeparator)
		if len(sKeyValuesPair) < 2 {
			continue
		}

		attrs[sKeyValuesPair[0]] = strings.Split(sKeyValuesPair[1], valuesSeparator)
	}

	if len(attrs) == 0 {
		return nil
	}

	return &attrs
}
