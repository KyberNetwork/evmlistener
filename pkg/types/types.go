package types

import (
	"math/big"

	"github.com/KyberNetwork/evmlistener/protobuf/pb"
	"github.com/ethereum/go-ethereum/common"
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
	Number      *big.Int `json:"number"`
	Hash        string   `json:"hash"`
	Timestamp   uint64   `json:"timestamp"`
	ParentHash  string   `json:"parentHash"`
	ReorgedHash string   `json:"reorgedHash"`
	Logs        []Log    `json:"logs"`
}

func (b Block) ToProtobuf() *pb.Block {
	logs := make([]*pb.Log, len(b.Logs))
	for i, l := range b.Logs {
		logs[i] = l.ToProtobuf()
	}

	return &pb.Block{
		Number:      b.Number.Uint64(),
		Hash:        common.FromHex(b.Hash),
		Timestamp:   b.Timestamp,
		ParentHash:  common.FromHex(b.ParentHash),
		ReorgedHash: common.FromHex(b.ReorgedHash),
		Logs:        logs,
	}
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"revertedBlocks"`
	NewBlocks      []Block `json:"newBlocks"`
}

func (m Message) ToProtobuf() *pb.Message {
	revertedBlocks := make([]*pb.Block, len(m.RevertedBlocks))
	for i, b := range m.RevertedBlocks {
		revertedBlocks[i] = b.ToProtobuf()
	}
	newBlocks := make([]*pb.Block, len(m.NewBlocks))
	for i, b := range m.NewBlocks {
		newBlocks[i] = b.ToProtobuf()
	}

	return &pb.Message{
		RevertedBlocks: revertedBlocks,
		NewBlocks:      newBlocks,
	}
}
