package types

import (
	"math/big"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
)

// Header contains block header information.
type Header struct {
	Hash       string   `json:"hash"`
	ParentHash string   `json:"parentHash"`
	Number     *big.Int `json:"number"`
	Time       uint64   `json:"timestamp"`
}

// Block contains information of block.
type Block struct {
	Number      *big.Int               `json:"number"`
	Hash        string                 `json:"hash"`
	Timestamp   uint64                 `json:"timestamp"`
	ParentHash  string                 `json:"parentHash"`
	ReorgedHash string                 `json:"reorgedHash"`
	Logs        []Log                  `json:"logs"`
	TxnTrace    []*pb.TransactionTrace `json:"-"`
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"revertedBlocks"`
	NewBlocks      []Block `json:"newBlocks"`
}
