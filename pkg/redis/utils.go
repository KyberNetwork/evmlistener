package redis

import (
	"encoding/json"
	"strings"
)

// Encode marshals a data type in go to a slice of bytes.
func Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}

// Decode unmarshals data from a slice of bytes to a data type in go.
func Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

// FormatKey returns a key from a list of strings.
func FormatKey(args ...string) string {
	var key strings.Builder
	for _, arg := range args {
		key.WriteString(arg)
	}

	return key.String()
}
