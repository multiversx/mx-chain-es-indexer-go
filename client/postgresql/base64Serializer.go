package postgres

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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
		case [][]byte:
			base64Data, err := base64.StdEncoding.DecodeString(str)
			if err != nil {
				return err
			}

			var data [][]byte
			err = json.Unmarshal(base64Data, &data)
			if err != nil {
				return err
			}

			fieldValue.Set(reflect.ValueOf(data))
			return nil
		case string:
			str = v
		default:
			return fmt.Errorf("failed to unmarshal base64 value: %#v", dbValue)
		}

		data, err := base64.StdEncoding.DecodeString(str)
		if err != nil {
			return err
		}

		fieldValue.Set(reflect.ValueOf(data))
	}

	field.ReflectValueOf(ctx, dst).Set(fieldValue.Elem())

	return nil
}

// Value implements serializer interface
func (Base64Serializer) Value(ctx context.Context, field *schema.Field, dst reflect.Value, fieldValue interface{}) (interface{}, error) {
	var data []byte
	var err error
	if fieldValue != nil {
		switch v := fieldValue.(type) {
		case []byte:
			data = v
		case [][]byte:
			data, err = json.Marshal(v)
			if err != nil {
				return "", err
			}
		case string:
			data = []byte(v)
		default:
			return "", fmt.Errorf("failed to marshal base64 value: %#v", fieldValue)
		}
	}

	data = bytes.Trim(data, "\x00")

	return base64.StdEncoding.EncodeToString(data), nil
}
