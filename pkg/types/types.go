package types

import (
	"math/big"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
)

// Header contains block header information.
type Header struct {
	Hash             string   `json:"hash"`
	ParentHash       string   `json:"parentHash"`
	Number           *big.Int `json:"number"`
	Time             uint64   `json:"timestamp"`
	UncleHash        string   `json:"-"`
	Coinbase         string   `json:"-"`
	StateRoot        string   `json:"-"`
	TransactionsRoot string   `json:"-"`
	ReceiptRoot      string   `json:"-"`
	LogsBloom        []byte   `json:"-"`
	Difficulty       *big.Int `json:"-"`
	GasLimit         uint64   `json:"-"`
	GasUsed          uint64   `json:"-"`
	Timestamp        uint64   `json:"-"`
	ExtraData        []byte   `json:"-"`
	MixHash          string   `json:"-"`
	Nonce            uint64   `json:"-"`
	BaseFeePerGas    *big.Int `json:"-"`
}

func (h *Header) ToProtobuf(blockHash string) *pb.BlockHeader {
	if h == nil {
		return nil
	}

	difficulty := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if h.Difficulty != nil {
		difficulty = &pb.BigInt{Bytes: h.Difficulty.Bytes()}
	}
	baseFeePerGas := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if h.BaseFeePerGas != nil {
		baseFeePerGas = &pb.BigInt{Bytes: h.BaseFeePerGas.Bytes()}
	}

	return &pb.BlockHeader{
		ParentHash:       []byte(h.ParentHash),
		UncleHash:        []byte(h.UncleHash),
		Coinbase:         []byte(h.Coinbase),
		StateRoot:        []byte(h.StateRoot),
		TransactionsRoot: []byte(h.TransactionsRoot),
		ReceiptRoot:      []byte(h.ReceiptRoot),
		LogsBloom:        h.LogsBloom,
		Difficulty:       difficulty,
		Number:           h.Number.Uint64(),
		GasLimit:         h.GasLimit,
		GasUsed:          h.GasUsed,
		Timestamp:        h.Timestamp,
		ExtraData:        h.ExtraData,
		MixHash:          []byte(h.MixHash),
		Nonce:            h.Nonce,
		Hash:             []byte(blockHash),
		BaseFeePerGas:    baseFeePerGas,
		TotalDifficulty:  nil, // TODO: not found in the RPC call
	}
}

type Txn struct {
	To                   string         `json:"-"`
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
	Hash                 string         `json:"-"`
	From                 string         `json:"-"`
	TransactionIndex     uint64         `json:"-"`
}

type AccessTuple struct {
	Address     string   `json:"-"`
	StorageKeys [][]byte `json:"-"`
}

func (tx *Txn) ToProtobuf() *pb.TransactionTrace {
	if tx == nil {
		return nil
	}

	accessList := make([]*pb.AccessTuple, len(tx.AccessList))
	for i, a := range tx.AccessList {
		accessList[i].Address = []byte(a.Address)
		accessList[i].StorageKeys = a.StorageKeys
	}

	gasPrice := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if tx.GasPrice != nil {
		gasPrice = &pb.BigInt{Bytes: tx.GasPrice.Bytes()}
	}
	value := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if tx.Value != nil {
		value = &pb.BigInt{Bytes: tx.Value.Bytes()}
	}
	maxFeePerGas := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if tx.MaxFeePerGas != nil {
		maxFeePerGas = &pb.BigInt{Bytes: tx.MaxFeePerGas.Bytes()}
	}
	maxPriorityFeePerGas := &pb.BigInt{Bytes: big.NewInt(0).Bytes()}
	if tx.MaxPriorityFeePerGas != nil {
		maxPriorityFeePerGas = &pb.BigInt{Bytes: tx.MaxPriorityFeePerGas.Bytes()}
	}

	return &pb.TransactionTrace{
		To:                   []byte(tx.To),
		Nonce:                tx.Nonce,
		GasPrice:             gasPrice,
		GasLimit:             tx.GasLimit,
		Value:                value,
		Input:                tx.Input,
		V:                    tx.V,
		R:                    tx.R,
		S:                    tx.S,
		Type:                 pb.TransactionTrace_Type(tx.Type),
		AccessList:           accessList,
		MaxFeePerGas:         maxFeePerGas,
		MaxPriorityFeePerGas: maxPriorityFeePerGas,
		Hash:                 []byte(tx.Hash),
		From:                 []byte(tx.From),
		TransactionIndex:     &tx.TransactionIndex,
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
	Size         uint64   `json:"_"`
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
		Number:            b.Header.Number.Uint64(),
		Header:            header,
		TransactionTraces: txns,
		Logs:              logs,
		Size:              b.Size,
		Uncles:            nil, // TODO: I don't think we need this field
		BalanceChanges:    nil, // TODO: I don't think we need this field
		TraceCalls:        nil, // TODO: I don't think we need this field
		CodeChanges:       nil, // TODO: I don't think we need this field
		Ver:               0,   // TODO: I don't think we need this field
	}
}
