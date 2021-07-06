package converters

import "strings"

const (
	attributesSeparator = ";"
	keyValuesSeparator  = ":"
	valuesSeparator     = ","
	tagsKey             = "tags"
)

// ExtractTagsFromAttributes will extract tags from the attributes
func ExtractTagsFromAttributes(attributes []byte) []string {
	if len(attributes) == 0 {
		return nil
	}

	sAttributes := strings.Split(string(attributes), attributesSeparator)

	for _, keValuesPair := range sAttributes {
		sKeyValuesPair := strings.Split(keValuesPair, keyValuesSeparator)
		if len(sKeyValuesPair) < 2 {
			continue
		}
		if sKeyValuesPair[0] != tagsKey {
			continue
		}

		return strings.Split(sKeyValuesPair[1], valuesSeparator)
	}

	return nil
}
