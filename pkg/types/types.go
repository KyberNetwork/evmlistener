package types

import (
	"math/big"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"github.com/ethereum/go-ethereum/common"
)

// Header contains block header information.
type Header struct {
	Hash             string         `json:"hash"`
	ParentHash       string         `json:"parentHash"`
	Number           *big.Int       `json:"number"`
	Time             uint64         `json:"timestamp"`
	UncleHash        common.Hash    `json:"-"`
	Coinbase         common.Address `json:"-"`
	StateRoot        common.Hash    `json:"-"`
	TransactionsRoot common.Hash    `json:"-"`
	ReceiptRoot      common.Hash    `json:"-"`
	LogsBloom        []byte         `json:"-"`
	Difficulty       *big.Int       `json:"-"`
	GasLimit         uint64         `json:"-"`
	GasUsed          uint64         `json:"-"`
	Timestamp        uint64         `json:"-"`
	ExtraData        []byte         `json:"-"`
	MixHash          common.Hash    `json:"-"`
	Nonce            uint64         `json:"-"`
	BaseFeePerGas    *big.Int       `json:"-"`
}

func (h *Header) ToProtobuf(blockHash string) *pb.BlockHeader {
	if h == nil {
		return nil
	}

	return &pb.BlockHeader{
		ParentHash:       []byte(h.ParentHash),
		UncleHash:        h.UncleHash.Bytes(),
		Coinbase:         h.Coinbase.Bytes(),
		StateRoot:        h.StateRoot.Bytes(),
		TransactionsRoot: h.TransactionsRoot.Bytes(),
		ReceiptRoot:      h.ReceiptRoot.Bytes(),
		LogsBloom:        h.LogsBloom,
		Difficulty:       &pb.BigInt{Bytes: h.Difficulty.Bytes()},
		Number:           h.Number.Uint64(),
		GasLimit:         h.GasLimit,
		GasUsed:          h.GasUsed,
		Timestamp:        h.Timestamp,
		ExtraData:        h.ExtraData,
		MixHash:          h.MixHash.Bytes(),
		Nonce:            h.Nonce,
		Hash:             []byte(blockHash),
		BaseFeePerGas:    &pb.BigInt{Bytes: h.BaseFeePerGas.Bytes()},
		TotalDifficulty:  nil, // TODO: not found in the RPC call
	}
}

type Txn struct {
	To                   common.Address `json:"-"`
	Nonce                uint64         `json:"-"`
	GasPrice             *big.Int       `json:"-"`
	GasLimit             uint64         `json:"-"`
	Value                *big.Int       `json:"-"`
	Input                []byte         `json:"-"`
	V                    []byte         `json:"-"`
	R                    []byte         `json:"-"`
	S                    []byte         `json:"-"`
	Type                 uint8          `json:"-"`
	AccessList           []*AccessTuple `json:"-"`
	MaxFeePerGas         *big.Int       `json:"-"`
	MaxPriorityFeePerGas *big.Int       `json:"-"`
	Hash                 common.Hash    `json:"-"`
	From                 common.Address `json:"-"`
}

type AccessTuple struct {
	Address     common.Address `json:"-"`
	StorageKeys [][]byte       `json:"-"`
}

func (tx *Txn) ToProtobuf() *pb.TransactionTrace {
	if tx == nil {
		return nil
	}

	accessList := make([]*pb.AccessTuple, len(tx.AccessList))
	for i, a := range tx.AccessList {
		accessList[i].Address = a.Address.Bytes()
		accessList[i].StorageKeys = a.StorageKeys
	}

	return &pb.TransactionTrace{
		To:                   tx.To.Bytes(),
		Nonce:                tx.Nonce,
		GasPrice:             &pb.BigInt{Bytes: tx.GasPrice.Bytes()},
		GasLimit:             tx.GasLimit,
		Value:                &pb.BigInt{Bytes: tx.Value.Bytes()},
		Input:                tx.Input,
		V:                    tx.V,
		R:                    tx.R,
		S:                    tx.S,
		Type:                 pb.TransactionTrace_Type(tx.Type),
		AccessList:           accessList,
		MaxFeePerGas:         &pb.BigInt{Bytes: tx.MaxFeePerGas.Bytes()},
		MaxPriorityFeePerGas: &pb.BigInt{Bytes: tx.MaxPriorityFeePerGas.Bytes()},
		Hash:                 tx.Hash.Bytes(),
		From:                 tx.From.Bytes(),
		TransactionIndex:     nil, // TODO: not found in the RPC call
		GasUsed:              0,   // TODO: not found in the RPC call
		Receipt:              nil, // TODO: not found in the RPC call
	}
}

// Block contains information of block.
type Block struct {
	Number       *big.Int `json:"number"`
	Hash         string   `json:"hash"`
	Timestamp    uint64   `json:"timestamp"`
	ParentHash   string   `json:"parentHash"`
	ReorgedHash  string   `json:"reorgedHash"`
	Logs         []Log    `json:"logs"`
	Transactions []Txn    `json:"-"`
	Header       Header   `json:"-"`
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"revertedBlocks"`
	NewBlocks      []Block `json:"newBlocks"`
}

func (b *Block) ToProtobuf() *pb.Block {
	if b == nil {
		return nil
	}

	logs := make([]*pb.Log, len(b.Logs))
	for i, l := range b.Logs {
		logs[i] = l.ToProtobuf()
	}

	header := b.Header.ToProtobuf(b.Hash)

	txns := make([]*pb.TransactionTrace, len(b.Transactions))
	for i, tx := range b.Transactions {
		txns[i] = tx.ToProtobuf()
	}

	return &pb.Block{
		Hash:              []byte(b.Hash),
		Number:            b.Number.Uint64(),
		Header:            header,
		TransactionTraces: txns,
		Logs:              logs,
		Uncles:            nil, // TODO: I don't think we need this field
		BalanceChanges:    nil, // TODO: I don't think we need this field
		TraceCalls:        nil, // TODO: I don't think we need this field
		CodeChanges:       nil, // TODO: I don't think we need this field
		Ver:               0,   // TODO: I don't think we need this field
		Size:              0,   // TODO: I don't think we need this field
	}
}
