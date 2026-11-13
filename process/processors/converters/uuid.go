package converters

import (
	"encoding/base64"
	"github.com/google/uuid"
)

// GenerateBase64UUID will generate a 24 bytes unique base string
func GenerateBase64UUID() string {
	uuidBytes := uuid.New()

	base64UUID := base64.URLEncoding.EncodeToString(uuidBytes[:])

	return base64UUID
}
