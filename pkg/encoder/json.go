package encoder

import "encoding/json"

type JSONEncoder struct{}

func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

func (e *JSONEncoder) Encode(data any) ([]byte, error) {
	return json.Marshal(data)
}

func (e *JSONEncoder) Decode(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
