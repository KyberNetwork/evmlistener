package encoder

import "encoding/json"

type JSONEncoder struct{}

func NewJSONEncoder() *JSONEncoder {
	return &JSONEncoder{}
}

func (e *JSONEncoder) Encode(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}

func (e *JSONEncoder) Decode(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}
