package encoder

import (
	"errors"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"google.golang.org/protobuf/proto"
)

type ProtobufEncoder struct{}

type ProtobufMessage interface {
	ToProtobuf() *pb.Message
}

func NewProtobufEncoder() *ProtobufEncoder {
	return &ProtobufEncoder{}
}

func (e *ProtobufEncoder) Encode(data interface{}) ([]byte, error) {
	cast, ok := data.(ProtobufMessage)
	if !ok {
		return nil, errors.New("mistype when encode with protobuf")
	}

	return proto.Marshal(cast.ToProtobuf())
}

func (e *ProtobufEncoder) Decode(data []byte, v interface{}) error {
	cast, ok := v.(proto.Message)
	if !ok {
		return errors.New("mistype when decode with protobuf")
	}

	return proto.Unmarshal(data, cast)
}
