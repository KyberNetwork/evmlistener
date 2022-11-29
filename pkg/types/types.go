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
	ParentHash  common.Hash `json:"parent_hash"`
	ReorgedHash common.Hash `json:"reorged_hash"`
	Logs        []types.Log `json:"logs"`
}

// Message ...
type Message struct {
	RevertedBlocks []Block `json:"reverted_blocks"`
	NewBlocks      []Block `json:"new_blocks"`
}
