package types

import (
	"encoding/json"
	"errors"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

// Log contains log information.
type Log struct {
	Address     string   `json:"address"`
	Topics      []string `json:"topics"`
	Data        []byte   `json:"data"`
	BlockNumber uint64   `json:"blockNumber"`
	TxHash      string   `json:"transactionHash"`
	TxIndex     uint     `json:"transactionIndex"`
	BlockHash   string   `json:"blockHash"`
	Index       uint     `json:"logIndex"`
	Removed     bool     `json:"removed"`
}

// MarshalJSON marshals as JSON.
func (l *Log) MarshalJSON() ([]byte, error) {
	type Log struct {
		Address     string         `json:"address"`
		Topics      []string       `json:"topics"`
		Data        hexutil.Bytes  `json:"data"`
		BlockNumber hexutil.Uint64 `json:"blockNumber"`
		TxHash      string         `json:"transactionHash"`
		TxIndex     hexutil.Uint   `json:"transactionIndex"`
		BlockHash   string         `json:"blockHash"`
		Index       hexutil.Uint   `json:"logIndex"`
		Removed     bool           `json:"removed"`
	}

	var enc Log
	enc.Address = l.Address
	enc.Topics = l.Topics
	enc.Data = l.Data
	enc.BlockNumber = hexutil.Uint64(l.BlockNumber)
	enc.TxHash = l.TxHash
	enc.TxIndex = hexutil.Uint(l.TxIndex)
	enc.BlockHash = l.BlockHash
	enc.Index = hexutil.Uint(l.Index)
	enc.Removed = l.Removed

	return json.Marshal(&enc)
}

// UnmarshalJSON unmarshals from JSON.
//
//nolint:cyclop
func (l *Log) UnmarshalJSON(input []byte) error {
	type Log struct {
		Address     *string         `json:"address"`
		Topics      []string        `json:"topics"`
		Data        *hexutil.Bytes  `json:"data"`
		BlockNumber *hexutil.Uint64 `json:"blockNumber"`
		TxHash      *string         `json:"transactionHash"`
		TxIndex     *hexutil.Uint   `json:"transactionIndex"`
		BlockHash   *string         `json:"blockHash"`
		Index       *hexutil.Uint   `json:"logIndex"`
		Removed     *bool           `json:"removed"`
	}

	var dec Log
	if err := json.Unmarshal(input, &dec); err != nil {
		return err
	}
	if dec.Address == nil {
		return errors.New("missing required field 'address' for Log")
	}
	l.Address = *dec.Address
	if dec.Topics == nil {
		return errors.New("missing required field 'topics' for Log")
	}
	l.Topics = dec.Topics
	if dec.Data == nil {
		return errors.New("missing required field 'data' for Log")
	}
	l.Data = *dec.Data
	if dec.BlockNumber != nil {
		l.BlockNumber = uint64(*dec.BlockNumber)
	}
	if dec.TxHash == nil {
		return errors.New("missing required field 'transactionHash' for Log")
	}
	l.TxHash = *dec.TxHash
	if dec.TxIndex != nil {
		l.TxIndex = uint(*dec.TxIndex)
	}
	if dec.BlockHash != nil {
		l.BlockHash = *dec.BlockHash
	}
	if dec.Index != nil {
		l.Index = uint(*dec.Index)
	}
	if dec.Removed != nil {
		l.Removed = *dec.Removed
	}

	return nil
}

func (l *Log) ToProtobuf() *pb.Log {
	if l == nil {
		return nil
	}

	topics := make([][]byte, len(l.Topics))

	for i, t := range l.Topics {
		topics[i] = []byte(t)
	}

	return &pb.Log{
		Address:             []byte(l.Address),
		Topics:              topics,
		Data:                l.Data,
		BlockHash:           []byte(l.BlockHash),
		BlockNumber:         l.BlockNumber,
		TransactionIndex:    uint32(l.TxIndex),
		TransactionHash:     []byte(l.TxHash),
		TransactionLogIndex: uint32(l.Index),
		BlockIndex:          0, // TODO: I don't know how to get this, and is it necessary?
	}
}
