package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"reflect"

	"gorm.io/gorm/schema"
)

// Base64Serializer
type Base64Serializer struct {
}

// Scan implements serializer interface
func (Base64Serializer) Scan(ctx context.Context, field *schema.Field, dst reflect.Value, dbValue interface{}) (err error) {
	fieldValue := reflect.New(field.FieldType)

	if dbValue != nil {
		var str string
		switch v := dbValue.(type) {
		case []byte:
			str = string(v)
		case string:
			str = v
		default:
			return fmt.Errorf("failed to unmarshal JSONB value: %#v", dbValue)
		}

		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}

		fieldValue.Set(reflect.ValueOf(data))
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())
	return
}

// Value implements serializer interface
func (Base64Serializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	var data []byte
	if fieldValue != nil {
		switch v := fieldValue.(type) {
		case []byte:
			data = v
		case string:
			data = []byte(v)
		default:
			return "", fmt.Errorf("failed to unmarshal base64 value: %#v", fieldValue)
		}
	}

	data = bytes.Trim(data, "\x00")

	return base64.StdEncoding.EncodeToString(data), nil
}
