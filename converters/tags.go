package converters

import "strings"

const (
	attributesSeparator = ";"
	keyValuesSeparator  = ":"
	valuesSeparator     = ","
	tagsKey             = "tags"
	metadataKey         = "metadata"
)

// ExtractTagsFromAttributes will extract tags from the attributes
func ExtractTagsFromAttributes(attributes []byte) []string {
	return extractFromAttributes(attributes, tagsKey)
}

// ExtractMetaDataFromAttributes will extract metadata from attributes
func ExtractMetaDataFromAttributes(attributes []byte) string {
	res := extractFromAttributes(attributes, metadataKey)
	if len(res) < 1 {
		return ""
	}

	return res[0]
}

func extractFromAttributes(attributes []byte, key string) []string {
	if len(attributes) == 0 {
		return nil
	}

	sAttributes := strings.Split(string(attributes), attributesSeparator)

	for _, keValuesPair := range sAttributes {
		sKeyValuesPair := strings.Split(keValuesPair, keyValuesSeparator)
		if len(sKeyValuesPair) < 2 {
			continue
		}
		if sKeyValuesPair[0] != key {
			continue
		}

		return strings.Split(sKeyValuesPair[1], valuesSeparator)
	}

	return nil
}
