package redis

import (
	"encoding/json"
	"strings"
)

// Encode marshals a data type in go to a slice of bytes.
func Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

// Decode unmarshals data from a slice of bytes to a data type in go.
func Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

// FormatKey returns a key from a list of strings.
func FormatKey(sep string, args ...string) string {
	return strings.Join(args, sep)
}
