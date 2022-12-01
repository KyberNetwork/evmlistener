package types

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// Block contains information of block.
type Block struct {
	Number      *big.Int    `json:"number"`
	Hash        common.Hash `json:"hash"`
	ParentHash  common.Hash `json:"parentHash"`
	ReorgedHash common.Hash `json:"reorgedHash"`
	Logs        []types.Log `json:"logs"`
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"revertedBlocks"`
	NewBlocks      []Block `json:"newBlocks"`
}
