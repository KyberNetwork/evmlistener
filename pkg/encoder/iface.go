package encoder

type Encoder interface {
	Encode(data any) ([]byte, error)
	Decode(data []byte, v any) error
}
