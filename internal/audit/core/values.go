package core

import (
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
)

func SerializeValues(values []interface{}, separator string) string {
	parts := make([]string, len(values))
	for i, value := range values {
		parts[i] = FormatDatabaseValue(value)
	}
	return strings.Join(parts, separator)
}

func FormatDatabaseValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return "<nil>"
	case []byte:
		if len(v) == 16 {
			return formatUUIDBytes(v)
		}
		return string(v)
	default:
		if uuidValue, ok := formatUUIDArrayValue(v); ok {
			return uuidValue
		}
		return fmt.Sprintf("%v", v)
	}
}

func formatUUIDBytes(value []byte) string {
	encoded := hex.EncodeToString(value)
	if len(encoded) != 32 {
		return encoded
	}

	return fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		encoded[0:8],
		encoded[8:12],
		encoded[12:16],
		encoded[16:20],
		encoded[20:32],
	)
}

func formatUUIDArrayValue(value interface{}) (string, bool) {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() || rv.Kind() != reflect.Array || rv.Len() != 16 {
		return "", false
	}

	if rv.Type().Elem().Kind() != reflect.Uint8 {
		return "", false
	}

	buf := make([]byte, rv.Len())
	for i := 0; i < rv.Len(); i++ {
		buf[i] = byte(rv.Index(i).Uint())
	}

	return formatUUIDBytes(buf), true
}
