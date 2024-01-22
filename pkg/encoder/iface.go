package encoder

type Encoder interface {
	Encode(data interface{}) ([]byte, error)
	Decode(data []byte, v interface{}) error
}
