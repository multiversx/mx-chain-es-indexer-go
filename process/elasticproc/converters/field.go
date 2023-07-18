package converters

import "github.com/multiversx/mx-chain-es-indexer-go/data"

// TruncateFieldIfExceedsMaxLength will truncate the provided field if the max length exceeds
func TruncateFieldIfExceedsMaxLength(field string) string {
	if len(field) > data.MaxFieldLength {
		return field[:data.MaxFieldLength]
	}

	return field
}

// TruncateFieldIfExceedsMaxLengthBase64 will truncate the provided field if the max length exceeds
// this function will be used for the fields that after will be base64 encoded
func TruncateFieldIfExceedsMaxLengthBase64(field string) string {
	if len(field) > data.MaxLengthForBase64EncodedFields {
		return field[:data.MaxLengthForBase64EncodedFields]
	}

	return field
}

//TruncateSliceElementsIfExceedsMaxLength will truncate the provided slice of the field if the max length is exceeded
func TruncateSliceElementsIfExceedsMaxLength(fields []string) []string {
	var localFields []string
	for _, field := range fields {
		localFields = append(localFields, TruncateFieldIfExceedsMaxLength(field))
	}

	return localFields
}
