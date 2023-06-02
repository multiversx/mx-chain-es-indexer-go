package converters

import "github.com/multiversx/mx-chain-es-indexer-go/data"

// TruncateFieldIfExceedsMaxLength will truncate the provided field if the max length exceeds
func TruncateFieldIfExceedsMaxLength(field string) string {
	if len(field) > data.MaxFieldLength {
		return field[:data.MaxFieldLength]
	}

	return field
}

// TruncateSliceElementsIfExceedsMaxLength will truncate the provided slice of field if the max length exceeds
func TruncateSliceElementsIfExceedsMaxLength(fields []string) []string {
	var localFields []string
	for _, field := range fields {
		localFields = append(localFields, TruncateFieldIfExceedsMaxLength(field))
	}

	return localFields
}
