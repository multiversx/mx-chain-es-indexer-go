package converters

import (
	"strings"
	"unicode/utf8"

	"github.com/multiversx/mx-chain-es-indexer-go/data"
)

// TruncateFieldIfExceedsMaxLength will truncate the provided field if the max length exceeds.
// The length is checked on the value that will be indexed: the JSON encoder replaces every invalid UTF-8 byte with the
// replacement character, which has 3 bytes, so a field with invalid bytes can be up to 3 times bigger after encoding
func TruncateFieldIfExceedsMaxLength(field string) string {
	if len(field) <= data.MaxFieldLength && utf8.ValidString(field) {
		return field
	}

	return truncateAsValidUTF8(field, data.MaxFieldLength)
}

// truncateAsValidUTF8 returns the field as a valid UTF-8 string of maximum maxLength bytes.
// Every invalid byte is replaced with the replacement character (the same thing that the JSON encoder does)
// and the field is cut only between characters
func truncateAsValidUTF8(field string, maxLength int) string {
	sb := strings.Builder{}
	for i := 0; i < len(field); {
		// an invalid byte is decoded as utf8.RuneError with the size 1
		char, size := utf8.DecodeRuneInString(field[i:])
		if sb.Len()+utf8.RuneLen(char) > maxLength {
			break
		}

		_, _ = sb.WriteRune(char)
		i += size
	}

	return sb.String()
}

// TruncateFieldIfExceedsMaxLengthBase64 will truncate the provided field if the max length after base64 encoding would be exceeded
// this function will be used for the fields that after will be base64 encoded
func TruncateFieldIfExceedsMaxLengthBase64(field string) string {
	if len(field) > data.MaxKeywordFieldLengthBeforeBase64Encoding {
		return field[:data.MaxKeywordFieldLengthBeforeBase64Encoding]
	}

	return field
}

// TruncateSliceElementsIfExceedsMaxLength will truncate the provided slice of the field if the max length is exceeded
func TruncateSliceElementsIfExceedsMaxLength(fields []string) []string {
	var localFields []string
	for _, field := range fields {
		localFields = append(localFields, TruncateFieldIfExceedsMaxLength(field))
	}

	return localFields
}
